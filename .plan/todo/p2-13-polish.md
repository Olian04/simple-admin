# Step p2-13 — Comments cleanup + docs + CI + .plan

Status: `done`

## Sub-tasks

- [x] Stale comments: `registry.go` "future discovery poll" → "the discovery poller"; (resources.go "later step" fixed in p2-6)
- [x] README: package layout (backend/schemavalidation/views/views form/poller), poll_interval built, widget list (color/json/tags/code/markdown real; stubs), widget data contract (FieldContext + siblings), GET / redirect
- [x] docs/AGENTS.md: layout + dependency direction + conventions (widgets, shutdown, sort, dist/simple-admin) refreshed
- [x] CI/govulncheck: `make lint` govulncheck is report-only (`-` prefix) with note; `go.mod` bumped `go 1.26` → `go 1.26.1` (patched stdlib)
- [x] `.plan` refreshed (progress board, decisions 31–32, todos)

## Verification

- `go tool templ generate` + build + vet + tests + golangci (0 issues) all pass on go 1.26.1
- govulncheck remains report-only (stdlib advisories tracked to toolchain)

## Commit

- (next)
