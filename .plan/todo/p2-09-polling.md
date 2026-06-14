# Step p2-9 — Discovery polling loop

Status: `done`

## Sub-tasks

- [x] Extract `assembleRuntimes(ctx,cfg,configDir)` from `buildRegistry` (sources→descriptors→validator+backend.Client+Service→[]*Runtime)
- [x] `internal/app/poller.go`: `startDiscoveryPoller` (ticker on smallest discovery `poll_interval`; each tick re-assemble + `registry.Replace`; keep current set on error; stop on ctx) + `discoveryInterval`
- [x] Wire `startDiscoveryPoller` in `bootstrap.go` (no-op when no discovery backend)

## Verification

- build/vet pass
- Integration: gateway discovery backend with `poll_interval: 2s` → admin logs `discovery refreshed` repeatedly (2× in 5s); atomic `Registry.Replace` keeps readers live

## Commit

- (next)
