# Agent context: app template

App-first HTTP echo service template — use with repo `README.md` and [`configs/config.example.yaml`](../configs/config.example.yaml).

## Layout

| Path | Role |
| --- | --- |
| `cmd/echo` | Entry: load config, `logging.Setup`, signal context, `app.Run`. |
| `internal/app` | Wire Prometheus registry, router, dual listeners. |
| `internal/config` | Root YAML `Load`; `Config` + `Labels`, `HTTPSection`, `MetricsSection`, `LoggingSection` (defaults/`WithDefaults`/`Validate`). |
| `internal/domain/echo` | Echo domain logic only. |
| `internal/transport/http` | Routes + middleware (request logs + metrics increment). |
| `internal/transport/metricshttp` | `/metrics` server lifecycle. |
| `internal/observability/logging` | slog setup from config. |
| `internal/observability/metrics` | Prometheus counter + handler. |
| `test/unit/...` | Unit tests beside mirrored paths. |
| `configs` | YAML contract examples. |

## Dependency direction

`cmd` → `internal/app` → domain, transport, observability, aggregated `internal/config`. Domain must not import app, transports, observability.

## Configuration

`APP_CONFIG_FILE` optional; unset uses defaults matching example file shapes.

Section structs (`Labels`, HTTP, metrics, logging maps) carry `Default*` constants in-package, `WithDefaults()`, `Validate()`. Root `config.Config.Validate()` delegates down.

Prometheus **`labels`** keys: legacy Prometheus label ASCII rules (`model.LegacyValidation`). **`metrics.metric_prefix`**: composed `{prefix}_http_requests_total` must be valid legacy metric name.

## Commands (`make`)

`lint`, `test`, `test-race`, `run`, `build` (output `./dist/echo`).

---

## Go proverbs ([source](https://go-proverbs.github.io/), caveman compress)

Concurrency channels coordinate mutex serializes · not parallelism · small interface sharp · zero value useful · `any` untyped tame it · gofmt settles bikeshed · tiny copy beats dep hairball syscall/cgo build tags isolate · cgo not Go · `unsafe` no contract · clarity beats wit · reflection stay cold path · errors values inspect wrap once · architecture name docs users · panic stays in `main` / hard startup.

## Uber style distill ([guide](https://github.com/uber-go/guide/blob/master/style.md), caveman compress)

Rare `*Iface` · `var _ I = (*T)(nil)` at export boundary · defer unlock pairs · chan buffer zero or one usually · slice/map copy exported API boundaries · typed errors `%w` chain handle once · assert comma-ok · goroutine bounded ctx/waitgroup · no zombie `init()` · globals inject not mutate · exits from `main` only · strconv hot paths · structs field-named literals · table tests sub `t.Run`.

---

Canon links above beat bullet memory when tradeoff unclear.
