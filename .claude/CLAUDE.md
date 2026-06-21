# CLAUDE.md — Tower

## Project overview

Tower is a CLI network health-check tool built on `github.com/mismatched/libtower`. It exposes TCP, TLS, HTTP, HTTPS, DNS, ICMP ping, and WebSocket checks as subcommands with JSON output.

## Commands

```bash
go build -o bin/tower .                  # Build CLI
go build -o bin/tower-mcp ./cmd/tower-mcp  # Build MCP server
go vet ./...                             # Vet
go test -v ./...                         # Test (non-root)
sudo go test -v -race ./...              # Test (root, includes ping)
./bin/tower --help                       # Show help
```

## Architecture

```
tower.go              → CLI entry point, 9 subcommands, signal handling
tower_test.go         → Tests for config parsing and CLI behavior
cmd/tower-mcp/main.go → MCP server (9 tools = 9 libtower checks)
config/config.go      → YAML config parser (flat list with type discriminator)
config/ping.go        → PING config type
config/tcp.go         → TCP config type
util/method.go        → HTTP method normalization
```

## Conventions

- All CLI output is JSON via `encoding/json`
- Every check uses `context.Context` variants (no `context.Background()` in actions)
- `signal.NotifyContext` provides graceful shutdown on SIGINT/SIGTERM
- Subcommands validate `cmd.NArg() == 0` before accessing args
- Flags use `time.ParseDuration` for timeout parsing
- JSON output goes to `cmd.Writer` (not `os.Stdout` directly)

## Dependencies

- `github.com/mismatched/libtower` — all health check primitives
- `github.com/urfave/cli/v3` — CLI framework
- `gopkg.in/yaml.v2` — config file parsing
