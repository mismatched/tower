// Package main — MCP stdio server for tower.
//
// Implements the Model Context Protocol over stdin/stdout using only the
// standard library.  No SDK dependency.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/mismatched/libtower"
	"github.com/mismatched/tower/config"
	"github.com/mismatched/tower/util"
)

// --- JSON-RPC / MCP types ---

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Result  any    `json:"result,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type initializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    capabilities `json:"capabilities"`
	ServerInfo      serverInfo   `json:"serverInfo"`
}

type capabilities struct {
	Tools *struct{} `json:"tools,omitempty"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type toolDef struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	InputSchema any    `json:"inputSchema"`
}

type callToolParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments,omitempty"`
}

type callToolResult struct {
	Content []contentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type contentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// --- tool definitions ---

var tools = []toolDef{
	{
		Name:        "tower_ping",
		Description: "Send ICMP echo requests to a host to measure latency and check reachability. Requires root/CAP_NET_RAW.",
		InputSchema: objSchema{"type": "object", "properties": objSchema{
			"host":  objSchema{"type": "string", "description": "hostname or IP to ping"},
			"count": objSchema{"type": "integer", "description": "number of pings", "default": 1},
		}, "required": []string{"host"}},
	},
	{
		Name:        "tower_dns",
		Description: "Resolve a hostname to an IP address via DNS, optionally querying a specific DNS server.",
		InputSchema: objSchema{"type": "object", "properties": objSchema{
			"addr":    objSchema{"type": "string", "description": "address to resolve"},
			"from":    objSchema{"type": "string", "description": "optional DNS server (ip:port)"},
			"timeout": objSchema{"type": "string", "description": "timeout duration", "default": "5s"},
		}, "required": []string{"addr"}},
	},
	{
		Name:        "tower_tcp",
		Description: "Check if a TCP port is open and accepting connections on a given host.",
		InputSchema: objSchema{"type": "object", "properties": objSchema{
			"host":    objSchema{"type": "string", "description": "hostname or IP"},
			"port":    objSchema{"type": "integer", "description": "port number"},
			"timeout": objSchema{"type": "string", "description": "timeout duration", "default": "5s"},
		}, "required": []string{"host", "port"}},
	},
	{
		Name:        "tower_tls",
		Description: "Check a TLS-secured port, optionally using a client certificate for mutual TLS authentication.",
		InputSchema: objSchema{"type": "object", "properties": objSchema{
			"host":    objSchema{"type": "string", "description": "hostname or IP"},
			"port":    objSchema{"type": "integer", "description": "port number"},
			"cert":    objSchema{"type": "string", "description": "client certificate file path"},
			"key":     objSchema{"type": "string", "description": "client private key file path"},
			"timeout": objSchema{"type": "string", "description": "timeout duration", "default": "5s"},
		}, "required": []string{"host", "port"}},
	},
	{
		Name:        "tower_http",
		Description: "Check the HTTP status code and response metadata for a URL.",
		InputSchema: objSchema{"type": "object", "properties": objSchema{
			"url":     objSchema{"type": "string", "description": "URL to check"},
			"method":  objSchema{"type": "string", "description": "HTTP method", "default": "GET"},
			"timeout": objSchema{"type": "string", "description": "timeout duration", "default": "10s"},
		}, "required": []string{"url"}},
	},
	{
		Name:        "tower_trace",
		Description: "Trace an HTTP request with per-phase timing: DNS, TLS handshake, connect, time-to-first-byte, and total.",
		InputSchema: objSchema{"type": "object", "properties": objSchema{
			"url":    objSchema{"type": "string", "description": "URL to trace"},
			"method": objSchema{"type": "string", "description": "HTTP method", "default": "GET"},
		}, "required": []string{"url"}},
	},
	{
		Name:        "tower_https",
		Description: "Check a TLS certificate's validity, with optional expiry warning (e.g. warn if expiring within 30 days).",
		InputSchema: objSchema{"type": "object", "properties": objSchema{
			"host":                 objSchema{"type": "string", "description": "hostname to check TLS certificate"},
			"port":                 objSchema{"type": "integer", "description": "port number", "default": 443},
			"timeout":              objSchema{"type": "string", "description": "timeout duration", "default": "5s"},
			"warn_if_expiring":     objSchema{"type": "string", "description": "warn if cert expires within duration, e.g. 720h for 30 days"},
			"insecure_skip_verify": objSchema{"type": "boolean", "description": "skip TLS certificate verification"},
		}, "required": []string{"host"}},
	},
	{
		Name:        "tower_ws",
		Description: "Check WebSocket connectivity by performing a WebSocket upgrade handshake.",
		InputSchema: objSchema{"type": "object", "properties": objSchema{
			"url":     objSchema{"type": "string", "description": "WebSocket URL (ws:// or wss://)"},
			"timeout": objSchema{"type": "string", "description": "timeout duration", "default": "5s"},
		}, "required": []string{"url"}},
	},
	{
		Name:        "tower_check",
		Description: "Run multiple network health checks from an inline YAML config. Supports tcp, https, dns, ping, http, and ws check types.",
		InputSchema: objSchema{"type": "object", "properties": objSchema{
			"yaml": objSchema{"type": "string", "description": "inline tower config YAML"},
		}, "required": []string{"yaml"}},
	},
}

// objSchema is a map used to build JSON Schema objects inline.
type objSchema map[string]any

// --- server entry point ---

// runServe starts the MCP stdio server.  It reads JSON-RPC requests from stdin
// and writes responses to stdout.  Logs go to stderr.
func runServe() error {
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	s := bufio.NewScanner(os.Stdin)
	s.Buffer(make([]byte, 0, 1024*1024), 1024*1024) // 1 MiB max message

	for s.Scan() {
		line := s.Bytes()
		if len(line) == 0 {
			continue
		}

		var req rpcRequest
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("bad request: %v", err)
			continue
		}
		if req.JSONRPC != "2.0" {
			continue
		}

		switch req.Method {
		case "initialize":
			res := rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: initializeResult{
					ProtocolVersion: "2024-11-05",
					Capabilities:    capabilities{Tools: &struct{}{}},
					ServerInfo:      serverInfo{Name: "tower", Version: "0.2.0"},
				},
			}
			writeJSON(os.Stdout, res)

		case "tools/list":
			res := rpcResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result:  map[string]any{"tools": tools},
			}
			writeJSON(os.Stdout, res)

		case "tools/call":
			var params callToolParams
			if err := json.Unmarshal(req.Params, &params); err != nil {
				writeJSON(os.Stdout, rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: map[string]any{"code": -32602, "message": err.Error()}})
				continue
			}
			result := callTool(context.Background(), params)
			writeJSON(os.Stdout, rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: result})

		case "notifications/initialized":
			// no response needed

		default:
			log.Printf("unknown method: %s", req.Method)
		}
	}
	return s.Err()
}

