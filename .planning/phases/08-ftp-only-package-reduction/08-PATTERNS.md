# Phase 8 Pattern Map

## Closest Existing Artifacts

- `ftpsync/options.go`, `ftpsync/service.go`, `ftpsync/errors.go`, `ftpsync/hooks.go` — public API contract that must remain source-compatible.
- `ftpsync/oneshot.go` — current one-shot implementation and primary dependency leak; replace legacy adapter with package-local FTP operations.
- `ftpsync/background.go` — already library-local except for its call into `executeSyncOnce`; preserve by keeping the same `executeSyncOnce` contract.
- `driver/ftp/ftp.go`, `driver/ftp/encoding.go`, `driver/ftp/file*.go` — behavior source for FTP path encoding, passive mode, timeout, traversal, and CRUD operations; copy/internalize the minimal behavior, not the broad driver abstraction.
- `ftpsync/*_test.go` — regression baseline for public behavior; extend dependency tests rather than weakening them.

## Code Patterns to Preserve

- Public tests use package `ftpsync` or external-style package tests to prove consumers do not rely on internals.
- Error classification uses `ErrorKind` plus `errors.As`/`errors.Is` compatibility.
- Hook callbacks are synchronous, compact, and credential-safe.
- Background lifecycle reuses `executeSyncOnce`, waits for workers, and keeps runtime sync failures non-terminal.

## Patterns to Avoid

- Do not hide old runtime imports behind build tags.
- Do not keep `core.VFS`, legacy `sync.Option`, `logger.Logger`, or `retry.Retry` in `ftpsync` internals.
- Do not weaken dependency-boundary tests to permit the old runtime chain.
