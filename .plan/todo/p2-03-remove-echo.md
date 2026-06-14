# Step p2-3 — Remove echo scaffold + rename product → simple-admin

Status: `done`

## Sub-tasks

- [x] Delete `internal/domain/echo/`, `internal/transport/http/handlers/echo.go`, `test/unit/domain/echo/`
- [x] `router.go`: drop echo import + `POST /echo` + `svc echo.Service` param
- [x] `bootstrap.go`: drop echo import + `echo.NewService()`; logs "echo service started" → "simple-admin started"
- [x] `main.go`: cli `Name`/`Usage` → simple-admin
- [x] `defaults.go`: `DefaultMetricPrefix` `echo` → `simple_admin` (valid Prometheus prefix)
- [x] `configs/config.example.yaml`: `labels.service` → simple-admin, `metric_prefix` → simple_admin
- [x] `Makefile`: build output `dist/echo` → `dist/simple-admin` + comments
- [x] `goreleaser/.goreleaser.yaml`: fix stale ldflags module path (`go-app-template` → `simple-admin`), `echo-template_*`/gatekeeper text → simple-admin
- [x] `goreleaser/Dockerfile`: image title/description → simple-admin
- [x] `.github/workflows/release.yml`: stale `architecture-template/` paths → root; `echo-template` image/summary → simple-admin

## Verification

- build/vet/tests pass; no `echo` left in Go (non-templ, non-example-backend)

## Notes

- `docs/AGENTS.md`/`README.md` echo mentions handled in p2-13 (docs sweep).

## Commit

- (next)
