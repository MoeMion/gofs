# Phase 09 Research: Verification, Examples, and Migration Docs

**Phase:** 09-verification-examples-and-migration-docs  
**Date:** 2026-04-28  
**Status:** Complete

## Research Question

What is needed to plan final verification, examples, and migration documentation for the reduced `ftpsync` local Go module?

## Inputs

- `.planning/ROADMAP.md` Phase 9 success criteria
- `.planning/REQUIREMENTS.md` VERIFY-01, VERIFY-02, VERIFY-03, DOC-01, DOC-02, DOC-03
- `.planning/phases/09-verification-examples-and-migration-docs/09-CONTEXT.md` locked decisions D-01 through D-15
- Current package files under `ftpsync/`
- Context7 docs for `/fclairamb/ftpserverlib`

## Findings

### Module path and import contract

- The current root module is `github.com/no-src/gofs` in `go.mod`.
- Locked decision D-02 requires changing the root module path to `ftpsync`.
- Current external-style tests in `ftpsync_test` import `github.com/no-src/gofs/ftpsync`; these imports must become `ftpsync` after the module path change.
- Dependency boundary tests currently blacklist old `github.com/no-src/gofs/...` runtime package paths. After the module path change, the blacklist should continue to reject legacy dependency fragments and should also prove that `go list ./...` exposes only the final package path expected for the local module.

### Existing automated coverage

Existing tests already cover significant VERIFY-01 and VERIFY-03 pieces:

- `ftpsync/validation_test.go` covers supported/unsupported endpoint validation, password-safe validation errors, FTP option fields, and public `FTPSyncService` construction.
- `ftpsync/context_test.go` covers context cancellation for `SyncOnce` and `StartBackground`, unsupported FTP→local background behavior, and public context-aware contracts.
- `ftpsync/oneshot_test.go` covers one-shot local→FTP, FTP→local, cwd safety, partial failures, typed FTP option propagation, and compact result summaries through package-local fakes.
- `ftpsync/background_test.go` covers background local→FTP lifecycle, debounce/coalescing, non-terminal sync errors, readiness, stop/cancel behavior, and deterministic shutdown.
- `ftpsync/dependency_boundary_test.go` proves package graph reduction and old runtime dependency exclusion.

Remaining verification needs:

- A single checklist-style test should make the final required coverage discoverable so VERIFY-01 and VERIFY-03 cannot regress silently.
- Real FTP integration tests must exercise public `FTPSyncService.SyncOnce` against an actual FTP protocol server for both local→FTP and FTP→local flows, not just fakes.
- Path encoding and passive-mode defaults should remain explicitly asserted after the module path change.

### Real FTP test server choice

Locked decisions D-04 and D-06 require a Go-native test server/fixture, loopback-only, test-scoped, with no production FTP server mode.

Context7 resolved `FTP Server Library` to `/fclairamb/ftpserverlib`:

- Library: `github.com/fclairamb/ftpserverlib`
- Purpose: Build FTP servers with pluggable drivers and filesystem backends.
- Relevant API pattern from docs:
  - Implement a main driver with `GetSettings`, `ClientConnected`, `ClientDisconnected`, `AuthUser`, and `GetTLSConfig`.
  - Configure `ftpserver.Settings` with `ListenAddr`, `PublicHost`, `PassiveTransferPortRange`, `ConnectionTimeout`, and `DefaultTransferType`.
  - Start with `server.Listen()` then `go server.Serve()`; stop with `server.Stop()`.
  - For test fixtures, bind `ListenAddr` to `127.0.0.1:0` where supported by the listener, advertise `PublicHost: "127.0.0.1"`, and use passive transfer ports on loopback only.

Recommended approach:

- Add `github.com/fclairamb/ftpserverlib` as a test dependency.
- Build a test-only helper in `ftpsync/real_ftp_integration_test.go`.
- The helper should start an FTP server on loopback, authenticate one fixed test user, serve a temp directory, return host/port/root credentials to tests, and stop in `t.Cleanup`.
- The tests should call `NewFTPSyncService(...).SyncOnce(ctx)` using typed options and assert files moved through the server root.

### Documentation state

- `README.md` still describes the old CLI/server application, Docker install, SFTP, MinIO, remote server/client, task mode, file server, and legacy FTP URL examples.
- Phase 9 must rewrite or substantially replace README content so it describes the reduced local `ftpsync` module and typed Go options.
- Migration docs should be explicit about the breaking shift from the old CLI/server application to local Go package invocation.

## Standard Stack

- Go standard `testing` package for unit, integration, and Example tests.
- `github.com/fclairamb/ftpserverlib` for Go-native loopback FTP integration fixture.
- Existing `github.com/jlaffaye/ftp` remains the production FTP client dependency.
- Existing `github.com/fsnotify/fsnotify` remains the background watcher dependency.

## Architecture Patterns to Preserve

- Keep all FTP server fixture code in `_test.go` files so no FTP server mode enters production builds.
- Exercise the public `FTPSyncService` API for real integration coverage; do not test only lower-level `ftpClient` helpers.
- Keep public examples in `package ftpsync_test` so they compile as consumer-facing examples using `import "ftpsync"`.
- Keep all configuration in typed `Options`, `Endpoint`, `FTPOptions`, `RetryOptions`, `IgnoreRule`, and `HookSet`; do not reintroduce YAML, CLI flags, URL parser, server/task, Docker, SFTP, or MinIO concepts.

## Common Pitfalls

- Leaving external tests importing `github.com/no-src/gofs/ftpsync` after changing `go.mod` will break the locked local module decision.
- Adding a test FTP server package outside `_test.go` would violate the no production FTP server mode decision.
- Rewriting README without Go Example tests would fail D-07 and DOC-01 because examples must be compiler-checked.
- Real FTP tests must bind only to loopback and must not depend on Docker, Python, or external services.

## Validation Architecture

The final phase should pass the following verification commands:

```bash
go test ./...
go test ./ftpsync -run 'TestIntegrationRealFTP' -count=1
go test ./ftpsync -run 'Example' -count=1
go list ./...
go list -deps ./...
```

Expected evidence:

- `go test ./...` succeeds after module path and dependency updates.
- Real FTP integration tests prove public local→FTP and FTP→local one-shot flows.
- Example tests compile with `import "ftpsync"`.
- Dependency boundary test continues rejecting old runtime imports and stale non-FTP dependencies.
- README links to migration documentation and documents only the supported local module package contract.
