# Coding Conventions

**Analysis Date:** 2026-04-23

## Naming Patterns

**Files:**
- Use lowercase package directories with short nouns such as `action`, `auth`, `conf`, `core`, `server`, and `sync`.
- Use snake_case filenames for multiword files such as `cmd/gofs/main.go`, `server/session_test.go`, `api/apiclient/grpc_client.go`, and `sync/remote_client_sync.go`.
- Reserve generated suffixes for protobuf output in `api/task/task.pb.go`, `api/task/task_grpc.pb.go`, `api/auth/auth.pb.go`, and `api/auth/auth_grpc.pb.go`.
- Use Go test suffixes consistently: unit tests in `*_test.go` such as `action/action_test.go`, integration suites in `integration/*_test.go`, and benchmarks in files like `internal/toplist/toplist_benchmark_test.go`.

**Functions:**
- Exported functions use PascalCase, for example `cmd.RunWithConfig` in `cmd/gofs.go`, `logger.NewTestLogger` in `logger/logger.go`, and `server.GenerateAddr` in `server/server.go`.
- Unexported helpers use camelCase, for example `runWithConfig` in `cmd/gofs.go`, `parseIgnoreFile` in `ignore/ignore.go`, and `testIntegrationClientServer` in `integration/integration_test.go`.
- Constructor-style functions commonly use `New*` naming, such as `sync.NewSyncOption` in `sync/option.go`, `server.NewServerOption` in `server/option.go`, `logger.NewConsoleLogger` in `logger/logger.go`, and `result.New` in `result/result.go`.

**Variables:**
- Configuration-heavy code prefers short local aliases like `c`, `cp`, `r`, `pi`, `srv`, and `wd`, visible in `cmd/gofs.go`, `sync/option.go`, `server/option.go`, and `result/result.go`.
- Error values are almost always named `err` and checked immediately after the call site, as in `cmd/gofs.go`, `core/vfs.go`, `api/task/task.go`, and `integration/integration_test.go`.
- Test tables use `testCases` with per-case `tc` iteration, visible in `action/action_test.go`, `server/server_test.go`, `conf/parser_test.go`, and `server/session_test.go`.

**Types:**
- Struct types use concise domain nouns such as `Config` in `conf/config.go`, `Option` in `sync/option.go` and `server/option.go`, and `Logger` in `logger/logger.go`.
- Interfaces are small and behavior-oriented, such as `Ignore` in `ignore/ignore.go` and `Result` in `result/result.go`.

## Code Style

**Formatting:**
- No repository-local formatter config is detected; formatting follows standard Go tooling conventions.
- Source files are gofmt-style with tabs for indentation and grouped imports, visible in `cmd/gofs.go`, `core/vfs.go`, and `logger/logger.go`.
- Comments generally start with the exported identifier name to match Go doc expectations, for example `// RunWithConfig...` in `cmd/gofs.go`, `// Config...` in `conf/config.go`, and `// GenerateAddr...` in `server/server.go`.

**Linting:**
- No `.golangci.yml`, `.golangci.yaml`, `.editorconfig`, or other repo-local lint config is detected at the project root.
- The codebase includes targeted cyclomatic-complexity suppressions with `//gocyclo:ignore` in `cmd/gofs.go` and `encrypt/encrypt_test.go`.
- CI quality checks focus on build, race-enabled tests, vulnerability scanning, and CodeQL via `.github/workflows/go.yml`, `.github/workflows/govulncheck.yml`, and `.github/workflows/codeql.yml`.

## Import Organization

**Order:**
1. Standard library imports first, for example in `cmd/gofs.go`, `core/vfs.go`, and `api/apiclient/grpc_client.go`.
2. Internal project packages under `github.com/no-src/gofs/...` next, as in `cmd/gofs.go` and `api/api_test.go`.
3. Third-party packages appear after or alongside internal packages depending on gofmt grouping, such as `github.com/no-src/nsgo/fsutil` in `cmd/gofs.go`, `github.com/kevinburke/ssh_config` in `core/vfs.go`, and gRPC packages in `api/task/task.go`.

**Path Aliases:**
- Import aliases are used sparingly and only for clarity or collision avoidance, such as `authapi` in `api/apiclient/grpc_client.go`.
- No custom module path aliases are defined; imports use full module paths from `go.mod`.

## Dependency Management Habits

- Dependencies are managed with Go modules through `go.mod` and `go.sum` at the repository root.
- The module path is `github.com/no-src/gofs` in `go.mod`.
- Direct dependencies are declared explicitly in `go.mod`, while transitive packages remain indirect there rather than being wrapped in additional manifests.
- New external libraries should be added through `go.mod`; there is no secondary package manager or vendored dependency directory detected.
- Internal packages are reused extensively instead of duplicating logic, for example `cmd/gofs.go` composes `auth`, `conf`, `logger`, `monitor`, `report`, `retry`, `server`, `sync`, and `wait`.

