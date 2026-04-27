---
phase: 08-ftp-only-package-reduction
plan: 02
subsystem: library-core
tags: [go, ftp, ftpsync, dependency-pruning, one-shot-sync]
requires:
  - phase: 08-ftp-only-package-reduction
    provides: dependency blacklist and typed public API guards
provides:
  - Package-local FTP client operations for ftpsync one-shot push and pull flows
  - Context-aware retry helper independent of legacy retry/logger packages
  - SyncOnce execution path without core, sync, ignore, logger, or retry imports
affects: [phase-08-package-reduction, phase-09-verification-docs, ftpsync]
tech-stack:
  added: []
  patterns:
    - package-local FTP seam using github.com/jlaffaye/ftp behind unexported ftpCore interface
    - typed-options-only one-shot execution without legacy VFS adapter construction
key-files:
  created:
    - ftpsync/internal_ftp.go
    - ftpsync/internal_retry.go
  modified:
    - ftpsync/oneshot.go
    - ftpsync/oneshot_test.go
    - ftpsync/background_test.go
key-decisions:
  - "Replaced SyncOnce legacy VFS/sync adapter with a package-local FTP client seam so ftpsync no longer imports core, sync, logger, retry, or ignore."
  - "Kept retry and ignore behavior inside ftpsync using typed options, preserving public API shape while shrinking the default package graph."
patterns-established:
  - "Package-local FTP core: one-shot code calls an unexported ftpCore interface and openFTPClient factory for fakeable tests."
  - "Context-aware retry: retryWithContext bounds attempts from RetryOptions and exits promptly on context cancellation."
requirements-completed: [PRUNE-01, PRUNE-02, PRUNE-03]
duration: 32min
completed: 2026-04-27
---

# Phase 08 Plan 02: Package-Local FTP One-Shot Core Summary

**SyncOnce now runs disk↔FTP flows through ftpsync-local FTP and retry helpers instead of the legacy core/sync/logger/retry runtime adapter.**

## Performance

- **Duration:** 32 min
- **Started:** 2026-04-27T10:09:10Z
- **Completed:** 2026-04-27T10:41:14Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Added `ftpsync/internal_ftp.go` with a small unexported FTP client wrapper over `github.com/jlaffaye/ftp`, including connection options, path encoding, directory creation, upload, removal, walking, download, and close behavior.
- Added `ftpsync/internal_retry.go` with context-aware bounded retry behavior sourced from public `RetryOptions`, without importing old retry/logger packages.
- Rewrote `ftpsync/oneshot.go` so `executeSyncOnce` dispatches local→FTP and FTP→local through package-local helpers and no longer imports legacy `core`, `sync`, `ignore`, `logger`, or `retry` packages.
- Updated one-shot and background tests to use package-local seams/fakes, preserving compact result, partial failure, hook, cwd-safety, and background lifecycle coverage without a live FTP server.
- Verified the dependency blacklist for `./ftpsync` now passes after removing the legacy one-shot adapter import chain.

## Task Commits

Each task was committed atomically:

1. **Task 1: Introduce package-local FTP and retry helpers** - `b72c339` (feat)
2. **Task 2: Remove legacy adapter imports from one-shot execution** - `3d0cc13` (feat)

**Plan metadata:** pending final docs commit

## Files Created/Modified

- `ftpsync/internal_ftp.go` - Package-local FTP client, path codec, FTP file info, and FTP path helpers.
- `ftpsync/internal_retry.go` - Context-aware retry helper driven by `RetryOptions`.
- `ftpsync/oneshot.go` - One-shot local→FTP and FTP→local orchestration without legacy adapter imports.
- `ftpsync/oneshot_test.go` - Package-local fake FTP client tests for one-shot behavior and partial failures.
- `ftpsync/background_test.go` - Updated background test executor seams after removing adapter parameter.

## Decisions Made

- Replaced the legacy `syncOnceAdapter` with a direct unexported `ftpCore` interface so tests can fake FTP behavior while production uses `github.com/jlaffaye/ftp`.
- Kept ignore matching in `ftpsync/oneshot.go` rather than importing the old `ignore` package, preserving literal/glob/regexp typed-rule semantics needed by the public API.
- Implemented FTP→local writes directly with destination-root mapping and `ensureTargetUnderRoot` rather than relying on legacy sync materialization, preserving cwd safety at the library boundary.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Adapted test seam signatures after removing syncOnceAdapter**
- **Found during:** Task 2 (Remove legacy adapter imports from one-shot execution)
- **Issue:** Existing one-shot and background tests accepted `syncOnceAdapter`, which no longer exists after removing the legacy adapter.
- **Fix:** Updated test executor seam signatures and fakes to use the package-local FTP factory instead.
- **Files modified:** `ftpsync/oneshot_test.go`, `ftpsync/background_test.go`
- **Verification:** `go test ./ftpsync -run 'Test(SyncOnce|StartBackground|PackageDependencyBoundaryRejectsOldRuntime|PublicAPIStaysTypedOptionsOnly)' -count=1`
- **Committed in:** `3d0cc13`

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Required to complete the planned adapter removal; no scope expansion.

## Issues Encountered

- The system `PATH` did not include `gofmt`; formatting was run via `/usr/local/go/bin/gofmt`.
- The `github.com/jlaffaye/ftp` concrete `Retr` method returns `*ftp.Response`; the package-local interface was adapted with a small wrapper returning `io.ReadCloser` for fakes.

## Verification

- `go test ./ftpsync -run 'TestSyncOnce(LocalToFTP|FTPToLocal|Partial|NeverWritesToCWD|Builds)' -count=1`
- `go test ./ftpsync -run 'Test(SyncOnce|StartBackground|PackageDependencyBoundaryRejectsOldRuntime|PublicAPIStaysTypedOptionsOnly)' -count=1`
- `go test ./ftpsync -count=1`
- `go list -deps ./ftpsync`

## Known Stubs

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 08 Plan 03 can now delete or isolate old runtime packages knowing `ftpsync` no longer depends on the legacy one-shot sync runtime.
- Phase 08 Plan 04 can use the passing dependency blacklist and `go list -deps ./ftpsync` output as evidence for module dependency cleanup.

---
*Phase: 08-ftp-only-package-reduction*
*Completed: 2026-04-27*

## Self-Check: PASSED

- Found expected created files: `ftpsync/internal_ftp.go`, `ftpsync/internal_retry.go`, and this summary.
- Verified task commits exist in git log: `b72c339`, `3d0cc13`.
- No unexpected tracked-file deletions were present in task commits.
