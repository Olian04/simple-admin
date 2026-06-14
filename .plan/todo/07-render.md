# Step 7 — Render engine (schema → UI model)

Status: `done`

## Sub-tasks

- [x] `internal/ui/render/render.go`: `Columns(fields)` (scalars/enums, omit array/object), `FormFields(fields,data,errs)` (exclude readOnly, seed values+errors), `Cell`/`DisplayValue` (bool→✓/✗, number, string), `ValueString`, `InputType` (kind/format → HTML input type)
- [x] tests `test/unit/ui/render/render_test.go` (columns exclude array/object, form excludes readonly + seeds, display values, input types)

## Touched files

- `internal/ui/render/render.go`, `test/unit/ui/render/render_test.go`

## Verification

- `go build ./...` + `go test ./test/unit/ui/...` pass

## Follow-ups

- list/detail templ (step 8) consume `Columns`/`Cell`; form templ (step 9) consumes `FormFields`/`InputType`.
- Auto-select built-in widgets + JSONPath arg resolution = step 12 (extends this package: registry.go + args.go).