func writeJSON(w io.Writer, v any) {
	b, _ := json.Marshal(v)
	w.Write(b)
	w.Write([]byte{'\n'})
}

// --- tool dispatch ---

func callTool(ctx context.Context, p callToolParams) callToolResult {
	switch p.Name {
	case "tower_ping":
		return handlePing(ctx, p.Arguments)
	case "tower_dns":
		return handleDNS(ctx, p.Arguments)
	case "tower_tcp":
		return handleTCP(ctx, p.Arguments)
	case "tower_tls":
		return handleTLS(ctx, p.Arguments)
	case "tower_http":
		return handleHTTP(ctx, p.Arguments)
	case "tower_trace":
		return handleTrace(ctx, p.Arguments)
	case "tower_https":
		return handleHTTPS(ctx, p.Arguments)
	case "tower_ws":
		return handleWS(ctx, p.Arguments)
	case "tower_check":
		return handleCheck(ctx, p.Arguments)
	default:
		return callToolResult{IsError: true, Content: []contentItem{{Type: "text", Text: fmt.Sprintf("unknown tool: %s", p.Name)}}}
	}
}

func strArg(args map[string]any, key, def string) string {
	if v, ok := args[key]; ok {
		return fmt.Sprint(v)
	}
	return def
}

func intArg(args map[string]any, key string, def int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			return int(n)
		case int:
			return n
		}
	}
	return def
}

