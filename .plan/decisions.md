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

---

## 2026-06-14 — Step 5 (descriptor sources)

18. **JSON-Schema → `[]Field` normalizer lives in `domain/resource` (`schema.go`), not split across sources/ui.** Pure `encoding/json`; preserves property declaration order via a token-stream pass. Both `source/manual` and `source/openapi` call it; `ui/render` consumes the resulting `Schema.Fields`. — Avoids duplicating the schema-walk; keeps domain the single owner of schema→field structure.

19. **MVP happy-path = manual mode against PocketBase.** A committed `configs/schemas/post.json` + `update_verb: PATCH` drives the real end-to-end (steps 8–11). Generic OpenAPI resource-discovery (`source/openapi`) is the secondary path, tested against a spec fixture. — Generic discovery from arbitrary specs is brittle; don't gate the MVP on it.

20. **Manual mode capabilities default to full CRUD**, with `update_verb` (PUT|PATCH, default PUT) configurable per resource. — Manual config has no per-verb info; full CRUD is the sensible default and the verb override handles PATCH-only backends like PocketBase.

21. **`source/openapi` uses a focused `encoding/json` parser, NOT `pb33f/libopenapi` (for now).** Supersedes decision 3's lib choice for the MVP. Our reference spec is JSON with local component `$ref`s; a small parser (resolve `#/components/schemas/*`, pair collection/item paths, derive capabilities + auth scheme from `securitySchemes`) is robust and keeps the module lean. — Revisit libopenapi if we need YAML specs, OpenAPI 3.1 nuances, or remote `$ref`s. Discovery polling is still deferred; this is the one-shot parse.

---

## 2026-06-14 — Phase 2 (cleanup/refactor + features)

22. **`.plan/` is a STANDING REPO RULE** for every plan/phase, not just Phase 1 — recorded in `.plan/README.md`.

23. **Routing keeps the `/ui` prefix**; add `GET /` → 302 `/ui` (+ `GET /favicon.ico` → 204). — User chose this over dropping the prefix (resolves the #4-vs-#7 conflict in their request).

24. **Remove the `echo` scaffold; rename product `echo` → `simple-admin`** (cli/binary/logs/`DefaultMetricPrefix`). — Kills the biggest "service" naming confusion; `echo` was leftover template.

25. **Renames:** `internal/upstream` → `internal/backend` (`backend.Client`, `backend.Error`); `internal/validation` → `internal/schemavalidation`. **Keep `internal/source`** (descriptor origin, genuinely distinct). Rename widget `render.Registry` → `form.WidgetRegistry`.

26. **Package layout:** delete `internal/ui/`; views move to top-level `internal/views` with a `internal/views/form` subpackage (form rendering: form.templ + fields + WidgetRegistry + args + autoselect + built-in widget templ). `form` may import `views`; `views` must not import `form`.

27. **Built-in widgets become Templ components** (compile-time); user-supplied `template_file`s stay runtime `html/template` (can't compile user files). `FormFieldView.Widget templ.Component`.

28. **Widget rich-editor JS = CDN by default, config-overridable** via a new `ui:` config section (`MonacoCDN`, `EasyMDECDN`, `PageSize`). Airgap can self-host. — Avoids vendoring multi-MB Monaco / Node build.

29. **Phase-2 feature scope:** discovery polling, pagination/sort, flash, widgets (code/markdown/tags/color/json). **Staying stubbed:** low-value widgets + bearer/basic/oauth auth.

30. **CI/govulncheck:** bump Go toolchain to latest 1.26.x; if stdlib-only advisories persist (DB ahead of releases), make the govulncheck step report-only with a documented note rather than a red gate.

31. **Widget field context standardized + enriched (`form.FieldContext`).** ONE value type passed to every widget — built-in Templ components AND user `html/template` files get the same data + their Args. Carries the field's full `resource.Field` metadata (`Field`) plus a queryable `Siblings` view (`Sibling`/`HasSibling`/`SiblingValue`/`SiblingsOfKind`/`BoolSiblings`/`IDSiblings`). Additive — existing flat fields (Name/Value/Arg/…) kept. — Lets a widget work on its own data and look up related fields (bool flags, ids) regardless of widget type.

32. **PocketBase installer popup disabled** in example-backend (`ServeEvent.InstallerFunc = nil`); superuser via CLI if the dashboard is needed. — No browser setup page on `serve`.


