# libtower — Go Network Health Checks

Package: `github.com/mismatched/libtower` — flat package, no sub-packages.

## Quick Reference

| Goal | Approach |
|------|----------|
| TCP port check | `&libtower.TCP{URL: u, Timeout: 5*time.Second}` → `TCPPortCheck(ctx)` |
| TLS with client cert | Same `TCP` struct, set `CertFile`/`PrivateKeyFile`, call `TLSPortCheck(ctx)` |
| HTTP status | `&libtower.HTTP{URL: url, Method: "GET"}` → `HTTPStatus()` or `Check(ctx)` |
| Per-phase HTTP timing | `&libtower.HTTPTrace{URL: url, Method: "GET"}` → `Trace()` or `Check(ctx)` |
| TLS cert check + expiry | `&libtower.HTTPS{Host: host, Timeout: 5*time.Second}` → `HTTPSCheck(ctx)` |
| Cert expiry warning | Set `WarnIfExpiringWithin` → check `r.Warning` after `hs.Check(ctx)` |
| DNS resolution | `libtower.DNSLookup("example.com")` |
| DNS via specific server | `libtower.DNSLookupFrom("example.com", "8.8.8.8")` |
| ICMP ping | `libtower.Ping("example.com", 1)` — root required |
| WebSocket | `&libtower.WebSocket{URL: "ws://...", Timeout: 5*time.Second}` → `WSCheck(ctx)` |
| Composable checks | All types implement `Checker` — iterate `[]Checker` and call `Check(ctx)` |

## Conventions

- **Pointer receivers** — methods mutate the receiver, populating `Start`, `End`, `Duration`
- **Context** — all I/O methods accept `context.Context`. Short-name wrappers use `context.Background()`; context-aware variants carry a `Context` suffix (`PingContext`, `DNSLookupContext`, `DNSLookupFromContext`, `HTTPStatusContext`, `TraceContext`)
- **URLs** — TCP expects `*url.URL` with a `tcp://` scheme: `url.Parse("tcp://example.com:443")`
- **Timing** — every struct has `Start`, `End`, `Duration` set automatically by the check method

## Core Types

### Result

```go
type Result struct {
    OK       bool
    Duration time.Duration
    Data     CheckData   // PingData, DNSData, CertData, or nil
    Warning  error       // non-nil on soft warnings (cert expiring)
    Error    error
}
```

### Checker Interface

```go
type Checker interface {
    Check(ctx context.Context) Result
}
```

All check types implement `Checker`. Use it for uniform batch execution.

### CheckData Variants

| Type | Field | Kind |
|------|-------|------|
| `PingData` | `IP *net.IPAddr` | `"ping"` |
| `DNSData` | `IP *net.IPAddr` | `"dns"` |
| `CertData` | `NotAfter time.Time` | `"tls_cert"` |

### TCP

```go
u, _ := url.Parse("tcp://example.com:443")
t := &libtower.TCP{URL: u, Timeout: 5 * time.Second}

// Plain TCP
ok, err := t.TCPPortCheck(ctx)

// TLS with client cert
t.CertFile = "/path/to/cert.pem"
t.PrivateKeyFile = "/path/to/key.pem"
ok, err = t.TLSPortCheck(ctx)

// Or via Checker interface
r := t.Check(ctx)
```

### HTTPS

```go
hs := &libtower.HTTPS{
    Host:                  "example.com",
    Timeout:               5 * time.Second,
    WarnIfExpiringWithin:  30 * 24 * time.Hour, // 30 days
}
r := hs.Check(ctx)
if r.Warning != nil {
    // cert expiring soon
}
```

- `Port` defaults to `"443"`
- `InsecureSkipVerify` skips TLS verification
- `HTTPSCheck(ctx)` returns `(bool, time.Time, error)` — the time is the earliest cert NotAfter

### DNS

```go
// System resolver
ip, dur, err := libtower.DNSLookup("example.com")

// Specific server
ip, dur, err = libtower.DNSLookupFrom("example.com", "8.8.8.8")

// With context
ip, dur, err = libtower.DNSLookupContext(ctx, "example.com")

// Struct form (Checker interface)
d := &libtower.DNS{ADDR: "example.com"}
r := d.Check(ctx)
```

### HTTP

```go
h := &libtower.HTTP{
    URL:    "https://example.com",
    Method: "GET",
}
err := h.HTTPStatus()
// h.StatusCode, h.Status, h.Proto, h.Header, h.Duration populated
```

`HTTPTrace` captures per-phase timing:

```go
ht := &libtower.HTTPTrace{
    URL:    "https://example.com",
    Method: "GET",
}
ht.Trace()
// ht.DNS, ht.TLSHandshake, ht.Connect, ht.GotFirstResponseByte, ht.Total
```

### Ping

```go
// Standalone (root required)
ip, dur, err := libtower.Ping("example.com", 1)

// Struct form (Checker interface)
pc := &libtower.PingCheck{Addr: "example.com", Seq: 1}
r := pc.Check(ctx)
```

Requires root or `CAP_NET_RAW` on Linux. Without root: `"socket: operation not permitted"`.

### WebSocket

```go
ws := &libtower.WebSocket{
    URL:     "wss://example.com/ws",
    Timeout: 5 * time.Second,
}
err := ws.WSCheck(ctx)
```

Scheme defaults to `ws` if omitted. Port defaults to 80 (ws) or 443 (wss).

## Composing Checks

```go
checks := []libtower.Checker{
    &libtower.TCP{URL: u, Timeout: time.Second},
    &libtower.HTTPS{Host: "example.com", WarnIfExpiringWithin: 30 * 24 * time.Hour},
    &libtower.DNS{ADDR: "example.com"},
}
for _, c := range checks {
    r := c.Check(ctx)
    if !r.OK {
        log.Printf("FAIL %T: %v", c, r.Error)
        continue
    }
    if r.Warning != nil {
        log.Printf("WARN %T: %v", c, r.Warning)
    }
    log.Printf("OK   %T in %v", c, r.Duration)
}
```

## Anti-patterns

- Using raw `net.DialTimeout` or `crypto/tls.Dial` instead of libtower types — lose timing and unified Result
- Hand-rolling ICMP with `golang.org/x/net/icmp` — libtower handles IPv4/IPv6 detection
- Using `Ping()` without root — produces `"socket: operation not permitted"`
- Failing to inspect `r.Warning` from `HTTPS.Check()` when `WarnIfExpiringWithin` is set

## Install

```bash
go get github.com/mismatched/libtower
```

Import: `"github.com/mismatched/libtower"`
