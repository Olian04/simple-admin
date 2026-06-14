# App-first Go architecture template

Runnable standalone module (copy/adapt into a new repo). Thin `cmd`, `internal/app` wiring, domain + transport + observability; **`labels`** (logs + Prometheus `ConstLabels`) and **`http` / `metrics` / `logging`** blocks typed inside **`internal/config`** (`defaults.go`, `labels.go`, `http.go`, `metrics.go`, `logging.go`, root aggregate).

## Layout

| Path | Role |
| --- | --- |
| `cmd/<binary>` | Entry: env config path, slog setup, signal context, `app.Run`. |
| `internal/app` | Composition root + HTTP + optional metrics listener. |
| `internal/domain/<name>` | Pure behavior (sample: echo). |
| `internal/config` | Root YAML `Load`, aggregate type `Config`, YAML-aligned structs + defaults/`WithDefaults`/`Validate`. |
| `internal/transport/*` | HTTP routes, metrics server helper. |
| `internal/observability/*` | slog helpers, Prometheus registry. |
| `configs/` | Product-shaped example YAML. |
| `test/unit/` | Unit tests mirror package paths under `internal/`. |
| `.cursor/` | Same Cursor layout as sibling Bifrost repo (skills, rules, agents). |
| `.github/workflows/` | CI (lint + race) and tag release via GoReleaser. |

## Dependency direction

`cmd` → `internal/app` → domain, transports, observability, config. Domain avoids HTTP / slog / Prometheus imports.

## Config

Point `APP_CONFIG_FILE` at a YAML file (omit for built-in defaults). See [`configs/config.example.yaml`](configs/config.example.yaml):

- **`labels`** — map attached to every log line and Prometheus const labels on app metrics; keys must satisfy legacy Prometheus label rules.
- **`http.listen_addr`** — main server bind.
- **`metrics.enabled`** / **`listen_addr`** / **`metric_prefix`** — scrape endpoint and metric namespace (`{metric_prefix}_http_requests_total`).
- **`logging`** — `level` (`debug|info|warn|error`), `format` (`json|text`), `stream` (`stdout|stderr`).

## Run locally

```bash
go run ./cmd/echo
APP_CONFIG_FILE=configs/config.example.yaml go run ./cmd/echo
```

```bash
curl -X POST http://localhost:8080/echo -H 'Content-Type: application/json' -d '{"message":" hello "}'
curl http://localhost:9090/metrics
```

Scrape uses Prometheus exposition; `labels` appear as metric const labels where registered.

## Bootstrap new repo from template

Requires git `origin`, infers Go module path, renames cmd default (`echo`), runs `go mod tidy`, deletes itself unless `BOOTSTRAP_KEEP=1`:

```bash
./bootstrap.sh
```

## Checks

Inside this directory:

```bash
make lint
make test
make test-race
```

From enclosing Bifrost repo:

```bash
make test-template
```

## Releases

Push tag `v*`; [`.github/workflows/release.yml`](.github/workflows/release.yml) runs GoReleaser against [`goreleaser/.goreleaser.yaml`](goreleaser/.goreleaser.yaml). [`goreleaser/Dockerfile`](goreleaser/Dockerfile) expects binary artifact `echo` on build context root for manual image builds.

## Agent orientation

[`docs/AGENT_CONTEXT.md`](docs/AGENT_CONTEXT.md): layout, config contract, distilled Go Proverbs + Uber-style guide (compressed).
