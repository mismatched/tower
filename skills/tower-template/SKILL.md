# Tower Template — GitHub Actions Health Checks

Template repo: `github.com/mismatched/tower-template` — pre-built GitHub Actions workflows for running tower health checks in CI.

## Quick Reference

| Goal | Approach |
|------|----------|
| Add health checks to a repo | Click **Use this template** on `tower-template`, edit targets, commit |
| Check TCP + HTTP + HTTPS + DNS | Use `tower-health.yml` (every 6h) |
| Check TLS cert expiry | Use `tls-cert-expiry.yml` (daily at 8am) or add `warn_if_expiring` via config |
| Manual one-off check | Use `workflow_dispatch` trigger in either workflow |
| Cert expiry warning | Use `config` input with a YAML file (composite action doesn't expose `warn_if_expiring` directly) |
| Install tower action in existing workflow | `uses: mismatched/tower@master` with `tcp`, `http`, `https`, `dns`, `ws`, `config`, or `timeout` inputs |

## Included Workflows

| Workflow | Schedule | Checks | Inputs |
|----------|----------|--------|--------|
| `tower-health.yml` | Every 6 hours + manual | `tcp`, `http`, `https`, `dns` | Quick in-line inputs |
| `tls-cert-expiry.yml` | Daily at 8am + manual | `https` (no warn) | Demonstrates cert-only check |

## Action Inputs (`mismatched/tower`)

| Input | Description | Example |
|-------|-------------|--------|
| `config` | Path to tower config YAML (needs `actions/checkout` first) | `.github/tower/checks.yml` |
| `tcp` | TCP `host:port` to check | `example.com:443` |
| `http` | HTTP/HTTPS URL for status check | `https://example.com` |
| `https` | Hostname for TLS cert check | `example.com` |
| `dns` | Address to resolve | `example.com` |
| `ws` | WebSocket URL to check | `wss://example.com/ws` |
| `timeout` | Default timeout duration | `10s` (default) |
| `version` | Tower version to install | `latest` (default) |

The action fails (exit code 1) if any check returns `"OK": false`.

## Cert Expiry Warnings

The composite action's inline inputs (`tcp`, `http`, `https`, `dns`, `ws`) don't support `warn_if_expiring`. To get cert expiry alerts, use a config file:

```yaml
# .github/tower/checks.yml
checks:
  - type: https
    host: "example.com"
    warn_if_expiring: 720h   # 30 days
```

```yaml
# .github/workflows/health-check.yml
steps:
  - uses: actions/checkout@v4
  - uses: mismatched/tower@master
    with:
      config: '.github/tower/checks.yml'
```

## Usage Pattern

1. Create a new repo from `tower-template` (GitHub "Use this template" button)
2. Edit `.github/workflows/tower-health.yml` — replace `example.com` targets
3. For cert expiry: create `.github/tower/checks.yml` with `warn_if_expiring`, add `config` input
4. Commit — workflows run on schedule and on push to master
5. Check the Actions tab for results

## Adding to an Existing Repo

Copy the workflow files into `.github/workflows/`:

```bash
cp tower-template/.github/workflows/tower-health.yml myproject/.github/workflows/
```

Or add a single step to any existing workflow:

```yaml
- uses: mismatched/tower@master
  with:
    tcp: 'myapp.example.com:443'
    timeout: '10s'
```

## Anti-patterns

- Using `https` input with cert expiry expectations — it only checks validity, not expiry windows; use a config file with `warn_if_expiring` instead
- Forgetting `actions/checkout@v4` before `config` input — the config file must be checked out first
- Setting `timeout` too low for remote checks — 10s is reasonable for internet targets
