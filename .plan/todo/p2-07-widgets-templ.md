# Step p2-7 — Built-in widgets → Templ

Status: `done`

## Sub-tasks

- [x] `internal/views/form/widgets.templ`: built-in widgets as Templ components — `widgetDatetime`, `widgetRelation` (real) + `widgetStubInput`/`widgetStubTextarea` (back the 21 not-yet-impl ids). Param named `fc` to avoid shadowing templ's implicit `ctx context.Context`
- [x] Rewrite `widgets.go`: `WidgetRegistry{user}` + `builtinWidgets map[string]func(FieldContext) templ.Component`; `Component(id, ctx) (templ.Component, bool)` (built-in → component; user file → html/template → `templ.Raw`); `Has`; kept `FieldContext`/`Arg`
- [x] Delete `internal/views/form/templates/*.html.tmpl` + embed
- [x] `FormFieldView.CustomHTML template.HTML` → `Widget templ.Component`; `form.templ` dispatch `if f.Widget != nil { @f.Widget }`
- [x] handler `renderWidget` returns `templ.Component` via `widgets.Component`; sets `fv.Widget`

## Verification

- generate/build/vet/tests pass
- Integration: discovery posts edit → datetime input + relation select render via templ; manual posts edit → `code` stub (textarea + "not yet implemented"); user `template_file` path still works (html/template → templ.Raw)

## Commit

- (next)
