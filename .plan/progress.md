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
| 11 | Credentials + apikey passthrough | todo | | 2026-06-14 |
| 12 | Custom fields + built-in islands | todo | | 2026-06-14 |
| 13 | Polish (CSP, docs, README mermaid) | todo | | 2026-06-14 |

Current focus: **Step 11**.
