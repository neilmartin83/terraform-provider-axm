# AGENTS.md — terraform-provider-axm

A Terraform provider for the Apple School & Business Manager API, built on
`terraform-plugin-framework` (protocol v6).

## Key commands

| Action | Command |
|--------|---------|
| Build | `go build -v ./...` |
| Format | `gofmt -s -w -e .` |
| Lint | `golangci-lint run` |
| Generate docs | `cd tools && go generate ./...` (requires Terraform installed) |
| Unit tests | `go test -v -cover -timeout=120s -parallel=10 ./...` |
| Single unit test | `go test ./internal/... -run '^TestName$'` |
| Acceptance tests | `TF_ACC=1 go test -v -cover -timeout=120m ./...` |
| Install | `go install -v ./...` |
| Full CI check | `make fmt lint generate` then unit tests (generated docs must be committed) |

## Architecture

- **`main.go`** — entrypoint; runs `providerserver.Serve` at `registry.terraform.io/neilmartin83/axm`
- **`internal/provider/`** — provider schema, env-var config, registers all resources/data sources/list resources
- **`internal/client/`** — OAuth2 JWT client, token caching (disk: `$TMPDIR/.axm/cache/`), rate-limit retry, all API calls
- **`internal/resources/`** — one package per resource type with standard files: `resource.go`, `crud.go`, `model_types.go`, `schema_types.go`, `data_source.go`, `list_resource.go` (and `_test.go` variants)
- **`internal/common/`** — shared helpers: `configure.go` (client extraction), `filters.go`, `scope.go`, `sets.go`, `timeouts.go`, `type_conversions.go`
- **`tools/tools.go`** — generates docs (`tfplugindocs`) and copyright headers (`copywrite`)

## Provider auth

Configurable via provider block attributes or env vars:
`AXM_TEAM_ID`, `AXM_CLIENT_ID`, `AXM_KEY_ID`, `AXM_PRIVATE_KEY`, `AXM_SCOPE`
(`scope` defaults to `business.api`; valid: `business.api`, `school.api`)

## API rate limits & retry

- GET: ~20 req/min, POST: ~10 req/hr (undocumented by Apple)
- On HTTP 429: reads `Retry-After` header. If ≤60s, waits and retries (up to 5). If >60s, returns error.
- Server errors (502/503/504) retry with exponential backoff (2s–30s, 5 attempts).

## Testing

- Unit tests run in CI on every PR/push to main.
- Acceptance tests require live API credentials (`AXM_*` env vars) and are **manual only** (`workflow_dispatch`). Never run `make testacc` without credentials.
- Acceptance runs on Terraform `1.13.*` and `1.14.*` with `-count=1 -p=1` (no cache, serial).
- Test fixtures and manual config live in `testing/` (gitignored).

## Conventional commits & release

- Release-please scans commits on `main` for conventional-commit messages and opens Release PRs.
- Merging a Release PR creates a GitHub release + `v*` tag.
- GoReleaser builds multi-platform binaries on tag push (`goreleaser release --clean`).

## Style conventions

- Resource packages: **sets** for user-supplied unordered collections; **lists** for computed API data from data sources.
- Copyright header: `// Copyright Neil Martin 2026\n// SPDX-License-Identifier: MPL-2.0`
- Dependencies: native Go, `golang.org/x/*`, and Terraform Plugin Framework only.
- Comments on constants/functions/types are OK; avoid comments inside type definitions or function bodies.
