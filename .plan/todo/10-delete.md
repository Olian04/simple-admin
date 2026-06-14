# Step 10 — Delete

Status: `done`

## Sub-tasks

- [x] `views/delete.templ` — Bulma `DeleteConfirm` modal (loaded into #sa-modal); confirm = hx-delete; cancel/background = `data-sa-close`
- [x] `web/static/js/app.js` — CSP-safe delegated modal-close handler; included in layout
- [x] handlers `DeleteConfirm` (modal) + `Delete` (Service.Delete → HX-Redirect to list; error → modal with notification); capability-gated
- [x] routes: `GET .../{id}/delete`, `DELETE .../{id}`
- [x] DeleteView model

## Touched files

- `internal/transport/http/{views/delete.templ,handlers/forms.go,router.go,views/{models.go,layout.templ}}`
- `web/static/js/app.js`

## Verification

- build/vet pass
- Integration vs PocketBase: GET delete → modal with hx-delete + data-sa-close; DELETE → 200 + HX-Redirect to list; **record removed upstream** (404 on re-fetch)

## Follow-ups

- Capability gating present (Delete cap → button + route 404 when absent); not separately demoed (example backend supports delete).
- Credentials passthrough = step 11.
