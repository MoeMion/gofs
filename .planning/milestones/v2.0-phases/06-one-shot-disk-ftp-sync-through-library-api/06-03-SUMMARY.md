---
phase: 06-one-shot-disk-ftp-sync-through-library-api
plan: "03"
subsystem: api
tags: [ftp, ftpsync, oneshot, ftp-to-local, cwd-safety, go, testing]

# Dependency graph
requires:
  - phase: 06-01
    provides: Compact one-shot result scaffolding, typed transfer errors, and internal typed-options adapter
  - phase: 06-02
    provides: Library-side best-effort one-shot orchestration pattern and compact summary/hook behavior
provides:
  - Public FTP→local one-shot execution through `FTPSyncService`
  - Explicit destination-root creation and cwd-safety regression coverage
  - FTP→local best-effort behavior aligned with Phase 6 `Result + error` semantics
affects: [phase-06, public-api, oneshot-execution, ftp-library, cwd-safety]

# Tech tracking
tech-stack:
  added: []
  patterns: ["FTP→local now uses the same library-side best-effort orchestration style as local→FTP while delegating file mutations to legacy pull sync methods", "Public cwd-safety coverage is enforced through FTPSyncService tests instead of only lower-level sync tests"]

key-files:
  created: [.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-03-SUMMARY.md]
  modified: [ftpsync/oneshot.go, ftpsync/oneshot_test.go, sync/driver_pull_client_sync.go, sync/sync.go]

key-decisions:
  - "Implemented FTP→local as a library-side best-effort walker instead of calling legacy SyncOnce directly, because direct SyncOnce would stop on first file failure and violate the locked Phase 6 best-effort decision."
  - "Kept destination-root enforcement explicit at the public API layer so FTP→local never falls back to cwd, even when the process working directory changes."
  - "Exposed source-tree walking through a small legacy `SourceWalker` interface so ftpsync can reuse pull semantics without reviving CLI/runtime entrypoints."

patterns-established:
  - "Library-side one-shot orchestration may add a narrow legacy seam when needed, as long as public API shape remains unchanged."
  - "Best-effort FTP→local behavior should continue later paths after file-level failures and still return compact `Result + error` output."

requirements-completed: [ONCE-02, ONCE-03, ONCE-04]

# Metrics
duration: 18min
completed: 2026-04-27
---

# Phase 06 Plan 03: FTP→Local One-Shot and CWD Safety Summary

**FTPSyncService now performs one-shot FTP→local execution with explicit destination-root creation, cwd-safety enforcement, and best-effort partial-failure behavior.**

## Performance

- **Duration:** 18 min
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments

- Implemented public `FTP→local` one-shot execution behind `FTPSyncService.SyncOnce`.
- Added explicit destination-root creation and root-boundary enforcement so writes stay under the configured local root.
- Added public regression tests for auto-create-root behavior, cwd safety, successful pull execution, and partial-failure `Result + error` handling.

## Files Created/Modified

- `ftpsync/oneshot.go` - Adds `DirectionFTPToLocal` execution branch, source walking, destination-root safety checks, and best-effort FTP→local orchestration.
- `ftpsync/oneshot_test.go` - Adds FTP→local success, auto-create-root, cwd-safety, and partial-failure tests plus source-walker fakes.
- `sync/driver_pull_client_sync.go` - Exposes legacy source walking and link-reading helpers for reuse by the library adapter.
- `sync/sync.go` - Adds the small legacy `SourceWalker` seam used by ftpsync.
- `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-03-SUMMARY.md` - Records plan completion, decisions, and verification.

## Decisions Made

- Replaced direct legacy `SyncOnce(remoteRoot)` usage with library-side per-path orchestration for FTP→local so Phase 6 best-effort behavior stays consistent across both directions.
- Treated file-level FTP→local failures as transfer failures that still preserve completed work counts and later successful writes.
- Kept cwd protection as a public API regression invariant rather than trusting lower-level path defaults.

## Deviations from Plan

### Auto-fixed Issues

1. **Best-effort pull needed a narrower legacy seam than originally assumed**
   - **Issue:** The first implementation attempted to reuse legacy FTP pull `SyncOnce` directly, but that path stopped on the first file failure and could not satisfy the locked best-effort Phase 6 behavior.
   - **Fix:** Added a minimal `sync.SourceWalker` interface plus `WalkSourceDir` / `ReadSourceLink` methods on the legacy pull sync so `ftpsync` could orchestrate per-path best-effort behavior without changing the public API.
   - **Impact:** No public API expansion; only a small internal seam was added.

## Verification

- `go test ./ftpsync -run 'TestSyncOnceFTPToLocal(AutoCreateRoot|Success|NeverWritesToCWD|PartialFailure)' -count=1` — passed
- `go test ./ftpsync ./sync ./driver/ftp -run 'TestSyncOnceFTPToLocal' -count=1` — passed
- `go fmt ./ftpsync ./sync` — passed

## Self-Check

- ✓ `FTP→local` no longer returns unsupported capability from `SyncOnce`
- ✓ explicit local root is auto-created when missing
- ✓ cwd-safety regression is covered through the public API
- ✓ partial failures return compact `Result + error`
