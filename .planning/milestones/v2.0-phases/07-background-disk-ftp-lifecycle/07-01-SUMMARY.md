---
phase: 07-background-disk-ftp-lifecycle
plan: "01"
subsystem: api
tags: [ftp, ftpsync, background-sync, lifecycle, handle, go, testing]

# Dependency graph
requires:
  - phase: 05-02
    provides: Context-aware StartBackground stub, Handle contract, validation ordering, and structured unsupported-capability errors
  - phase: 06-02
    provides: Local disk→FTP one-shot orchestration reused for background initial catch-up
provides:
  - Public local→FTP StartBackground dispatch through FTPSyncService without CLI invocation
  - Concrete background lifecycle handle with Done, Err, Wait, Stop, and internal readiness signaling
  - Initial local→FTP catch-up before background readiness while keeping FTP→local background unsupported
affects: [phase-07, phase-08, phase-09, public-api, background-lifecycle, ftpsync]

# Tech tracking
tech-stack:
  added: []
  patterns: ["Background mode starts as initial catch-up plus long-lived lifecycle handle", "Handle errors are mutex-protected and observable without failing StartBackground on transient sync errors"]

key-files:
  created: [ftpsync/background.go, ftpsync/background_test.go, .planning/phases/07-background-disk-ftp-lifecycle/07-01-SUMMARY.md]
  modified: [ftpsync/service.go, ftpsync/context_test.go]

key-decisions:
  - "Kept StartBackground available only for DirectionLocalToFTP and preserved ErrUnsupportedCapability method/direction context for FTP→local background attempts."
  - "Added Wait() to the public Handle interface so embedders have explicit wait-for-exit semantics instead of polling Done/Err manually."
  - "Implemented initial background startup as a dedicated catch-up step that reuses executeSyncOnce and records failures on the handle before readiness."

patterns-established:
  - "Background runners should reuse ftpsync one-shot orchestration for transfer passes rather than creating another FTP mutation path."
  - "Readiness is an internal testable barrier that is reached only after initial catch-up completes."

requirements-completed: [WATCH-01, WATCH-03, WATCH-05]

# Metrics
duration: 4min
completed: 2026-04-27
---

# Phase 07 Plan 01: Background Lifecycle Entry Point and Handle Summary

**FTPSyncService now starts supported local disk→FTP background runs with a waitable lifecycle handle and an initial catch-up sync before readiness.**

## Performance

- **Duration:** 4 min
- **Started:** 2026-04-27T08:41:32Z
- **Completed:** 2026-04-27T08:45:54Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Replaced the local→FTP `StartBackground` unsupported stub with validation-preserving dispatch into `executeStartBackground`.
- Added `ftpsync/background.go` with a concrete, mutex-protected lifecycle handle supporting `Done()`, `Err()`, `Wait()`, and `Stop(context.Context)`.
- Reused the existing `executeSyncOnce` local→FTP path for startup catch-up and exposed readiness only after that pass finishes.
- Kept FTP→local background mode explicitly unsupported with method and direction context.

## Task Commits

Each task was committed atomically:

1. **Task 1 RED: Add failing background lifecycle contract tests** - `2e5545e` (test)
2. **Task 1 GREEN: Implement background lifecycle handle contract** - `b91a65e` (feat)
3. **Task 2: Isolate initial background catch-up before readiness** - `c494ee7` (feat)

**Plan metadata:** Pending

_Note: Task 1 was TDD and includes separate RED/GREEN commits. Task 2 built on the Task 1 readiness test and added an implementation refactor commit after verification._

## Files Created/Modified

- `ftpsync/background.go` - Adds `executeStartBackground`, concrete `backgroundHandle`, stop/wait/error methods, initial catch-up execution, and internal readiness signaling.
- `ftpsync/background_test.go` - Covers waitable handle behavior, stop/done semantics, and initial sync before readiness.
- `ftpsync/service.go` - Expands the public `Handle` contract with `Wait()` and routes supported local→FTP background runs to `executeStartBackground`.
- `ftpsync/context_test.go` - Updates public contract expectations now that local→FTP background dispatch is implemented.

## Decisions Made

- Kept background support strictly local→FTP for v2.0; FTP→local remains blocked via `ErrUnsupportedCapability` to preserve WATCH-05 and D-11.
- Added `Wait()` directly to the public `Handle` interface because D-09 requires active wait-for-exit semantics for embedders.
- Recorded initial sync errors on the handle instead of making `StartBackground` fail, aligning startup with the long-running service behavior in D-03/D-04.

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None found. The current background runner intentionally stops after initial catch-up and idle wait until later Phase 7 plans add watcher/debounce behavior.

## Issues Encountered

- The shell PATH did not expose `gofmt`; verification used the installed tool at `/usr/local/go/bin/gofmt`.

## Threat Flags

None. No new network endpoints, auth paths, file access trust boundaries, or schema changes beyond the planned lifecycle API surface were introduced.

## TDD Gate Compliance

- RED gate: `2e5545e` added failing background lifecycle tests before implementation.
- GREEN gate: `b91a65e` implemented the lifecycle contract after the RED failure.
- REFACTOR gate: `c494ee7` isolated initial catch-up before readiness while keeping verification green.

## Verification

- `go test ./ftpsync -run 'Test(StartBackgroundChecksValidationAndContext|StartBackgroundRejectsFTPToLocal|BackgroundHandleWait)' -count=1` — passed.
- `go test ./ftpsync -run 'TestStartBackgroundInitialSyncBeforeReady' -count=1` — passed.
- `go test ./ftpsync -run 'Test(StartBackgroundChecksValidationAndContext|StartBackgroundRejectsFTPToLocal|BackgroundHandleWait|StartBackgroundInitialSyncBeforeReady)' -count=1` — passed.
- `go test ./ftpsync -count=1` — passed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 07-02 can add watcher/debounce/coalescing behavior on top of the waitable handle and initial catch-up sequence.
- Plan 07-03 can harden shutdown/error lifecycle semantics with the same handle contract without changing the public entrypoint.

## Self-Check: PASSED

- Found summary file: `.planning/phases/07-background-disk-ftp-lifecycle/07-01-SUMMARY.md`
- Found created/modified files: `ftpsync/background.go`, `ftpsync/background_test.go`, `ftpsync/service.go`, `ftpsync/context_test.go`
- Found task commits: `2e5545e`, `b91a65e`, `c494ee7`
- Verified plan acceptance criteria and plan verification commands passed in this environment.

---
*Phase: 07-background-disk-ftp-lifecycle*
*Completed: 2026-04-27*
