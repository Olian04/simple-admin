# Step 1 — Tooling (templ, deps, CI)

Status: `done`

## Sub-tasks

- [x] Add `templ` CLI to `go.mod` `tool ()` block (`go get -tool github.com/a-h/templ/cmd/templ`) — v0.3.1020
- [x] Makefile: `generate` target (`go tool templ generate`); `build`/`run`/`lint`/`test`/`test-race` depend on `generate`; added to `.PHONY`
- [x] `.gitignore`: ignore `*_templ.go` (generated)
- [x] CI `go-quality.yml`: `templ generate` step before lint gate
- [x] `go mod tidy`, verify `templ generate` + `go build ./...` + `go vet ./...` + tests
- [ ] Runtime deps deferred until imported (else `go mod tidy` drops unused): `a-h/templ` runtime (added with first `.templ`), `pb33f/libopenapi`, `ohler55/ojg`, promote `santhosh-tekuri/jsonschema/v6` to direct

## Touched files

- `go.mod` / `go.sum` (templ tool + its deps), `Makefile`, `.gitignore`, `.github/workflows/go-quality.yml`

## Verification

- `go tool templ version` → v0.3.1020
- `go tool templ generate` exit 0 (no `.templ` yet → no-op)
- `go build ./...`, `go vet ./...`, `go test ./...` all pass

## Notes / follow-ups

- Generated `*_templ.go` are gitignored; build/lint/test all regenerate first via the `generate` dep.
- libopenapi / ojg / jsonschema-direct land in their consuming steps (5, 12, 9 respectively).
