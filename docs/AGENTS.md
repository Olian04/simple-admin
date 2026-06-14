# Agent context: simple-admin

Portable schema-driven admin UI (Go + Templ + HTMX + Bulma) that drives CRUD on
*other* backends. Use with `README.md` + [`configs/config.example.yaml`](../configs/config.example.yaml).

> ⚠️ **ALL AGENTS: caveman communication mode at all times** (terse; drop articles/filler; code & commits normal). **On resume, read [`.plan/README.md`](../.plan/README.md) → `.plan/progress.md` → that step's `.plan/todo/*.md` → `.plan/decisions.md`** — the durable plan/decision/progress record.

## Layout

| Path | Role |
| --- | --- |
| `cmd/simple-admin` | Entry: load config, `logging.Setup`, signal ctx, `app.Run(ctx, cfg, configDir)`. |
| `internal/app` | Composition root: `buildRegistry` (sources → descriptors → Runtimes), `buildWidgets`, wire router. |
| `internal/config` | `backends:` config (manual\|discovery modes), union YAML types, validation. |
| `internal/domain/resource` | Canonical `ResourceDescriptor`, ports (`Store`/`Validator`/`DescriptorSource`), `Service`, swappable `Registry`, JSON-Schema→`[]Field`. **Pure stdlib.** |
| `internal/source/{manual,openapi}` | `DescriptorSource` impls → `[]ResourceDescriptor`. |
| `internal/app/poller.go` | Background discovery poller; re-runs sources on `poll_interval` → `Registry.Replace`. |
| `internal/backend` | CRUD HTTP client (`Store`) + ctx-credential apikey passthrough + typed errors. |
| `internal/schemavalidation` | jsonschema/v6 → `resource.Validator` (field-keyed errors). |
| `internal/views` | Templ page components (layout/list/detail/delete/credentials) + display helpers. |
| `internal/views/form` | Form Templ + `FormView`/`FormFields`/`InputType`, `WidgetRegistry`, `ResolveArgs` (JSONPath), `AutoSelectWidget`, `FieldContext`, widget components. |
| `internal/transport/http` | Router (`GET /`→`/ui`, `/ui/...` routes, favicon), handlers (ui/forms), credentials middleware, CSP. |
| `web/` | `//go:embed static` (Bulma, HTMX, app.css, JS islands). |
| `example-backend/` | **Separate Go module** (own go.mod): PocketBase reference backend + generated OpenAPI (installer disabled). |
| `test/unit/...` | Unit tests mirror `internal/` paths. |

## Dependency direction

`cmd` → `internal/app` → `{source, backend, schemavalidation, views, views/form, transport, config}`; `source/* → domain/resource`; `views/form` may import `views` (not vice-versa); **`domain/resource` imports only stdlib + own ports** (no OpenAPI/HTTP/config). OpenAPI lives under `internal/source/` so domain cannot reach it.

## Key conventions

- Templ components compile to gitignored `*_templ.go`; `make build/lint/test/run` depend on `generate`. Run `go tool templ generate` after editing `.templ`.
- Auth: no server secrets; per-backend creds in browser localStorage → `X-Sa-Credential` → passthrough. Manual auth scheme in YAML; discovery derives from `securitySchemes`. Only none/apikey active.
- Field widgets: built-ins are Templ components in `internal/views/form` (`widget_*` via `widgets.templ`); user `template_file`s are runtime `html/template`. Both get the same `form.FieldContext` (value + Args + `Field` metadata + sibling queries). JS islands (`web/static/js/islands/*.js`) re-mount on `htmx:afterSwap`.
- List responses may be a bare array or `{items:[…]}` envelope (PocketBase). Update verb is per-resource (PUT|PATCH). Sort uses PocketBase `-field` convention.
- Graceful shutdown: `app.Run` returns nil on ctx cancel (exit 0).

## Commands (`make`)

`generate`, `lint`, `test`, `test-race`, `run`, `build` (output `./dist/simple-admin`).

---

## Go proverbs ([source](https://go-proverbs.github.io/), caveman compress)

Concurrency channels coordinate mutex serializes · not parallelism · small interface sharp · zero value useful · `any` untyped tame it · gofmt settles bikeshed · tiny copy beats dep hairball syscall/cgo build tags isolate · cgo not Go · `unsafe` no contract · clarity beats wit · reflection stay cold path · errors values inspect wrap once · architecture name docs users · panic stays in `main` / hard startup.

## Uber style distill ([guide](https://github.com/uber-go/guide/blob/master/style.md), caveman compress)

Rare `*Iface` · `var _ I = (*T)(nil)` at export boundary · defer unlock pairs · chan buffer zero or one usually · slice/map copy exported API boundaries · typed errors `%w` chain handle once · assert comma-ok · goroutine bounded ctx/waitgroup · no zombie `init()` · globals inject not mutate · exits from `main` only · strconv hot paths · structs field-named literals · table tests sub `t.Run`.

---

Canon links above beat bullet memory when tradeoff unclear.
