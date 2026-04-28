---
phase: 07-background-disk-ftp-lifecycle
plan: 03
subsystem: ftpsync-lifecycle
tags: [go, ftpsync, fsnotify, lifecycle, cancellation]

requires:
  - phase: 07-background-disk-ftp-lifecycle
    provides: background local-to-FTP startup, coalesced watch loop, and Handle.Wait contract
provides:
  - deterministic Stop/context-cancel shutdown for background disk-to-FTP sync
  - final Wait result separated from latest runtime Err status
  - regression coverage for unsupported FTP-to-local background direction
affects: [ftpsync, background-sync, phase-08-package-reduction]

tech-stack:
  added: []
  patterns: [context-aware worker cancellation, waitgroup-backed background lifecycle, local-to-FTP-only public policy]

key-files:
  created: []
  modified: [ftpsync/background.go, ftpsync/background_test.go, ftpsync/context_test.go]

key-decisions:
  - "Background Stop now waits for the active sync worker before closing Done, so embedders can trust Wait as a full lifecycle barrier."
  - "Handle.Err remains the latest runtime health signal while Handle.Wait returns only the terminal lifecycle result after Done closes."
  - "Background sync policy remains local-to-FTP only; FTP-to-local background attempts continue to return ErrUnsupportedCapability."

patterns-established:
  - "Worker goroutines started by background handles are registered in a sync.WaitGroup and joined before Done closes."
  - "Latest runtime errors and final terminal results are stored separately to avoid reporting recoverable sync failures as shutdown failures."

requirements-completed: [WATCH-03, WATCH-04, WATCH-05]

duration: 4min
completed: 2026-04-27
---

# Phase 07 Plan 03: Background Disk→FTP Lifecycle Summary

**Deterministic background disk-to-FTP shutdown with worker joins, final Wait semantics, and local-to-FTP-only policy coverage**

## Performance

- **Duration:** 4 min
- **Started:** 2026-04-27T08:56:31Z
- **Completed:** 2026-04-27T08:59:57Z
- **Tasks:** 2
- **Files modified:** 3

## Accomplishments

- Added deterministic shutdown behavior for background handles by canceling the runner context, waiting for active sync workers, and closing `Done()` only after worker exit.
- Split terminal lifecycle status from latest runtime health: `Wait()` now returns the stored final result while `Err()` continues to report the latest observed runtime failure.
- Expanded tests for Stop, root context cancellation, stop/cancel idempotency, final Wait semantics, and the explicit local→FTP-only background policy.

## Task Commits

Each task was committed atomically:

1. **Task 1: Make Stop and context cancellation shut down the background runner deterministically per D-09 and D-10** - `a4ed120` (feat)
2. **Task 2: Return final lifecycle status cleanly while preserving unsupported background limits per D-11** - `863cf26` (test)

**Plan metadata:** pending final docs commit

## Files Created/Modified

- `ftpsync/background.go` - Adds worker waitgroup joining and returns terminal `final` errors from `Wait()` instead of latest runtime errors.
- `ftpsync/background_test.go` - Adds shutdown, cancellation, race/idempotency, and final lifecycle status regression tests.
- `ftpsync/context_test.go` - Adds table coverage proving background support remains limited to local→FTP and FTP→local remains unsupported.

## Decisions Made

- Background `Stop`/cancel now treats `Done()` as a full lifecycle barrier, not just a watch-loop exit signal, by waiting for sync worker goroutines before closing it.
- Recoverable background sync errors remain visible through `Err()` but do not become terminal `Wait()` failures after a clean Stop.
- No policy switch or FTP→local background path was added; unsupported direction behavior remains a hard public API boundary.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The initial Stop shutdown test passed against the pre-existing implementation because it did not prove an active worker was joined. The test was strengthened to hold an in-flight sync pass until Stop had canceled the context and to assert Stop did not return before that worker exited.

## User Setup Required

None - no external service configuration required.

## Verification

- `go test ./ftpsync -run 'TestBackgroundStopShutsDownDeterministically|TestBackgroundContextCancelStopsRunner' -count=1`
- `go test ./ftpsync -run 'TestBackgroundWaitReturnsFinalError|TestStartBackgroundRejectsFTPToLocal|TestStartBackgroundDirectionPolicyRemainsLocalToFTPOnly' -count=1`
- `go test ./ftpsync -run 'TestBackground(StopShutsDownDeterministically|ContextCancelStopsRunner|WaitReturnsFinalError)|TestStartBackgroundRejectsFTPToLocal' -count=1`
- `go test ./ftpsync -count=1`

## Known Stubs

None found in files created or modified by this plan.

## Self-Check: PASSED

- Verified `ftpsync/background.go`, `ftpsync/background_test.go`, `ftpsync/context_test.go`, and this summary file exist.
- Verified task commits `a4ed120` and `863cf26` exist in recent git history.

## Next Phase Readiness

- Phase 7 background disk→FTP lifecycle is complete and ready for milestone verification or Phase 8 package reduction.
- The final lifecycle contract is suitable for embedders that need to stop, wait, and inspect runtime health deterministically.

---
*Phase: 07-background-disk-ftp-lifecycle*
*Completed: 2026-04-27*
