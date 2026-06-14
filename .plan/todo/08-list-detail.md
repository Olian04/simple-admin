# Step 8 — List + detail + upstream read

Status: `done`

## Sub-tasks

- [x] `internal/upstream/`: `client.go` (Store impl: List/Get/Create/Update/Delete; List handles bare-array OR `{items:[]}` envelope; update verb from caps), `auth.go` (ctx credential + apikey passthrough header/query), `errors.go` (ErrNotFound/ErrUnauthorized/UpstreamError)
- [x] `app/resources.go`: build `upstream.New(d)` Store + `Service` per descriptor → live Runtimes
- [x] views: `resource.templ` (List table, Detail pairs, Error) + `helpers.go` (uiURL/uiPath/colspan) + ListView/DetailView/Column/ListRow/DetailPair/ErrorView models
- [x] handlers: `List`, `Detail`, `Index` (redirect to first resource), `lookup`/`renderError`/`renderStatus`, buildListView/buildDetailView
- [x] routes: `GET /ui/{backend}/{name}`, `GET /ui/{backend}/{name}/{id}`

## Touched files

- `internal/upstream/{client,auth,errors}.go`
- `internal/app/resources.go`, `internal/transport/http/{router.go,handlers/ui.go,views/{models.go,helpers.go,resource.templ}}`

## Verification

- build/vet/tests pass
- Integration: seeded posts in PocketBase; `GET /ui/example/posts` (manual) shows table with rows + New button; `GET /ui/example/posts/{id}` shows fields + Edit/Delete; `GET /ui/gateway/posts` (discovery) also renders. Upstream read works through real backend.

## Follow-ups

- Edit/Delete buttons present; their routes land in steps 9 (forms/update) and 10 (delete).
- Auth passthrough plumbed (ctx credential); middleware that sets it = step 11.
- List pagination params plumbed but UI controls deferred to polish (step 13).
