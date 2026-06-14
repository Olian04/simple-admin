# Step 5 — Descriptor sources (manual + openapi)

Status: `done`

## Sub-tasks

- [x] `internal/domain/resource/schema.go`: `FieldsFromJSONSchema` — order-preserving JSON-Schema→[]Field (token-stream property order, type/format/enum/min/max/maxLength/pattern/items/readOnly), `titleizeField`
- [x] `internal/source/manual/manual.go`: manual backends → descriptors; load schema from disk path or `url`; full-CRUD caps + `update_verb`; auth scheme from config; apply field templates (incl. arg conversion)
- [x] `internal/source/openapi/openapi.go`: focused JSON parser — fetch (with discovery apikey creds) + `Parse` (pure): pair collection/item paths, resolve component schema via `$ref`, capabilities + update verb from verbs present, auth scheme from `securitySchemes`
- [x] config: `update_verb` (PUT|PATCH, default PUT) on ResourceSection; example posts → PATCH
- [x] committed `configs/schemas/post.json` (manual MVP path schema)
- [x] tests: `test/unit/source/manual` (descriptor build, field order, required, enum, custom template), `test/unit/source/openapi` (Parse fixture → caps/auth/paths/fields)

## Touched files

- `internal/domain/resource/schema.go`
- `internal/source/manual/manual.go`, `internal/source/openapi/openapi.go`
- `internal/config/{resources.go,defaults.go}`, `configs/config.example.yaml`, `configs/schemas/post.json`
- `test/unit/source/{manual,openapi}/*_test.go`

## Verification

- `go build ./...`, `go vet ./...` exit 0; all unit tests pass

## Decisions (see decisions.md)

- 18: normalizer in domain; 19: MVP happy-path = manual mode; 20: manual caps full CRUD + update_verb; 21: openapi source uses encoding/json (not libopenapi) for MVP.

## Follow-ups

- `app/resources.go` (step 8) runs both sources, builds Stores/Validators/Runtimes, `Registry.Replace`.
- openapi source: relation field names come through as-is (e.g. `author`); fine.
