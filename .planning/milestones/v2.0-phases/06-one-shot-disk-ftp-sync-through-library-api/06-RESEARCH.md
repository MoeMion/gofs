# Phase 06 Research — One-Shot Disk↔FTP Sync Through Library API

**Date:** 2026-04-27
**Phase:** 06-one-shot-disk-ftp-sync-through-library-api
**Question:** What does Phase 6 need so `ftpsync.FTPSyncService` can execute one-shot disk→FTP and FTP→disk sync without reintroducing CLI/runtime coupling?

## Research Summary

Phase 6 should keep the public `ftpsync` package thin and reuse the existing FTP v1 sync machinery through a library-local adapter. The minimum-correct path is:

1. Keep `FTPSyncService.SyncOnce(ctx)` as the only public execution entrypoint.
2. Build internal `core.VFS` + `sync.Option` values from typed `ftpsync.Options`.
3. Reuse `sync.NewSync` so direction routing still lands in `NewFTPPushClientSync` and `NewFTPPullClientSync`.
4. Implement best-effort execution and summary/result collection in `ftpsync`, not by modifying public API shape or reviving CLI/report packages.
5. Preserve cwd safety by requiring explicit local roots and proving FTP→local writes stay under the configured destination root.

## Locked Inputs from Context

- **D-01 / D-02:** `SyncOnce` returns a compact summary `Result`, not a per-file report.
- **D-03 / D-04 / D-05:** one-shot execution is best-effort; partial success returns `Result` plus non-nil error.
- **D-06 / D-07:** do not add new policy knobs; follow existing FTP v1 semantics.
- **D-08 / D-09 / D-10:** FTP→local may auto-create the configured root but must never fall back to cwd.
- **D-11 / D-12:** keep public package `ftpsync`, public entrypoint `FTPSyncService`, and typed Go options only.

## Existing Architecture to Reuse

### Public API boundary already established

`ftpsync/service.go` already provides:

- `NewFTPSyncService(opts Options)`
- `Validate()`
- `SyncOnce(ctx)` and `StartBackground(ctx)` signatures
- typed `ErrorKind` classification in `ftpsync/errors.go`
- no-op hook normalization in `ftpsync/hooks.go`

This means Phase 6 should fill in execution behind the existing public contract rather than changing signatures or adding new public construction surfaces.

### Legacy sync entrypoints already route correctly

`sync/sync.go` dispatches:

- disk→FTP to `NewFTPPushClientSync`
- FTP→disk to `NewFTPPullClientSync`

The FTP-specific constructors are already thin wrappers over the generic driver-backed sync paths:

- `sync/ftp_push_client_sync.go`
- `sync/ftp_pull_client_sync.go`

This matches the project constraint to prefer minimal modifications to the VFS/driver/sync layers.

### FTP semantics already encoded in driver-backed paths

Relevant existing behavior:

- `sync/driver_push_client_sync.go` walks local files, creates remote dirs/files, writes files through the FTP driver, propagates file times, and only mirrors locally when `dest.HasLocalPath()` is explicitly true.
- `sync/driver_pull_client_sync.go` writes remote files to local disk, prefers driver-reported file times when available, and walks the FTP driver tree through `driver.WalkDir`.
- `driver/ftp/ftp.go` already owns connection, passive mode, timeout parsing, path encoding, reconnect, mkdir/create/remove/rename/chtimes behavior, and unsupported operations such as active mode and symlink limitations.

## Key Constraint: Best-Effort vs Existing SyncOnce

The legacy `sync.SyncOnce(path string) error` shape is not enough by itself for D-03/D-04/D-05 because it returns a single error and is structured to stop when traversal callbacks return an error.

### Minimal-correct implication

Best-effort behavior should be implemented in `ftpsync` by:

- creating the underlying FTP syncer through `sync.NewSync(opt)`
- walking the source tree from `ftpsync` itself
- calling the syncer methods (`Create`, `Write`, `Symlink`, etc.) per path
- collecting failures instead of aborting on the first one
- returning a compact summary `Result` plus a typed partial-failure error when some operations fail

This preserves the legacy write logic while moving library-specific orchestration into the public package.

## Recommended Internal Design

### 1. Build a library-local execution adapter

Add an internal `ftpsync` helper file such as `ftpsync/oneshot.go` that:

- converts typed `Options` into internal `core.VFS` values
- builds `retry.Retry`, `ignore.PathIgnore`, and logger dependencies locally
- constructs the underlying syncer via a small seam wrapping `sync.NewSync`
- dispatches by `Direction`

### 2. Keep public Result compact but useful

A good Phase 6 summary should stay high-level and library-friendly. Recommended fields:

- `Direction`
- `SourceRoot`
- `DestinationRoot`
- `StartedAt`
- `CompletedAt`
- `PathsVisited`
- `FilesAttempted`
- `DirectoriesAttempted`
- `FailureCount`
- `Partial` (or equivalent derived state)

Avoid per-file result slices in Phase 6 because that violates D-02.

### 3. Represent partial failure with typed transfer error

Keep the existing `ErrorKind` contract and return a wrapped `ErrTransfer` error for partial failure. The message should mention that the one-shot run completed with failures, while the `Result` carries the compact counts.

