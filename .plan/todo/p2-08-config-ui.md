# Step p2-8 — Config ui: section

Status: `done`

## Sub-tasks

- [x] `internal/config/ui.go`: `UISection{MonacoCDN, EasyMDECDN string; PageSize int}` + WithDefaults + Validate
- [x] `defaults.go`: DefaultMonacoCDN / DefaultEasyMDECDN / DefaultPageSize (25)
- [x] Aggregate `UI` into root `Config` (WithDefaults + Validate)
- [x] Thread to UI: `PageData.MonacoCDN/EasyMDECDN`; layout renders `<body data-monaco-cdn data-easymde-cdn>` (CSP-safe; islands read dataset); `UIHandler.ui` + NewUIHandler param; Router + bootstrap pass `cfg.UI`
- [x] `configs/config.example.yaml`: `ui:` block

## Verification

- generate/build/vet/tests pass (example-config-loads guard still green)

## Notes

- `PageSize` consumed by pagination (p2-10); CDN bases consumed by code/markdown islands (p2-12).

## Commit

- (next)
