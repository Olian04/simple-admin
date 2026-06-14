# Step 6 — Static assets + shell

Status: `done`

## Sub-tasks

- [x] Vendored `web/static/css/bulma.min.css` (1.0.2) + `web/static/js/htmx.min.js` (2.0.4); thin `web/static/css/app.css` (sidebar layout)
- [x] `web/embed.go` — `//go:embed static` (lives at `web/` because embed can't reach parent dirs from internal/)
- [x] `internal/transport/http/views/`: `models.go` (PageData/SidebarGroup/SidebarItem) + `layout.templ` (Page, Sidebar grouped by backend, Placeholder)
- [x] `internal/transport/http/handlers/ui.go` — UIHandler.Index renders shell; `pageData` builds sidebar from registry; `render` helper
- [x] `internal/app/resources.go` — `buildRegistry` runs manual + openapi sources → Registry (Runtimes carry descriptors; Store/Service added step 8)
- [x] Wire: `router.go` (+resources registry, `GET /static/`, `GET /ui`), `bootstrap.go` (+configDir, buildRegistry), `main.go` (pass configDir)
- [x] templ runtime now a direct dep (go mod tidy)

## Touched files

- `web/embed.go`, `web/static/**`
- `internal/transport/http/views/{models.go,layout.templ}`, `handlers/ui.go`, `router.go`
- `internal/app/{bootstrap.go,resources.go}`, `cmd/simple-admin/main.go`

## Verification

- `go tool templ generate`, `go build ./...`, `go vet ./...` exit 0
- Integration smoke: ran example-backend (8090) + admin (8080) with `configs/config.example.yaml`:
  - `GET /ui` → 200 HTML containing both backends (Example API + Service Gateway) and resources (Posts, Authors) — proves **manual + discovery sources both load** at startup against real PocketBase
  - `GET /static/css/bulma.min.css` → 200 (677KB)

## Decisions / follow-ups

- Static embed package at `web/` (not `internal/transport/http/assets/`) due to go:embed parent-dir rule. → noted.
- `/ui` currently renders shell+Placeholder; step 8 adds `/ui/{backend}/{name}` routes + redirect to first resource.
- `*_templ.go` are gitignored — CI/build regenerate.