### 4. Reuse existing hook surface

`ftpsync/hooks.go` is already library-local and synchronous. Phase 6 can emit:

- log messages at start/finish/error boundaries
- progress snapshots after each file write attempt
- sync events with operation/path/status/error-kind

Do not introduce global loggers, report.Reporter, or eventlog dependencies.

## Converting Typed Options to Existing Internals

### Local endpoints

Use `core.NewDiskVFS(localPath)` for explicit local roots.

### FTP endpoints

Use `core.NewVFS("ftp://...")` with the existing FTP query grammar because the legacy sync layer already expects that internal representation.

Required query values to preserve v1 semantics:

- `path=` only when an explicit local mirror path is intentionally required
- `remote_path=` from `FTPOptions.RemotePath`
- `ftp_user=` and `ftp_pass=`
- `ftp_passive=` from `FTPOptions.PassiveMode`
- `ftp_timeout=` from `FTPOptions.Timeout`
- `ftp_encoding=` from `FTPOptions.PathEncoding`

This does not violate typed-options-only public API, because the string URL remains an internal adapter detail inside `ftpsync`.

## Ignore Rule Handling

Public `IgnoreRule` values are typed, but the legacy sync stack consumes `ignore.PathIgnore`.

Recommended approach:

- implement a small `ftpsync` adapter that evaluates literal/glob/regexp patterns against walked paths and satisfies `ignore.PathIgnore`
- avoid writing temporary ignore files or reviving parser-only configuration

This keeps the public API typed and avoids filesystem/config side effects.

## CWD Safety Findings

The debug note `.planning/debug/sync-files-copied-to-cwd.md` showed the critical rule:

- remote backends must only perform local mirroring when a local path is explicitly configured
- implicit `.` must not become an accidental destination

For Phase 6, FTP→local must therefore:

- require a non-empty local destination root in public validation
- auto-create that root if absent
- ensure every write path resolves under that configured root
- add public-API regression tests that change cwd to a temp directory and prove no synced files appear there

## Risks / Pitfalls

### Pitfall 1: Reusing `sync.SyncOnce` directly loses best-effort behavior

Why: traversal stops on the first returned error.

Mitigation: library-side walker that calls syncer methods and accumulates failures.

### Pitfall 2: Growing the public API to expose legacy config knobs

Why: D-06 forbids policy expansion and Phase 5 locked typed options only.

Mitigation: keep delete/overwrite/compare policies internal and preserve existing behavior.

### Pitfall 3: Reintroducing runtime package dependencies

Why: Phase 8 still needs package reduction; Phase 6 must not undo the public boundary.

Mitigation: confine legacy reuse to internal adapter code and preserve `go list -deps ./ftpsync` dependency-boundary tests.

### Pitfall 4: Weak FTP→local path normalization

Why: remote paths are slash-based and cross-platform.

Mitigation: reuse the existing FTP pull pattern that resets `sourceAbsPath` to `source.RemotePath().Base()` and always join destination paths under the explicit local root.

## Recommended Verification Strategy

### Public package tests (`./ftpsync`)

Add/extend tests for:

- local→FTP one-shot success path through `FTPSyncService`
- FTP→local one-shot success path through `FTPSyncService`
- partial failure returns `Result` plus `ErrTransfer`
- cwd safety when local destination is omitted vs explicitly set
- dependency boundary still excluding CLI/server/report/SFTP/MinIO surfaces from the public package

### Existing legacy-package tests

Run:

- `go test ./ftpsync -count=1`
- `go test ./sync -count=1`
- `go test ./driver/ftp -count=1`

This keeps public-library behavior aligned with the existing FTP backend.

## Validation Architecture

Phase 6 should use `go test` as the feedback loop and require task-level automated verification after every meaningful code change.

- **Quick command:** `go test ./ftpsync -count=1`
- **Wave command:** `go test ./ftpsync ./sync ./driver/ftp -count=1`
- **Focus:** public one-shot API behavior, partial failure semantics, cwd safety, and v1 FTP regression alignment

## Standard Stack

- Existing packages only: `ftpsync`, `sync`, `driver/ftp`, `core`, `retry`, `ignore`, `logger`
- No new third-party dependency is required for Phase 6

## Architecture Patterns

- Public package stays typed and library-local
- Legacy sync/driver packages remain the execution engine
- Direction routing stays in `sync.NewSync`
- FTP-specific semantics remain in existing FTP driver-backed sync paths
- Library-specific result aggregation, partial-failure handling, and hook emission live in `ftpsync`

## Don’t Hand-Roll

- Do not implement a new FTP client in `ftpsync`; reuse `driver/ftp`
- Do not add a second sync engine; reuse `sync.NewSync` and existing push/pull syncers
- Do not add YAML/CLI/url parsing to the public API
- Do not add FTPS, active mode, background FTP→disk polling, or bidirectional conflict resolution

## Plan Implications

Phase 6 should be split into three sequential plans:

1. add the library-local one-shot adapter/result/partial-failure scaffolding
2. implement local→FTP one-shot execution through the adapter with hook/result updates
3. implement FTP→local execution plus explicit cwd-safety regression coverage

This keeps each plan within the context budget while covering all five `ONCE-*` requirements.
