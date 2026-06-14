# Step p2-12b — Standardize widget field context

Status: `done`

## Sub-tasks

- [x] Move `FieldContext` to `internal/views/form/context.go`; enrich with `Field resource.Field` (per-field schema metadata) + `Siblings []Sibling`
- [x] Sibling query API (template-friendly): `Sibling`/`HasSibling`/`SiblingValue`/`SiblingsOfKind`/`BoolSiblings`/`IDSiblings`
- [x] Handler `renderWidget` populates `Field` + `Siblings` (via `siblingsOf(ff, name)`); same context for built-in templ AND user html/template widgets
- [x] Unit test for sibling queries

## Verification

- generate/build/vet/golangci(0)/tests pass; sibling-query test green
- `TestUserTemplateGetsEnrichedContext`: a user html/template renders `.Field.Kind`/`.Arg`/`.SiblingValue`/`.BoolSiblings`/`.IDSiblings` — proves built-in templ AND user templates share the same enriched context

## Notes

- Additive: existing flat fields (Name/Label/Value/Args/OptionsURL/Arg) retained, so existing widgets/templates unchanged.
- Contract documented in `context.go` doc comment; README/AGENTS note added in p2-13.

## Commit

- (next)