## Error Handling

**Patterns:**
- Return errors directly and guard immediately after calls, as in `ignore/ignore.go`, `result/result.go`, and `api/task/task.go`.
- Wrap errors with context using `fmt.Errorf`, often with `%w`, for example in `cmd/gofs.go`, `sync/sync.go`, `server/session.go`, and `core/vfs.go`.
- Surface gRPC transport errors with `status.Errorf` in `api/task/task.go`.
- Use sentinel errors with `errors.Is` or `errors.As` in tests, as shown in `conf/parser_test.go`, `server/session_test.go`, and `encrypt/encrypt_test.go`.
- Logging helpers frequently combine side effects and error propagation, such as `logger.ErrorIf(...)` in `cmd/gofs.go` and server-related packages found by repository search.

## Logging

**Framework:**
- Logging is centralized through `github.com/no-src/log` wrapped by `logger.Logger` in `logger/logger.go`.

**Patterns:**
- Use `logger.NewConsoleLogger`, `logger.NewTestLogger`, or `logger.NewEmptyLogger` from `logger/logger.go` instead of constructing loggers inline.
- Use `InnerLogger()` for bootstrapping or parsing paths before runtime loggers are initialized, as in `core/vfs.go` and `cmd/gofs.go`.
- Test code often uses `logger.NewTestLogger()` to preserve realistic log behavior, such as `encrypt/encrypt_test.go`, `ignore/ignore_test.go`, and multiple `internal/*` tests.

## Comments

**When to Comment:**
- Comments describe exported identifiers and significant state transitions rather than line-by-line logic.
- Inline comments are used to explain operational branches, such as subprocess/daemon handling in `cmd/gofs.go` and SSH config parsing in `core/vfs.go`.

**JSDoc/TSDoc:**
- Not applicable. This repository is Go-based; documentation comments follow Go doc comment style.

## Function Design

**Size:**
- Most functions are compact and single-purpose, especially constructors and parsers in `logger/logger.go`, `server/option.go`, and `ignore/ignore.go`.
- Orchestrator functions can be large when coordinating lifecycle steps; `runWithConfig` in `cmd/gofs.go` is the clearest example and is explicitly marked with `//gocyclo:ignore`.

**Parameters:**
- Option-builder functions accept full config structs plus collaborating services, for example `sync.NewSyncOption` in `sync/option.go` and `server.NewServerOption` in `server/option.go`.
- Helper functions often pass `*testing.T` first in tests, then scenario-specific strings or structs, such as `testIntegrationClient` in `integration/integration_local_test.go`.

**Return Values:**
- Constructors commonly return `(value, error)` or interface types, such as `ignore.New` in `ignore/ignore.go` and `task.RegisterServer` in `api/task/task.go`.
- Lifecycle APIs return domain-specific interfaces, such as `result.Result` from `cmd.RunWithConfig` in `cmd/gofs.go` and `result.New` in `result/result.go`.
- Named returns are used selectively in more stateful functions, for example `func (c *client) Start() (err error)` in `api/apiclient/grpc_client.go` and `func NewSyncOption(...) Option` in `sync/option.go`.

## Module Design

**Exports:**
- Packages expose a small public surface around structs, interfaces, and constructors, with implementation details kept unexported, as in `ignore/ignore.go`, `api/task/task.go`, and `logger/logger.go`.
- Config and option packages favor data structs with field-based composition, as in `conf/config.go`, `sync/option.go`, and `server/option.go`.

**Barrel Files:**
- Not applicable. Go packages use package directories rather than barrel-export files.

## Config Practices

- Runtime configuration is centralized in `conf.Config` in `conf/config.go` with both `json` and `yaml` struct tags.
- Command execution flows parse flags into config first, then optionally overlay config file contents, as shown in `cmd/gofs.go` and tested in `conf/parser_test.go`.
- Config serialization back to CLI args is supported through `Config.ToArgs()` in `conf/config.go`.
- Sensitive configuration fields exist in the config struct, such as TLS paths and token/encryption secrets in `conf/config.go`; document file locations and names only, not secret values.
- Repo-local environment files are not part of the visible quality toolchain; no `.env` usage was read for this mapping.

## Code Style Observations

- Concurrency and lifecycle control are wrapped in small abstractions like `wait.WaitDone` and `result.Result` in `result/result.go`, then composed from orchestration code in `cmd/gofs.go` and integration tests in `integration/integration_test.go`.
- Test code favors explicit assertions with `if ... { t.Errorf(...) }` instead of assertion libraries, visible across `action/action_test.go`, `server/server_test.go`, `conf/parser_test.go`, and `api/api_test.go`.
- The codebase stays within the standard library plus a modest set of focused third-party packages declared in `go.mod`; no metaprogramming or custom code generation beyond protobuf output is detected.

---

*Convention analysis: 2026-04-23*
