# example-backend — reference compliant backend

A **separate Go module** (own `go.mod`, isolated from the admin UI's dependency
graph) that demonstrates the contract the `simple-admin` UI targets. Built on
[PocketBase](https://pocketbase.io) used as a Go framework.

## What it exposes

- `GET /openapi.json` — an OpenAPI 3 spec **generated from the PocketBase
  collection schemas** (see `openapi/spec.go`). Each base collection maps to its
  real REST endpoints and component schema.
- CRUD REST API per collection (PocketBase's built-in routes):
  - `GET    /api/collections/{name}/records`        — list (paginated `{items:[…]}` envelope)
  - `POST   /api/collections/{name}/records`        — create
  - `GET    /api/collections/{name}/records/{id}`   — read
  - `PATCH  /api/collections/{name}/records/{id}`   — update
  - `DELETE /api/collections/{name}/records/{id}`   — delete

Sample collections seeded on first boot: **authors** and **posts** (posts has a
`author` relation, a `status` enum, a `body` richtext, numeric/bool/date fields).
Access rules are public so the admin UI can drive it in local dev without auth.

## Run

```sh
go run . serve --http=127.0.0.1:8090 --dir=./pb_data
# then:
curl http://127.0.0.1:8090/openapi.json
```

Or via the local stack: `docker compose -f ../deploy/docker-compose.yml up --build`.

## Optional API-key gate

Set `EXAMPLE_API_KEY` to require an `X-API-Key` header on `/api/collections/*`.
The generated spec then advertises a matching `apiKey` security scheme. Used to
exercise the admin UI's API-key auth passthrough (admin step 11).

```sh
EXAMPLE_API_KEY=dev-secret-key go run . serve --http=127.0.0.1:8090
```

PocketBase's own token auth (superusers/auth collections) is left intact and is
the intended future test bed for bearer/token auth passthrough.
