# Step p2-6 — Restructure → internal/views + internal/views/form

Status: `done`

## Sub-tasks

- [x] `git mv internal/transport/http/views internal/views`
- [x] Create `internal/views/form` (package `form`): moved `form.templ`, `widgets.go`, `args.go`, `autoselect.go`, `templates/`
- [x] `internal/views/display.go` (package views): `Columns`/`Cell`/`DisplayValue`/`ValueString` (uses existing `views.Column`, dropped duplicate render.Column)
- [x] `internal/views/form/fields.go`: `FormField`/`FormFields`/`InputType` + moved `FormView`/`FormFieldView` out of views/models.go
- [x] `render.Registry` → `form.WidgetRegistry`; `NewRegistry` → `NewWidgetRegistry`; package `render` → `form` in moved files
- [x] Delete `internal/ui/` + old `internal/transport/http/views`
- [x] Fix importers: handlers (ui.go/forms.go), router.go, app/resources.go — views path + render→{views,form} qualifiers; UIHandler.widgets → *form.WidgetRegistry
- [x] Move tests → `test/unit/views` + `test/unit/views/form` (delete `test/unit/ui`)
- [x] Fixed stale "Validator wired in a later step" comment

## Verification

- templ generate + build/vet/tests/golangci (0 issues) pass; `/ui` smoke renders

## Notes

- `form` does not import `views` (no coupling); handlers import both. `views` never imports `form`.
- Built-in widgets still html/template here; templ conversion is p2-7.

## Commit

- (next)
