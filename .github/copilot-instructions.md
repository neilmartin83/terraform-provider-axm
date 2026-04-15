# Copilot instructions

## Build, test, and lint
- Build: `go build -v ./...` or `make build`
- Install: `go install -v ./...` or `make install`
- Format: `gofmt -s -w -e .` or `make fmt`
- Lint: `golangci-lint run` or `make lint`
- Generate docs/code: `cd tools && go generate ./...` or `make generate`
- Unit tests: `go test -v -cover -timeout=120s -parallel=10 ./...` or `make test`
- Single test: `go test ./internal/... -run '^TestName$'`
- Acceptance tests (require AXM_* env vars): `TF_ACC=1 go test -v -cover -timeout 120m ./...` or `make testacc`

## High-level architecture
- Terraform provider built on `terraform-plugin-framework` v1.17.0 (protocol v6) for the Apple School/Business Manager API.
- **internal/provider** defines provider schema/config, registers resources/data sources/list resources, and sets up logging.
- **internal/client** implements OAuth2 JWT auth, token caching, rate-limit handling (Retry-After <= 60s retries, otherwise error), and all API calls.
- **internal/resources** contains the provider constructs: resource + list resource for device management service, plus data sources for devices, MDM servers, assigned server info, and AppleCare coverage.
- **internal/common** holds shared helpers (timeouts, configure, type conversions) to be reused across resources/clients.
- **docs/** are generated from schema descriptions; **examples/** and **testing/** contain Terraform configs and manual API scripts.

## Key conventions
- Resource packages follow a consistent file split: `resource.go`, `crud.go`, `model_types.go`, `schema_types.go`, `helpers.go`, `list_resource.go`, `data_source.go` (with optional `mappings.go`, `input_builders.go`, `state_builders.go`).
- Prefer shared helpers in **internal/common** over duplicating logic in resource packages.
- Schema rule: use **sets** for user-supplied unordered collections; use **lists** for computed API data from data sources.
- Keep dependencies to native Go, `golang.org/x`, and Terraform Plugin Framework packages only.
- Code style: ensure short comments for constants, functions, variable sets, and types; avoid comments inside type definitions or function bodies.
