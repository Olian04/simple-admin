# Decision log

Append-only. Newest at bottom. Each entry: decision · rationale · date.

---

## 2026-06-14 — Initial planning session (all locked with user)

1. **Frontend stack = Templ + HTMX + JS islands.** Server-rendered Go (Templ compiled components) for shell + most CRUD UI; HTMX for partial swaps; small JS islands only for rich widgets (e.g. code editor, maps) and the credential store. No SPA, no Node toolchain. — Keeps stack Go-oriented; avoids a JS build pipeline.

2. **Resource description decoupled via canonical `ResourceDescriptor`.** Domain never imports OpenAPI/HTTP types. Sources (`manual`, `openapi`) produce descriptors. — Makes discovery pluggable; keeps domain pure.

3. **Discovery format = OpenAPI 3.** Parsed with `pb33f/libopenapi`. — Standard, many backends emit it; lib handles 3.1 + 2020-12 + `$ref`.

4. **CSS = Bulma** (vendored + embedded, single file, no JS/no build). — Rich class-based components (menu/card/table/form); Chakra UI & Material UI rejected (React-only, incompatible with Templ/HTMX SSR); Tailwind+DaisyUI rejected (needs a build step).

5. **Example/reference backend = SEPARATE Go module built on PocketBase.** `example-backend/` own `go.mod`, own Dockerfile, run via `deploy/docker-compose.yml`. Adds a generated `GET /openapi.json` from PocketBase collection schemas. — PocketBase's big dep tree must not contaminate the admin UI module (which stays backend-agnostic); PocketBase gives DB+CRUD+auth for free and becomes the future bearer-auth test bed.

6. **Auth = browser-stored credentials + passthrough; no server-side secrets.** Per-backend creds in browser localStorage keyed by stable `backendId`; injected via HTMX `htmx:configRequest`; server extracts + passes through to upstream. Implement `none` + `apikey`; `bearer/basic/oauth` stubbed (config Validate rejects). Auth **scheme** declared in YAML for manual backends, **derived from spec securitySchemes** for discovery backends. Discovery-request creds are the one server-side secret (env-expanded). — Multi-backend in parallel; no user store.

7. **Stable `backendId`.** Explicit YAML `id` preferred; else deterministic `sha256(base_url)` prefix; never random. — localStorage cred key must survive restarts.

8. **Config = one `backends:` list, two mutually-exclusive modes per backend** (`manual` with own JSON Schema, or `discovery` with a spec URL). OpenAPI only consumed by discovery. — Simpler mental model than separate sections.

9. **Field templates = descriptive library-agnostic ids + optional args + convention auto-select.** ids name the widget function (`code`, `geo`, `datetime`, `relation`), not the library. Args optional with best-effort defaults; values static or `{ jsonpath: '$.x' }` against the record. Renderer auto-assigns built-ins by schema/name convention; explicit config overrides. MVP implements `datetime` + `relation`; rest stubbed. — Empty config still yields rich UI.

10. **Repo layout discipline.** All admin-UI Go code under `internal/`; only `cmd/`, `test/`, and the separate `example-backend/` module live outside. Keeps existing `cmd/ internal/ test/` structure.

11. **`.plan/` resume contract** committed in-repo. — Durable cross-session progress tracking.

12. **Caveman communication mode** standing directive for all agents on this project.

13. **JSONPath lib = `ohler55/ojg` (`jp`)** for resolving template args from the record. — Standard JSONPath, fast, no cgo.

14. **JSON Schema validation = `santhosh-tekuri/jsonschema/v6`** (already transitive; promote to direct). — Spec-compliant, structured pointer errors map to form fields.

---

## 2026-06-14 — Step 2 (reference backend)

15. **List responses may be a pagination envelope `{items:[…], totalItems, …}`, not a bare array.** PocketBase (the reference backend) returns this shape. — Contract relaxed: admin upstream List client auto-detects bare-array vs `items` envelope. (PLAN contract bullet "list returns array" → "list returns array or `{items:[]}` envelope".)

16. **Update verb is PATCH for PocketBase** (no PUT). — Admin must honor the spec's actual update verb per resource (PUT or PATCH from capabilities), not assume PUT.

17. **PocketBase pinned v0.39.3**, CGO-free (modernc sqlite) → static Docker builds with `CGO_ENABLED=0`.

