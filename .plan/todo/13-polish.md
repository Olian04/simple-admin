# Step 13 — Polish

Status: `done`

## Sub-tasks

- [x] CSP + security headers middleware (`securityHeaders` in router): strict `default-src 'self'` etc. (no inline JS; moved the one inline style to `.sa-detail-label` in app.css), `X-Content-Type-Options: nosniff`, `Referrer-Policy`
- [x] README rewrite: elevator-pitch intro, **architecture mermaid diagram**, quick start (docker compose + local), config + compliant-backend contract, auth, field widgets, layout
- [x] `docs/AGENTS.md` rewritten for simple-admin + caveman directive + `.plan/` resume pointer + conventions
- [x] verified golangci-lint clean (0 issues) on the admin module
- [x] CSP header verified on running instance

## Deferred (noted, not blocking MVP)

- Pagination/sort UI controls (List params plumbed in upstream client + ListParams; no UI yet)
- Flash messages
- Discovery polling loop (config + swappable registry already exist)
- bearer/basic/oauth auth (stubbed; config rejects)

## Known issue (pre-existing / environmental)

- `govulncheck` (part of `make lint`) reports **Go standard-library advisories** (net/http, crypto/x509, etc.) tied to the Go toolchain release, not to project code. Resolution is a Go toolchain update on the build/CI runner. Tried bumping the go directive (1.26→1.26.1, toolchain auto-fetched 1.26.4) but advisories persisted (fix versions ahead of available patch), so reverted to `go 1.26` to keep the diff minimal. **golangci-lint passes; this is a Go-version maintenance item, not introduced by this work.**

## Touched files

- `internal/transport/http/{router.go,views/resource.templ}`, `web/static/css/app.css`
- `README.md`, `docs/AGENTS.md`

## Verification

- `go build ./...`, `go vet ./...`, `go tool golangci-lint run` (0 issues), all unit tests pass
- CSP + nosniff headers present on `/ui`
