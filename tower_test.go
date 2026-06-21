package main

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/mismatched/tower/config"
	"github.com/urfave/cli/v3"
)

func TestConfigParse(t *testing.T) {
	yml := `
checks:
  - type: tcp
    ip: "example.com"
    port: 443
    timeout: 5s
  - type: https
    host: "example.com"
    warn_if_expiring: 720h
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
`
	dir := t.TempDir()
	p := filepath.Join(dir, "test.yml")
	if err := os.WriteFile(p, []byte(yml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Parse(p)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(cfg.Checks) != 6 {
		t.Fatalf("got %d checks, want 6", len(cfg.Checks))
	}
	if cfg.Checks[0].Type != "tcp" {
		t.Errorf("checks[0].type = %q, want tcp", cfg.Checks[0].Type)
	}
	if cfg.Checks[1].Type != "https" {
		t.Errorf("checks[1].type = %q, want https", cfg.Checks[1].Type)
	}
	if cfg.Checks[2].Type != "dns" {
		t.Errorf("checks[2].type = %q, want dns", cfg.Checks[2].Type)
	}
	if cfg.Checks[3].Type != "ping" {
		t.Errorf("checks[3].type = %q, want ping", cfg.Checks[3].Type)
	}
	if cfg.Checks[4].Type != "http" {
		t.Errorf("checks[4].type = %q, want http", cfg.Checks[4].Type)
	}
	if cfg.Checks[5].Type != "ws" {
		t.Errorf("checks[5].type = %q, want ws", cfg.Checks[5].Type)
	}
}

func TestConfigParseInvalid(t *testing.T) {
	tests := []struct {
		name string
		yml  string
	}{
		{"bad yaml", "checks: [["},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			p := filepath.Join(dir, "test.yml")
			os.WriteFile(p, []byte(tt.yml), 0644)
			_, err := config.Parse(p)
			if err == nil {
				t.Error("expected error for invalid config")
			}
		})
	}
}

func TestConfigParseMissingFile(t *testing.T) {
	_, err := config.Parse("/nonexistent/tower-config.yml")
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestRunCheckUnknownType(t *testing.T) {
	ctx := context.Background()
	c := config.CheckConfig{Type: "unknown"}
	r := runCheck(ctx, c)
	if r.OK {
		t.Error("expected !OK for unknown check type")
	}
	if r.Error == nil {
		t.Error("expected error for unknown check type")
	}
}

func TestSubcommandMissingArgs(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cli.Command
	}{
		{"tcp", tcpCmd()},
		{"dns", dnsCmd()},
		{"http", httpCmd()},
		{"https", httpsCmd()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.cmd.Writer = io.Discard
			// Running subcommand without required arg should error
			err := tt.cmd.Run(context.Background(), []string{tt.name})
			if err == nil {
				t.Error("expected error for missing args")
			}
		})
	}
}

// Helper functions return the subcommand definitions for testing.
func tcpCmd() *cli.Command {
	return &cli.Command{
		Name:   "tcp",
		Flags:  []cli.Flag{&cli.StringFlag{Name: "timeout", Value: "5s"}},
		Action: tcpAction,
	}
}
func dnsCmd() *cli.Command {
	return &cli.Command{
		Name:   "dns",
		Flags:  []cli.Flag{&cli.StringFlag{Name: "timeout", Value: "5s"}},
		Action: dnsAction,
	}
}
func httpCmd() *cli.Command {
	return &cli.Command{
		Name:   "http",
		Flags:  []cli.Flag{&cli.StringFlag{Name: "timeout", Value: "10s"}},
		Action: httpAction,
	}
}
func httpsCmd() *cli.Command {
	return &cli.Command{
		Name: "https",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "timeout", Value: "5s"},
		},
		Action: httpsAction,
	}
}
