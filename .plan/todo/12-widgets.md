# Step 12 — Custom fields + built-in islands

Status: `done`

## Sub-tasks

- [x] `internal/ui/render/widgets.go`: `Registry` (embedded built-ins + user template_files keyed by config path, resolved vs baseDir), `FieldContext` (+`Arg` helper), `Render` (@templ.Raw'd HTML; html/template escapes record data)
- [x] `args.go`: `ResolveArgs` (static | `{jsonpath}` via ohler55/ojg)
- [x] `autoselect.go`: `AutoSelectWidget` (date/time→datetime; `{resource}`/`{resource}Id`→relation), `relationTarget`, `DatetimeInputType`
- [x] built-in templates: `datetime`, `relation` (real) + 21 stubs (code/geo/markdown/richtext/json textareas; file/image/color/tags/multiselect/duration/rating/slug/currency/percent/phone/email/url/password/country/timezone inputs) — BOM-free
- [x] `relation.js` island (fetch options + inject credential for gated backends); loaded in layout
- [x] handler `Options` JSON endpoint + route `GET /ui/{backend}/{name}/options`; `firstStringField` label
- [x] `buildFormView` → method; `renderWidget` (explicit config or auto-select → HTML); `FormFieldView.CustomHTML` + form.templ `@templ.Raw` dispatch
- [x] wire widget Registry through app(`buildWidgets`)→bootstrap→router→UIHandler
- [x] tests `test/unit/ui/render/widgets_test.go` (auto-select datetime/relation/none, ResolveArgs static+jsonpath, registry built-ins + render)

## Touched files

- `internal/ui/render/{widgets,args,autoselect}.go` + `templates/*.html.tmpl`
- `internal/transport/http/{handlers/{ui,forms}.go,views/{models.go,form.templ,layout.templ},router.go}`
- `internal/app/{resources.go,bootstrap.go}`, `web/static/js/islands/relation.js`
- `test/unit/ui/render/widgets_test.go`

## Verification

- build/vet/tests pass
- Integration vs PocketBase: manual posts edit → `code` textarea (explicit, body/status), `rating` stub, `datetime` (publishedAt); discovery posts edit → `relation` select for author (auto-selected) with `/ui/gateway/authors/options`; options endpoint returns `[{value,label}]` (author name resolved). JSONPath arg (`rating.max`) wired.

## Follow-ups (step 13 polish)

- CSP header (islands already external files; inline `required`/attrs fine).
- README + mermaid, pagination/sort, flash messages.
