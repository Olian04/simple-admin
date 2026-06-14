# Step 11 — Credentials + API-key passthrough

Status: `done`

## Sub-tasks

- [x] `web/static/js/islands/credentials.js`: localStorage `simple-admin:cred:{backendId}`; HTMX `htmx:configRequest` injector (backend id from /ui/{id}/ path → X-Sa-Credential + X-Sa-Backend headers); credentials-page save/clear/populate
- [x] `withCredentials` middleware in router: X-Sa-Credential header → `upstream.WithCredential` ctx (consumed by upstream client's apikey passthrough)
- [x] `/ui/credentials` page (handler + `credentials.templ` + CredentialsView/CredBackend models + schemeLabel) listing backends + auth scheme; sidebar "Credentials" link
- [x] `handleReadError`: ErrUnauthorized → 401 "Credential required" hint; ErrNotFound → 404; else 502 (List/Detail use it)
- [x] credentials.js loaded in layout

## Touched files

- `web/static/js/islands/credentials.js`
- `internal/transport/http/{router.go,handlers/ui.go,views/{layout.templ,credentials.templ,models.go,helpers.go}}`

## Verification

- build/vet/tests pass
- Integration vs gated PocketBase (EXAMPLE_API_KEY): `GET /ui/example/posts` with no credential → **401 + "Credential required" hint**; with `X-Sa-Credential: dev-secret` → admin forwards `X-API-Key` → **200 with row**. Credentials page renders input + scheme label per backend.

## Notes

- Server stores no secret; credential is read from header into ctx per request only.
- bearer/basic/oauth remain stubs (config Validate rejects; passthrough no-op).
