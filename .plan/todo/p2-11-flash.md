# Step p2-11 — Flash messages

Status: `done`

## Sub-tasks

- [x] `PageData.Flash`; layout renders a Bulma `is-success` notification (`data-sa-flash` + `data-sa-flash-close`)
- [x] `app.js`: delegated flash close + auto-dismiss after 4s (DOMContentLoaded + htmx:afterSwap)
- [x] `flashFromQuery` (created/updated/deleted → message); set on List + Detail full-page renders
- [x] Success redirects carry `?flash=`: create→`?flash=created` (detail), update→`?flash=updated`, delete→`?flash=deleted` (list)

## Verification

- generate/build/vet/tests pass
- Smoke: `GET /ui/gateway/posts?flash=created` renders the `sa-flash` "Created successfully." notification

## Commit

- (next)