func boolArg(args map[string]any, key string) bool {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func ok(v any) callToolResult {
	return callToolResult{Content: []contentItem{{Type: "text", Text: toJSON(v)}}}
}

func fail(msg string) callToolResult {
	return callToolResult{IsError: true, Content: []contentItem{{Type: "text", Text: msg}}}
}

func toJSON(v any) string { b, _ := json.Marshal(v); return string(b) }

// --- tool handlers ---

func handlePing(ctx context.Context, args map[string]any) callToolResult {
	host := strArg(args, "host", "")
	if host == "" {
		return fail("host is required")
	}
	count := intArg(args, "count", 1)
	if count < 1 {
		count = 1
	}
	ip, dur, err := libtower.PingContext(ctx, host, count)
	if err != nil {
		return fail(fmt.Sprintf("ping failed for %s: %v — ping requires root (sudo) or CAP_NET_RAW", host, err))
	}
	return ok(libtower.Result{OK: true, Duration: dur, Data: libtower.PingData{IP: ip}})
}

func handleDNS(ctx context.Context, args map[string]any) callToolResult {
	addr := strArg(args, "addr", "")
	if addr == "" {
		return fail("addr is required")
	}
	timeout, _ := time.ParseDuration(strArg(args, "timeout", "5s"))
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var r libtower.Result
	from := strArg(args, "from", "")
	if from != "" {
		ip, dur, err := libtower.DNSLookupFromContext(ctx, addr, from)
		if err != nil {
			return fail(fmt.Sprintf("DNS lookup from %s failed for %s: %v — check that the DNS server is reachable", from, addr, err))
		}
		r = libtower.Result{OK: true, Duration: dur, Data: libtower.DNSData{IP: ip}}
	} else {
		ip, dur, err := libtower.DNSLookupContext(ctx, addr)
		if err != nil {
			return fail(fmt.Sprintf("DNS lookup failed for %s: %v — check that the address is valid", addr, err))
		}
		r = libtower.Result{OK: true, Duration: dur, Data: libtower.DNSData{IP: ip}}
	}
	return ok(r)
}

func handleTCP(ctx context.Context, args map[string]any) callToolResult {
	host := strArg(args, "host", "")
	port := intArg(args, "port", 0)
	if host == "" || port == 0 {
		return fail("host and port are required")
	}
	timeout, _ := time.ParseDuration(strArg(args, "timeout", "5s"))
	u, err := parseTCPURL(host, port)
	if err != nil {
		return fail(fmt.Sprintf("invalid host:port %s:%d — %v", host, port, err))
	}
	r := (&libtower.TCP{URL: u, Timeout: timeout}).Check(ctx)
	if !r.OK {
		return fail(fmt.Sprintf("TCP port %s:%d is not open: %v — verify the host is reachable and the port is correct", host, port, r.Error))
	}
	return ok(r)
}

func handleTLS(ctx context.Context, args map[string]any) callToolResult {
	host := strArg(args, "host", "")
	port := intArg(args, "port", 0)
	if host == "" || port == 0 {
		return fail("host and port are required")
	}
	timeout, _ := time.ParseDuration(strArg(args, "timeout", "5s"))
	u, err := parseTCPURL(host, port)
	if err != nil {
		return fail(fmt.Sprintf("invalid host:port %s:%d — %v", host, port, err))
	}
	t := &libtower.TCP{
		URL:            u,
		Timeout:        timeout,
		CertFile:       strArg(args, "cert", ""),
		PrivateKeyFile: strArg(args, "key", ""),
	}
	tlsOK, err := t.TLSPortCheck(ctx)
	if err != nil {
		msg := fmt.Sprintf("TLS check failed for %s:%d: %v", host, port, err)
		if strArg(args, "cert", "") == "" {
			msg += " — provide cert and key for mutual TLS, or use tower_tcp for plain TCP"
		}
		return fail(msg)
	}
	return ok(libtower.Result{OK: tlsOK, Duration: t.Duration})
}

func handleHTTP(ctx context.Context, args map[string]any) callToolResult {
	urlStr := strArg(args, "url", "")
	if urlStr == "" {
		return fail("url is required")
	}
	method := util.HTTPMethod(strArg(args, "method", "GET"))
	timeout, _ := time.ParseDuration(strArg(args, "timeout", "10s"))
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	r := (&libtower.HTTP{URL: urlStr, Method: method}).Check(ctx)
	if !r.OK {
		return fail(fmt.Sprintf("HTTP check failed for %s: %v — verify the URL is accessible and the method is correct", urlStr, r.Error))
	}
	return ok(r)
}

func handleTrace(ctx context.Context, args map[string]any) callToolResult {
	urlStr := strArg(args, "url", "")
	if urlStr == "" {
		return fail("url is required")
	}
	method := util.HTTPMethod(strArg(args, "method", "GET"))

	r := (&libtower.HTTPTrace{URL: urlStr, Method: method}).Check(ctx)
	if !r.OK {
		return fail(fmt.Sprintf("HTTP trace failed for %s: %v — verify the URL is accessible", urlStr, r.Error))
	}
	return ok(r)
}

func handleHTTPS(ctx context.Context, args map[string]any) callToolResult {
	host := strArg(args, "host", "")
	if host == "" {
		return fail("host is required")
	}
	timeout, _ := time.ParseDuration(strArg(args, "timeout", "5s"))
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	hs := &libtower.HTTPS{
		Host:               host,
		Timeout:            timeout,
		InsecureSkipVerify: boolArg(args, "insecure_skip_verify"),
	}
	if p := intArg(args, "port", 0); p != 0 {
		hs.Port = strconv.Itoa(p)
	}
	if w := strArg(args, "warn_if_expiring", ""); w != "" {
		hs.WarnIfExpiringWithin, _ = time.ParseDuration(w)
	}

	r := hs.Check(ctx)
	if !r.OK {
		return fail(fmt.Sprintf("TLS cert check failed for %s: %v — verify the hostname is correct and reachable", host, r.Error))
	}
	return ok(r)
}

func handleWS(ctx context.Context, args map[string]any) callToolResult {
	urlStr := strArg(args, "url", "")
	if urlStr == "" {
		return fail("url is required")
	}
	timeout, _ := time.ParseDuration(strArg(args, "timeout", "5s"))

	r := (&libtower.WebSocket{URL: urlStr, Timeout: timeout}).Check(ctx)
	if !r.OK {
		return fail(fmt.Sprintf("WebSocket check failed for %s: %v — verify the URL uses ws:// or wss:// and the endpoint is reachable", urlStr, r.Error))
	}
	return ok(r)
}

func handleCheck(ctx context.Context, args map[string]any) callToolResult {
	yamlStr := strArg(args, "yaml", "")
	if yamlStr == "" {
		return fail("yaml is required — provide an inline tower config YAML string")
	}

	dir := os.TempDir() + "/tower-mcp"
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fail(fmt.Sprintf("failed to create temp directory: %v", err))
	}
	path := dir + "/check.yml"
	if err := os.WriteFile(path, []byte(yamlStr), 0600); err != nil {
		return fail(fmt.Sprintf("failed to write config: %v", err))
	}

	cfg, err := config.Parse(path)
	if err != nil {
		return fail(fmt.Sprintf("failed to parse config YAML: %v — check the YAML syntax and that each entry has a valid 'type' field", err))
	}

	var results []libtower.Result
	for _, c := range cfg.Checks {
		results = append(results, runCheck(ctx, c))
	}
	return ok(results)
}

func parseTCPURL(host string, port int) (*url.URL, error) {
	return url.Parse("tcp://" + host + ":" + strconv.Itoa(port))
}
