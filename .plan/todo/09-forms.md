# Step 9 — Forms + write + validation

Status: `done`

## Sub-tasks

- [x] `internal/validation/validator.go`: santhosh-tekuri/jsonschema v6 wrapper → `resource.Validator`; maps BasicOutput errors to field→message (incl. root-level "missing properties" → required)
- [x] Wire Validator into each Service (fail fast on bad schema) in `app/resources.go`
- [x] `views/form.templ`: `Form` (full) + `FormBody` (partial); inputs by kind (text/number/checkbox/select/textarea), inline errors, hx-post/hx-put → `#sa-form` outerHTML swap
- [x] `views.FormView`/`FormFieldView` models
- [x] handlers `forms.go`: NewForm/EditForm (full page), Create/Update (parse → validate → upstream), `parseForm` (type coercion per kind), 422 FormBody re-render with errors, HX-Redirect on success, capability gating
- [x] routes: `GET .../new`, `POST .../`, `GET .../{id}/edit`, `PUT .../{id}`
- [x] tests `test/unit/validation` (pass, missing-required, maxLength+enum field mapping)

## Touched files

- `internal/validation/validator.go`, `internal/app/resources.go`
- `internal/transport/http/{router.go,handlers/forms.go,views/{models.go,form.templ}}`
- `test/unit/validation/validator_test.go`

## Verification

- build/vet/tests pass
- Integration vs PocketBase: new form renders (title/status/etc); valid create → 200 + `HX-Redirect` to new record; invalid create (no title) → 422 with inline "is required"; edit form prefilled; PUT update → 200 + redirect; **persisted upstream via PATCH** (verified record title/status/rating changed)

## Follow-ups

- Credential passthrough middleware (step 11) sets ctx credential the upstream client already consumes.
- Delete (step 10) adds confirm modal + DELETE route.
- Custom field widgets / islands (step 12) extend FormBody input dispatch.
