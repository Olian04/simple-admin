# Step 3 — Config (backends, manual|discovery)

Status: `done`

## Sub-tasks

- [x] `internal/config/resources.go`: `BackendSection` (manual|discovery, mutually exclusive), `ResourceSection`, `ResourcePaths`, `DiscoverySection`, `AuthScheme`, `SchemaSource` (path | `{url}`), `FieldTemplate` (shorthand | `{template,args}` | `{template_file}`), `TemplateArg` (static | `{jsonpath}`)
- [x] Custom `UnmarshalYAML` for the union types (SchemaSource/FieldTemplate/TemplateArg)
- [x] `WithDefaults`: derive `id` (sha256(base_url) prefix) when omitted, label titleize, id_field=id, auth.type=none, poll_interval=5m
- [x] `Validate`: slug ids, base_url parse, exactly-one mode, discovery forbids backend-level auth, apikey requires in/name, bearer/basic/oauth → "not yet supported", backend auth forbids `value`, item path contains {id}, unique ids/resource names
- [x] Wire `Backends` into root `Config` (WithDefaults loop + Validate loop + id uniqueness); `fmt` import
- [x] `defaults.go`: DefaultAuthType/IDField/PollInterval
- [x] `load.go`: `expandSecrets` env-expands `discovery.auth.value` only
- [x] `configs/config.example.yaml`: manual + discovery examples
- [x] Tests `test/unit/config/resources_test.go` (parse, defaults, env expand, all reject cases, example-config-loads guard)

## Touched files

- `internal/config/{resources.go,config.go,defaults.go,load.go}`
- `configs/config.example.yaml`
- `test/unit/config/resources_test.go`

## Verification

- `go build ./...`, `go vet ./...` exit 0
- `go test ./test/unit/config/...` pass (11 cases incl. example-config load)

## Commit

- (next)

## Follow-ups

- Domain (step 4) maps these config types → `ResourceDescriptor` via `source/manual`.
- Item path uses `{id}`; PocketBase update verb is PATCH (capabilities, step 5/8).
