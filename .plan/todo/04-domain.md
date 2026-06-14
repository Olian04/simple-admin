# Step 4 — Domain canonical model

Status: `done`

## Sub-tasks

- [x] `internal/domain/resource/descriptor.go`: `ResourceDescriptor`, `Backend`, `Schema`, `Field`/`FieldKind`, `Capabilities` (incl. `UpdateVerb`), `AuthScheme`, `CRUDPaths`, `FieldTemplate`/`TemplateArg`; `Key()` helper
- [x] `ports.go`: `Store`, `Validator`, `DescriptorSource` interfaces; `Record`, `ListParams`, `ListResult`; `ValidationError` (field→message)
- [x] `service.go`: `Service` (validate-before-store on writes) + `Runtime` (descriptor + service)
- [x] `registry.go`: concurrent, atomically-swappable `Registry` grouped by backend (Replace/Lookup/Backends/ResourcesFor/First)
- [x] Pure: imports only stdlib (context, sync, sort, strings)
- [x] Tests `test/unit/domain/resource/resource_test.go` (registry ops + service validation gating)

## Touched files

- `internal/domain/resource/{descriptor,ports,service,registry}.go`
- `test/unit/domain/resource/resource_test.go`

## Verification

- `go build ./...`, `go vet ./internal/domain/...` exit 0
- `go test ./test/unit/domain/...` pass

## Follow-ups

- `source/manual` + `source/openapi` (step 5) build `[]ResourceDescriptor`.
- `upstream` (step 8) implements `Store`; `validation` (step 9) implements `Validator`.
- `app/resources.go` assembles `Runtime`s and `Registry.Replace`.
