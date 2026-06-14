# Progress board

Status: `todo` · `in_progress` · `blocked` · `done`. Update date on every change.

| Step | Title | Status | Note | Updated |
|---|---|---|---|---|
| 0 | Set up `.plan/` dir | done | README/PLAN/decisions/progress/todo created + committed | 2026-06-14 |
| 1 | Tooling (templ, deps, CI) | done | templ tool v0.3.1020 + Makefile/gitignore/CI; runtime deps deferred to consumers | 2026-06-14 |
| 2 | Reference backend (separate PocketBase module) | done | PocketBase v0.39.3 module + generated /openapi.json + Dockerfiles + compose; smoke-tested | 2026-06-14 |
| 3 | Config (backends, manual\|discovery) | done | resources.go union types + validation + env-expand + example + tests | 2026-06-14 |
| 4 | Domain canonical model | done | descriptor/ports/service/registry, pure stdlib, tests | 2026-06-14 |
| 5 | Descriptor sources (manual + openapi) | done | domain schema normalizer + manual + openapi(json) sources + tests | 2026-06-14 |
| 6 | Static assets + shell (Bulma/Templ) | done | embed assets + templ shell + sidebar from registry; integration smoke green | 2026-06-14 |
| 7 | Render engine (schema → UI model) | done | render.go columns/form/display/input + tests | 2026-06-14 |
| 8 | List + detail + upstream read | done | upstream client (Store) + list/detail templ + routes; integration verified vs PocketBase | 2026-06-14 |
| 9 | Forms + write + validation | done | validator + form templ/handlers + 422 re-render; create/update verified vs PocketBase | 2026-06-14 |
| 10 | Delete | done | confirm modal + DELETE route + CSP-safe modal-close js; verified vs PocketBase | 2026-06-14 |
| 11 | Credentials + apikey passthrough | done | credentials.js + middleware + page; passthrough verified vs gated PocketBase | 2026-06-14 |
| 12 | Custom fields + built-in islands | done | widget registry + args(jsonpath) + auto-select + relation island/options + stubs; verified | 2026-06-14 |
| 13 | Polish (CSP, docs, README mermaid) | done | CSP+headers, README+mermaid+contract, AGENTS.md; lint clean (govulncheck=stdlib/toolchain, noted) | 2026-06-14 |

Phase 1 (MVP) complete (steps 0–13 done).

## Phase 2 — cleanup/refactor + high-value features

| Step | Title | Status | Note | Updated |
|---|---|---|---|---|
| p2-0 | `.plan` Phase-2 setup + standing rule | done | 3e64e3f | 2026-06-14 |
| p2-1 | Graceful shutdown (exit 0 on Ctrl+C) | done | ctx.Done→nil + main ignores Canceled + test | 2026-06-14 |
| p2-2 | GET / → /ui + favicon 204 | done | GET /{$}→302 /ui, favicon 204; smoke ok | 2026-06-14 |
| p2-3 | Remove echo scaffold + rename product → simple-admin | done | echo deleted; product renamed (cli/binary/logs/prefix/goreleaser/release) | 2026-06-14 |
| p2-4 | Rename internal/upstream → internal/backend | done | pkg moved; UpstreamError→Error; importers updated | 2026-06-14 |
| p2-5 | Rename internal/validation → internal/schemavalidation | done | pkg+test moved; importer updated | 2026-06-14 |
| p2-6 | Restructure → internal/views + internal/views/form | done | ui/ deleted; views top-level + views/form; WidgetRegistry; lint clean | 2026-06-14 |
| p2-7 | Built-in widgets → templ | done | widgets.templ components + WidgetRegistry dispatch; html.tmpl deleted; verified | 2026-06-14 |
| p2-8 | Config ui: section (CDN URLs, page size) | done | UISection + data attrs to layout; threaded handler/router/bootstrap | 2026-06-14 |
| p2-9 | Discovery polling loop | done | assembleRuntimes + poller ticker + Registry.Replace; 2s smoke fires | 2026-06-14 |
| p2-10 | Pagination + sort UI | done | ListParams sort + pagination/sortable headers + HTMX partial; verified | 2026-06-14 |
| p2-11 | Flash messages | in_progress | | 2026-06-14 |
| p2-12 | High-value widgets (color/json/tags/markdown/code) | todo | | 2026-06-14 |
| p2-13 | Comments cleanup + docs + CI + .plan | todo | | 2026-06-14 |

Current focus: **Phase 2, Step p2-0**. Staying stubbed: low-value widgets + bearer/basic/oauth auth. Routing keeps `/ui` prefix (+ `GET /` redirect). `internal/source` kept.
