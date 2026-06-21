// tower
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/mismatched/libtower"
	"github.com/mismatched/tower/config"
	"github.com/mismatched/tower/util"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := &cli.Command{
		Name:    "tower",
		Usage:   "network uptime and status checker",
		Version: "0.0.1",
		Authors: []any{
			"Dariush Abbasi <poshtehani@gmail.com>",
			"Hasan Aminfar <aminfar69@gmail.com>",
		},
		Commands: []*cli.Command{
			{
				Name:      "ping",
				Usage:     "ICMP ping a host (requires root)",
				ArgsUsage: "<host>",
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "count",
						Aliases: []string{"c"},
						Value:   1,
						Usage:   "number of pings",
					},
				},
				Action: pingAction,
			},
			{
				Name:      "dns",
				Usage:     "DNS resolve an address",
				ArgsUsage: "<addr>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "from",
						Usage: "query specific DNS server (ip:port)",
					},
					&cli.StringFlag{
						Name:  "timeout",
						Value: "5s",
						Usage: "timeout duration",
					},
				},
				Action: dnsAction,
			},
			{
				Name:      "tcp",
				Usage:     "check if a TCP port is open",
				ArgsUsage: "<host:port>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "timeout",
						Value: "5s",
						Usage: "timeout duration",
					},
				},
				Action: tcpAction,
			},
			{
				Name:      "tls",
				Usage:     "check a TLS port with optional client certificate",
				ArgsUsage: "<host:port>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "cert",
						Usage: "client certificate file",
					},
					&cli.StringFlag{
						Name:  "key",
						Usage: "client private key file",
					},
					&cli.StringFlag{
						Name:  "timeout",
						Value: "5s",
						Usage: "timeout duration",
					},
				},
				Action: tlsAction,
			},
			{
				Name:      "http",
				Usage:     "check HTTP status of a URL",
				ArgsUsage: "<url>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "method",
						Aliases: []string{"X"},
						Usage:   "HTTP method",
					},
					&cli.StringFlag{
						Name:  "timeout",
						Value: "10s",
						Usage: "timeout duration",
					},
				},
				Action: httpAction,
			},
			{
				Name:      "trace",
				Usage:     "trace HTTP request with per-phase timing",
				ArgsUsage: "<url>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "method",
						Aliases: []string{"X"},
						Usage:   "HTTP method",
					},
				},
				Action: traceAction,
			},
			{
				Name:      "https",
				Usage:     "check TLS certificate validity",
				ArgsUsage: "<host>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "port",
						Usage: "port (default 443)",
					},
					&cli.StringFlag{
						Name:  "timeout",
						Value: "5s",
						Usage: "timeout duration",
					},
					&cli.StringFlag{
						Name:  "warn",
						Usage: "warn if cert expires within duration (e.g., 720h)",
					},
					&cli.BoolFlag{
						Name:  "insecure",
						Usage: "skip TLS certificate verification",
					},
				},
				Action: httpsAction,
			},
			{
				Name:      "ws",
				Usage:     "check WebSocket connectivity",
				ArgsUsage: "<url>",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "timeout",
						Value: "5s",
						Usage: "timeout duration",
					},
				},
				Action: wsAction,
			},
			{
				Name:      "check",
				Usage:     "run batch checks from a config file",
				ArgsUsage: "<config.yml>",
				Action:    checkAction,
			},
			{
				Name:   "serve",
				Usage:  "run MCP server over stdio",
				Action: serveAction,
				Hidden: true,
			},
		},
	}

	err := app.Run(ctx, os.Args)
	if err != nil {
		log.Fatalln(err)
	}
}

// pingAction sends ICMP echo requests (requires root).
func pingAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return fmt.Errorf("host required")
	}
	host := cmd.Args().First()
	count := cmd.Int("count")
	if count < 1 {
		count = 1
	}

	ip, dur, err := libtower.PingContext(ctx, host, count)
	r := libtower.Result{OK: err == nil, Duration: dur, Data: libtower.PingData{IP: ip}, Error: err}
	json.NewEncoder(cmd.Writer).Encode(r)
	if !r.OK {
		return r.Error
	}
	return nil
}

// dnsAction resolves an address via DNS, optionally via a specific server.
func dnsAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return fmt.Errorf("address required")
	}
	addr := cmd.Args().First()
	timeout, _ := time.ParseDuration(cmd.String("timeout"))
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	server := cmd.String("from")
	var r libtower.Result
	if server != "" {
		ip, dur, err := libtower.DNSLookupFromContext(ctx, addr, server)
		r = libtower.Result{OK: err == nil, Duration: dur, Data: libtower.DNSData{IP: ip}, Error: err}
	} else {
		ip, dur, err := libtower.DNSLookupContext(ctx, addr)
		r = libtower.Result{OK: err == nil, Duration: dur, Data: libtower.DNSData{IP: ip}, Error: err}
	}
	json.NewEncoder(cmd.Writer).Encode(r)
	if !r.OK {
		return r.Error
	}
	return nil
}

// tcpAction checks if a TCP port is open.
func tcpAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return fmt.Errorf("host:port required")
	}
	u, err := url.Parse("tcp://" + cmd.Args().First())
	if err != nil {
		return err
	}
	timeout, _ := time.ParseDuration(cmd.String("timeout"))

	t := &libtower.TCP{URL: u, Timeout: timeout}
	r := t.Check(ctx)
	json.NewEncoder(cmd.Writer).Encode(r)
	if !r.OK {
		return r.Error
	}
	return nil
}

