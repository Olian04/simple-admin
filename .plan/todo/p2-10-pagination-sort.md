# Step p2-10 — Pagination + sort UI

Status: `done`

## Sub-tasks

- [x] `resource.ListParams` += `Sort string`, `Desc bool`
- [x] `backend.Client.List`: add `sort` query (PocketBase `-field` for desc) + always page/perPage
- [x] `ListView` += Page/PerPage/TotalItems/TotalPages/Sort/Desc + URL/indicator helpers (SortURL, PageURL, SortIndicator, HasPrev/Next)
- [x] `resource.templ`: sortable column-header links + `Pagination` component (Bulma); HTMX `hx-get` → `#sa-main` swap + `hx-push-url`
- [x] `pageLabel` helper
- [x] List handler: `listParams` (page/perPage default ui.PageSize/sort/desc), compute TotalPages, `isHTMX` → render `views.List` partial vs full Page; `atoiDefault`

## Verification

- generate/build/vet/tests pass
- Integration (30 seeded): page 1 shows pagination nav + "Page 1 / 2"; sort header links carry `sort=title`; `?page=2` with `HX-Request` returns just the `<section>` (no `<html>`)

## Notes

- Sort param uses PocketBase `-field` convention (backend-specific); documented limitation.

## Commit

- (next)
