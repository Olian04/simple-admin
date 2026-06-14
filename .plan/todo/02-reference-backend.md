# Step 2 — Reference backend (separate PocketBase module)

Status: `done`

## Sub-tasks

- [x] `example-backend/` separate Go module (`go.mod` = github.com/Olian04/simple-admin-example-backend); PocketBase v0.39.3 (CGO-free modernc sqlite)
- [x] `main.go` — embed PocketBase; seed `authors` + `posts` collections on bootstrap (public rules); optional `EXAMPLE_API_KEY` gate on `/api/collections/*`
- [x] `openapi/spec.go` — `Generate(app, advertiseAPIKey)` builds OpenAPI 3.0 from base collections: component schema per collection (field→JSON-schema mapping), CRUD paths, apiKey securityScheme when gated
- [x] `Dockerfile` (alpine, static CGO-free binary)
- [x] root `Dockerfile` for admin UI (templ generate + static build)
- [x] `deploy/docker-compose.yml` — example-backend + admin services (admin↔backend config wiring deferred to step 3, commented)
- [x] `example-backend/README.md`, `example-backend/.gitignore` (pb_data, binaries)
- [x] smoke test: spec + create + list verified against running server

## Touched files

- `example-backend/{go.mod,go.sum,main.go,openapi/spec.go,Dockerfile,README.md,.gitignore}`
- `Dockerfile`, `deploy/docker-compose.yml`

## Verification

- `go build ./...` in `example-backend/` → exit 0
- Ran `serve`: `GET /openapi.json` lists authors+posts with CRUD paths + schemas; `POST` create + `GET` list (envelope `{items,totalItems,...}`) work
- `docker compose -f deploy/docker-compose.yml config` → valid

## Decisions / follow-ups

- **List response is a `{items:[…]}` envelope** (PocketBase), not a bare array. Admin upstream List client (step 8) must extract `items` (auto-detect array vs `items` envelope). Contract note: list MAY return an envelope. → recorded in decisions.md.
- PocketBase update verb is **PATCH** (no PUT). Capabilities/update path must use PATCH for this backend.
- Relation field named `author` (not `authorId`); relation auto-select convention (step 12) should singularize/relax matching, or rename later.
- Admin↔backend compose wiring + `deploy/admin-config.yaml` land in step 3 once `backends:` config exists.