// tlsAction checks a TLS port, optionally with client certificates.
func tlsAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return fmt.Errorf("host:port required")
	}
	u, err := url.Parse("tcp://" + cmd.Args().First())
	if err != nil {
		return err
	}
	timeout, _ := time.ParseDuration(cmd.String("timeout"))

	t := &libtower.TCP{
		URL:            u,
		Timeout:        timeout,
		CertFile:       cmd.String("cert"),
		PrivateKeyFile: cmd.String("key"),
	}
	ok, err := t.TLSPortCheck(ctx)
	r := libtower.Result{OK: ok, Duration: t.Duration, Error: err}
	json.NewEncoder(cmd.Writer).Encode(r)
	if !r.OK {
		return r.Error
	}
	return nil
}

// httpAction performs an HTTP status check.
func httpAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return fmt.Errorf("url required")
	}
	urlStr := cmd.Args().First()
	method := util.HTTPMethod(cmd.String("method"))
	timeout, _ := time.ParseDuration(cmd.String("timeout"))
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	h := &libtower.HTTP{URL: urlStr, Method: method}
	r := h.Check(ctx)
	json.NewEncoder(cmd.Writer).Encode(r)
	if !r.OK {
		return r.Error
	}
	return nil
}

// traceAction traces an HTTP request with per-phase timing.
func traceAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return fmt.Errorf("url required")
	}
	urlStr := cmd.Args().First()
	method := util.HTTPMethod(cmd.String("method"))

	ht := &libtower.HTTPTrace{URL: urlStr, Method: method}
	r := ht.Check(ctx)
	json.NewEncoder(cmd.Writer).Encode(r)
	if !r.OK {
		return r.Error
	}
	return nil
}

// httpsAction checks TLS certificate validity.
func httpsAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return fmt.Errorf("host required")
	}
	host := cmd.Args().First()
	timeout, _ := time.ParseDuration(cmd.String("timeout"))
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	hs := &libtower.HTTPS{
		Host:                host,
		Port:                cmd.String("port"),
		Timeout:             timeout,
		InsecureSkipVerify:  cmd.Bool("insecure"),
		WarnIfExpiringWithin: 0,
	}
	if w := cmd.String("warn"); w != "" {
		hs.WarnIfExpiringWithin, _ = time.ParseDuration(w)
	}

	r := hs.Check(ctx)
	json.NewEncoder(cmd.Writer).Encode(r)
	if !r.OK {
		return r.Error
	}
	return nil
}

// wsAction checks WebSocket connectivity.
func wsAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return fmt.Errorf("url required")
	}
	wsURL := cmd.Args().First()
	timeout, _ := time.ParseDuration(cmd.String("timeout"))

	ws := &libtower.WebSocket{URL: wsURL, Timeout: timeout}
	r := ws.Check(ctx)
	json.NewEncoder(cmd.Writer).Encode(r)
	if !r.OK {
		return r.Error
	}
	return nil
}

// checkAction runs batch checks from a config file.
func checkAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.NArg() == 0 {
		return fmt.Errorf("config file required")
	}
	cfg, err := config.Parse(cmd.Args().First())
	if err != nil {
		return err
	}

	results := make([]libtower.Result, 0, len(cfg.Checks))
	for _, c := range cfg.Checks {
		results = append(results, runCheck(ctx, c))
	}
	return json.NewEncoder(cmd.Writer).Encode(results)
}

// serveAction runs the MCP stdio server.
func serveAction(ctx context.Context, cmd *cli.Command) error {
	return runServe()
}

// runCheck dispatches a CheckConfig to the appropriate libtower check.
func runCheck(ctx context.Context, c config.CheckConfig) libtower.Result {
	switch c.Type {
	case "tcp":
		u, err := url.Parse("tcp://" + c.IP + ":" + strconv.Itoa(c.Port))
		if err != nil {
			return libtower.Result{OK: false, Error: err}
		}
		t := &libtower.TCP{URL: u, Timeout: time.Duration(c.Timeout)}
		return t.Check(ctx)

	case "https":
		hs := &libtower.HTTPS{
			Host:                  c.Host,
			Timeout:               time.Duration(c.Timeout),
			InsecureSkipVerify:    c.InsecureSkipVerify,
			WarnIfExpiringWithin:  c.WarnIfExpiringWithin,
		}
		if c.Port != 0 {
			hs.Port = strconv.Itoa(c.Port)
		}
		return hs.Check(ctx)

	case "dns":
		ip, dur, err := libtower.DNSLookupContext(ctx, c.Addr)
		return libtower.Result{OK: err == nil, Duration: dur, Data: libtower.DNSData{IP: ip}, Error: err}

	case "ping":
		count := c.Count
		if count < 1 {
			count = 1
		}
		ip, dur, err := libtower.PingContext(ctx, c.Host, count)
		return libtower.Result{OK: err == nil, Duration: dur, Data: libtower.PingData{IP: ip}, Error: err}

	case "http":
		method := util.HTTPMethod(c.Method)
		h := &libtower.HTTP{URL: c.URL, Method: method}
		return h.Check(ctx)

	case "ws":
		ws := &libtower.WebSocket{URL: c.URL, Timeout: time.Duration(c.Timeout)}
		return ws.Check(ctx)

	default:
		return libtower.Result{OK: false, Error: fmt.Errorf("unknown check type: %s", c.Type)}
	}
}
