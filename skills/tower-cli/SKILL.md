# Tower CLI — Network Health Checks

Binary: `tower` — single binary, no runtime dependencies.

```bash
go install github.com/mismatched/tower@latest
```

## Quick Reference

| Goal | Command |
|------|---------|
| Ping a host | `tower ping example.com --count 3` (requires root) |
| DNS resolve | `tower dns example.com` |
| DNS via specific server | `tower dns example.com --from 8.8.8.8` |
| TCP port check | `tower tcp example.com:443 --timeout 5s` |
| TLS port check | `tower tls example.com:443 --cert client.crt --key client.key` |
| HTTP status check | `tower http https://example.com --method HEAD` |
| HTTP trace timing | `tower trace https://example.com --method GET` |
| TLS cert check | `tower https example.com --warn 720h` |
| TLS cert (insecure) | `tower https example.com --insecure` |
| WebSocket check | `tower ws wss://example.com/ws --timeout 10s` |
| Batch from config | `tower check config.yml` |
| MCP server | `tower serve` (stdio, hidden command) |

## All Commands

```
tower ping    <host>          [--count N]
tower dns     <addr>          [--from server] [--timeout DUR]
tower tcp     <host:port>     [--timeout DUR]
tower tls     <host:port>     [--cert FILE] [--key FILE] [--timeout DUR]
tower http    <url>           [--method M] [--timeout DUR]
tower trace   <url>           [--method M]
tower https   <host>          [--port N] [--warn DUR] [--insecure] [--timeout DUR]
tower ws      <url>           [--timeout DUR]
tower check   <config.yml>
tower serve                   (MCP stdio server)
```

## JSON Output

Every command prints JSON to stdout. Structure matches libtower's `Result`:

```json
{
  "OK": true,
  "Duration": 124304875,
  "Data": {"IP": {"IP": "93.184.216.34", "Zone": ""}},
  "Warning": null,
  "Error": null
}
```

- `Data` varies by check: `PingData`, `DNSData`, `CertData`, or `null`
- `Warning` is non-null on cert expiry warnings (HTTPS with `--warn`)
- `tower check` outputs a JSON array of results

## Config File Format

```yaml
checks:
  - type: tcp
    ip: "example.com"
    port: 443
    timeout: 5s
  - type: https
    host: "example.com"
    warn_if_expiring: 720h
    insecure_skip_verify: false
  - type: dns
    addr: "example.com"
  - type: ping
    host: "example.com"
    count: 3
  - type: http
    url: "https://example.com"
    method: GET
  - type: ws
    url: "wss://example.com/ws"
```

Valid types: `tcp`, `https`, `dns`, `ping`, `http`, `ws`. The `type` field is required on every entry.

## MCP Server (`tower serve`)

Runs a Model Context Protocol server over stdio. Zero SDK dependencies. 9 tools.

### Configuring in Claude Code

Add to `.mcp.json` or `settings.json`:

```json
{
  "mcpServers": {
    "tower": {
      "command": "tower",
      "args": ["serve"]
    }
  }
}
```

Or install from the marketplace plugin `tower-mcp`.

### MCP Tools

| Tool | Description | Root? |
|------|-------------|-------|
| `tower_ping` | ICMP ping | yes |
| `tower_dns` | DNS resolve, optional custom server | no |
| `tower_tcp` | TCP port check | no |
| `tower_tls` | TLS port check with client cert | no |
| `tower_http` | HTTP status + response metadata | no |
| `tower_trace` | Per-phase HTTP timing breakdown | no |
| `tower_https` | TLS cert validity + expiry warning | no |
| `tower_ws` | WebSocket upgrade handshake | no |
| `tower_check` | Batch from inline YAML config | varies |

All tools return JSON matching libtower's `Result` struct. Failed checks return `isError: true` with actionable error messages.

Example tool call result:
```json
{
  "content": [{
    "type": "text",
    "text": "{\"OK\":true,\"Duration\":294588875,\"Data\":null,\"Warning\":null,\"Error\":null}"
  }]
}
```

## Conventions

- **All output is JSON** — no human-readable text mode
- **Exit code** — non-zero on check failure, zero on success
- **signal.NotifyContext** — SIGINT/SIGTERM cancel in-flight checks
- **Ping requires root** — `sudo tower ping ...` or `CAP_NET_RAW` on Linux
- **Timeouts** — `--timeout` accepts Go duration strings (`5s`, `1m`, `500ms`)

## Anti-patterns

- Using `tower ping` without `sudo` — produces "socket: operation not permitted"
- Using old v1-style flat flags (`--ping`, `--dns`) — must use subcommands
- Expecting human-readable output — all output is JSON
- Running `tower check` with a YAML file that has no `type` field — each entry must have `type`
