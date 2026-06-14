# Step p2-4 — Rename internal/upstream → internal/backend

Status: `done`

## Sub-tasks

- [x] `git mv internal/upstream internal/backend`; `package upstream` → `package backend` (3 files)
- [x] `UpstreamError` → `Error`; error message prose "upstream"→"backend"; keep `Client`, `ErrNotFound`, `ErrUnauthorized`, `WithCredential`/`CredentialFromContext`
- [x] Update importers: `app/resources.go` (`backend.New`), `router.go` (`backend.WithCredential`), `handlers/ui.go` + `handlers/forms.go` (`backend.ErrNotFound`/`ErrUnauthorized`)

## Verification

- build/vet/tests pass

## Commit

- (next)
