# Step p2-2 — Root + favicon routes

Status: `done`

## Sub-tasks

- [x] `router.go`: `GET /{$}` → 302 `/ui` (exact-root anchor, doesn't swallow other routes)
- [x] `GET /favicon.ico` → 204

## Verification

- build pass; smoke: `GET /` → 302 Location `/ui`; `/favicon.ico` → 204

## Commit

- (next)
