# Agent context: simple-admin

Portable schema-driven admin UI (Go + [webui](https://github.com/Olian04/webui)) that drives CRUD on
*other* backends. Use with `README.md` + [`configs/config.example.yaml`](../configs/config.example.yaml).

> ⚠️ **ALL AGENTS: caveman communication mode at all times** (terse; drop articles/filler; code & commits normal). **On resume, read [`.plan/README.md`](../.plan/README.md) → `.plan/progress.md` → that step's `.plan/todo/*.md` → `.plan/decisions.md`** — the durable plan/decision/progress record.

## Layout

| Path | Role |
| --- | --- |
| `cmd/simple-admin` | Entry: load config, `logging.Setup`, signal ctx, `app.Run(ctx, cfg, configDir)`. |
| `internal/app` | Composition root: `buildRegistry` (sources → descriptors → Runtimes), `ui.NewLive`, wire router; poller triggers `Live.Rebuild`. |
| `internal/config` | `backends:` config (manual\|discovery modes), union YAML types, validation. |
| `internal/domain/resource` | Canonical `ResourceDescriptor`, ports (`Store`/`Validator`/`DescriptorSource`), `Service`, swappable `Registry`, JSON-Schema→`[]Field`. **Pure stdlib.** |
| `internal/source/{manual,openapi}` | `DescriptorSource` impls → `[]ResourceDescriptor`. |
| `internal/app/poller.go` | Background discovery poller; re-runs sources on `poll_interval` → `Registry.Replace`. |
| `internal/backend` | CRUD HTTP client (`Store`) + ctx-credential apikey passthrough + typed errors. |
| `internal/schemavalidation` | jsonschema/v6 → `resource.Validator` (field-keyed errors). |
| `internal/ui` | webui declaration from the `Registry`: `accessors.go` (schema field → `webui.Accessor[Record]`), `pages.go` (list/detail/new/edit; delete is a confirmed toolbar action), `credentials.go`, `session.go` (in-memory per-visitor credentials), `ui.go` (`Build`, `Live` atomic swap). |
| `internal/transport/http` | Router (`/healthz`, `GET /`→`/ui/`, UI under `/ui/`), request counting/logging. |
| `example-backend/` | **Separate Go module** (own go.mod): PocketBase reference backend + generated OpenAPI (installer disabled). |
| `test/unit/...` | Unit tests mirror `internal/` paths. |

## Dependency direction

`cmd` → `internal/app` → `{source, backend, schemavalidation, ui, transport, config}`; `source/* → domain/resource`; `ui → {domain/resource, backend}`; **`domain/resource` imports only stdlib + own ports** (no OpenAPI/HTTP/config). OpenAPI lives under `internal/source/` so domain cannot reach it.

## Key conventions

- UI is a webui declaration (needs Go 1.27). No HTML/CSS/JS here; if webui cannot do something, record it in `docs/webui-gaps.md` instead of working around it with markup or script.
- Auth: no stored secrets; per-visitor creds in server memory (`ui.Sessions`, `sa_session` cookie) → ctx credential → passthrough. Manual auth scheme in YAML; discovery derives from `securitySchemes`. Only none/apikey active.
- Inputs: `ui.accessor` maps a `resource.Field` to a webui accessor; config `field_hints` (`relation|slider|select`) override it. Relation options come from `Select.LoadOptions` (first 200 records, 10 s cache). Writes send only writable fields; schema errors become `webui.Reject` beside the field.
- Registry changes (discovery) → `Live.Rebuild` recompiles the whole app and swaps it atomically; a failed compile keeps the old UI. webui counts rebuilds into the same `metrics.Registry.Registerer()`.
- List responses may be a bare array or `{items:[…]}` envelope (PocketBase). Update verb is per-resource (PUT|PATCH). Sort uses PocketBase `-field` convention.
- Graceful shutdown: `app.Run` returns nil on ctx cancel (exit 0).

## Commands (`make`)

`lint`, `test`, `test-race`, `run`, `build` (output `./dist/simple-admin`).

---

## Go proverbs ([source](https://go-proverbs.github.io/), caveman compress)

Concurrency channels coordinate mutex serializes · not parallelism · small interface sharp · zero value useful · `any` untyped tame it · gofmt settles bikeshed · tiny copy beats dep hairball syscall/cgo build tags isolate · cgo not Go · `unsafe` no contract · clarity beats wit · reflection stay cold path · errors values inspect wrap once · architecture name docs users · panic stays in `main` / hard startup.

## Uber style distill ([guide](https://github.com/uber-go/guide/blob/master/style.md), caveman compress)

Rare `*Iface` · `var _ I = (*T)(nil)` at export boundary · defer unlock pairs · chan buffer zero or one usually · slice/map copy exported API boundaries · typed errors `%w` chain handle once · assert comma-ok · goroutine bounded ctx/waitgroup · no zombie `init()` · globals inject not mutate · exits from `main` only · strconv hot paths · structs field-named literals · table tests sub `t.Run`.

---

Canon links above beat bullet memory when tradeoff unclear.
