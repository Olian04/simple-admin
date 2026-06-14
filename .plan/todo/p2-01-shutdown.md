# Step p2-1 — Graceful shutdown

Status: `done`

## Sub-tasks

- [x] `bootstrap.go`: `ctx.Done()` case returns `nil` (logs "shutting down"), not `ctx.Err()`
- [x] `main.go`: ignore `context.Canceled` from `root.Run` (no error log, exit 0); add `errors` import
- [x] test: `test/unit/app/shutdown_test.go` — canceled ctx → `Run` returns nil

## Touched files

- `internal/app/bootstrap.go`, `cmd/simple-admin/main.go`, `test/unit/app/shutdown_test.go`

## Verification

- build + test pass; manual Ctrl+C → exit 0 (verified in p2 smoke later)

## Commit

- (next)
