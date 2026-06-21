# Tower

Network health-check tools for the command line.

[![Go Reference](https://pkg.go.dev/badge/github.com/mismatched/tower.svg)](https://pkg.go.dev/github.com/mismatched/tower)
[![test](https://github.com/mismatched/tower/actions/workflows/test.yml/badge.svg)](https://github.com/mismatched/tower/actions/workflows/test.yml)

## Install

```bash
go install github.com/mismatched/tower@latest
```

## Usage

All commands output JSON. Each result includes `OK`, `Duration`, `Data`, `Warning`, and `Error` fields.

```
tower [command] [args...] [flags...]
```

### Commands

#### ping — ICMP ping (requires root)

```bash
sudo tower ping example.com
sudo tower ping example.com --count 3
```

#### dns — DNS resolve

```bash
tower dns example.com
tower dns example.com --from 8.8.8.8
tower dns example.com --timeout 3s
```

#### tcp — TCP port check

```bash
tower tcp example.com:443
tower tcp example.com:22 --timeout 10s
```

#### tls — TLS port check with client certificate

```bash
tower tls example.com:443
tower tls example.com:443 --cert client.crt --key client.key
```

#### http — HTTP status check

```bash
tower http https://example.com
tower http https://example.com --method HEAD
tower http https://example.com --timeout 30s
```

#### trace — HTTP trace with per-phase timing

```bash
tower trace https://example.com
tower trace https://example.com --method POST
```

#### https — TLS certificate check

```bash
tower https example.com
tower https example.com --port 8443
tower https example.com --warn 720h
tower https example.com --insecure
```

#### ws — WebSocket check

```bash
tower ws wss://example.com/ws
tower ws ws://example.com:8080/ws --timeout 10s
```

#### check — Batch checks from config file

```bash
tower check config.yml
```

#### serve — MCP server (stdio)

```bash
tower serve
```

## Config File

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

## JSON Output

Every command prints a single JSON object (or array for `check`):

```json
{
  "OK": true,
  "Duration": 124304875,
  "Data": null,
  "Warning": null,
  "Error": null
}
```

## MCP Server

Tower includes a built-in MCP server over stdio. 9 tools, zero SDK dependencies.

```bash
tower serve
```

Configure in `.mcp.json`:

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

Or install from the **[tower skills marketplace](#claude-code-skills)**.

## Build

```bash
go build -o bin/tower .
```

## Test

```bash
# Non-root (all tests except ping)
go test -v ./...

# Root (includes ping tests)
sudo go test -v -race ./...
```

## GitHub Actions

Tower can be used as a GitHub Action to run health checks in your CI pipeline. See **[tower-template](https://github.com/mismatched/tower-template)** for starter workflows.

### Quick start

```yaml
# .github/workflows/health-check.yml
name: health-check
on:
  schedule:
    - cron: '0 */6 * * *'
  workflow_dispatch:

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: mismatched/tower@master
        with:
          tcp: 'example.com:443'
          http: 'https://example.com'
          timeout: '10s'
```

### TLS cert expiry warnings

Cert expiry warnings require a config file (the composite action doesn't expose `warn_if_expiring` as an input). Create a config and use the `config` input:

```yaml
# .github/tower/checks.yml
checks:
  - type: https
    host: "example.com"
    warn_if_expiring: 720h
```

```yaml
# .github/workflows/health-check.yml
steps:
  - uses: actions/checkout@v4
  - uses: mismatched/tower@master
    with:
      config: '.github/tower/checks.yml'
```

### Action inputs

| Input | Description |
|-------|-------------|
| `config` | Path to tower config YAML file (requires `actions/checkout` first) |
| `tcp` | TCP `host:port` to check |
| `http` | HTTP/HTTPS URL for status check |
| `https` | Hostname for TLS cert check |
| `dns` | Address to resolve |
| `ws` | WebSocket URL to check |
| `timeout` | Default timeout (default: `10s`) |
| `version` | Tower version to install (default: `latest`) |

The action fails if any check has `"OK": false`.

## Claude Code Skills

Tower ships with Claude Code skills for the CLI, MCP server, Go library, and GitHub Actions templates. Add as a plugin source:

```json
{
  "plugins": {
    "tower-skills": {
      "source": "https://github.com/mismatched/tower"
    }
  }
}
```

Installed skills:

| Plugin | Skills |
|--------|--------|
| `tower` | `tower-cli`, `libtower`, `tower-template` |
| `tower-mcp` | `tower-cli` (MCP server setup) |

## License

MIT
