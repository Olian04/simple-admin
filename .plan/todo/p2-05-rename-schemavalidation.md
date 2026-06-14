# Step p2-5 — Rename internal/validation → internal/schemavalidation

Status: `done`

## Sub-tasks

- [x] `git mv internal/validation internal/schemavalidation` + `test/unit/validation` → `test/unit/schemavalidation`
- [x] `package validation` → `package schemavalidation`; test qualifiers updated
- [x] `app/resources.go`: import + `schemavalidation.New`

## Verification

- build + schemavalidation test pass

## Commit

- (next)
