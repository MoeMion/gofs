---
phase: 01-ftp-endpoint-contract-routing
plan: "02"
subsystem: sync
tags: [ftp, sync, monitor, routing, go, testing]

# Dependency graph
requires:
  - phase: 01-01
    provides: FTP-aware VFS classification, ftp query parsing, and default-port handling
provides:
  - FTP-specific sync constructor entry points for disk→FTP and FTP→disk routing
  - FTP-specific monitor constructor entry point for FTP source routing
  - Regression tests proving FTP factory dispatch avoids generic unsupported branches
affects: [monitor, sync, ftp-driver, phase-2]

# Tech tracking
tech-stack:
  added: []
  patterns: ["Add protocol routing with thin constructor placeholders that defer backend behavior to later phases"]

key-files:
  created: [sync/ftp_push_client_sync.go, sync/ftp_pull_client_sync.go, monitor/ftp_pull_client_monitor.go, sync/sync_test.go, monitor/monitor_test.go, .planning/phases/01-ftp-endpoint-contract-routing/01-02-SUMMARY.md]
  modified: [sync/sync.go, monitor/monitor.go]

key-decisions:
  - "FTP factory targets return explicit phase-2-deferred errors so routing becomes testable without introducing premature protocol behavior."
  - "Regression tests assert FTP combinations do not fall back to generic unsupported errors in sync and monitor factories."

patterns-established:
  - "New backend routing is introduced by adding a protocol-specific constructor surface first, then dispatch branches, then focused regression tests."
  - "Phase-1 backend placeholders must avoid concrete protocol libraries and keep failure text explicit but secret-free."

requirements-completed: [FTP-01, FTP-02]

# Metrics
duration: 13min
completed: 2026-04-23
---

# Phase 1 Plan 02: FTP Routing Summary

**FTP sync and monitor factories now route `ftp://` endpoints into explicit FTP constructor placeholders with regression coverage instead of generic unsupported-path fallthrough.**

## Performance

- **Duration:** 13 min
- **Started:** 2026-04-23T09:24:21Z
- **Completed:** 2026-04-23T09:37:41Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- Added thin FTP sync entry points for push and pull flows that intentionally defer backend implementation to Phase 2.
- Added a thin FTP pull monitor entry point so FTP sources route through a defined monitor path.
- Extended sync and monitor factories plus package tests so FTP combinations no longer hit generic unsupported-path logic.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add thin FTP sync and monitor entry points** - `dc0e370` (feat)
2. **Task 2: Route FTP through sync and monitor factories with regression tests** - `3db0d5f` (test)

**Plan metadata:** Pending

## Files Created/Modified
- `sync/ftp_push_client_sync.go` - Adds exported FTP push sync constructor placeholder with a Phase 2 deferred error.
- `sync/ftp_pull_client_sync.go` - Adds exported FTP pull sync constructor placeholder with the shared deferred error.
- `monitor/ftp_pull_client_monitor.go` - Adds exported FTP pull monitor constructor placeholder with an explicit deferred error.
- `sync/sync.go` - Routes disk→FTP and FTP→disk combinations into FTP-specific constructors.
- `monitor/monitor.go` - Routes FTP sources into the FTP-specific pull monitor constructor.
- `sync/sync_test.go` - Verifies FTP sync routing reaches FTP-specific constructors instead of unsupported fallback.
- `monitor/monitor_test.go` - Verifies FTP monitor routing reaches the FTP-specific monitor constructor instead of unsupported fallback.

## Decisions Made
- Used constructor-level placeholder errors rather than partial backend wiring to preserve the architecture boundary that leaves real FTP behavior to Phase 2.
- Kept the new FTP routing adjacent to existing SFTP and MinIO branches so non-FTP execution order and behavior remain unchanged.
- Verified routing behavior via focused factory tests instead of broader end-to-end protocol tests, which belong to later FTP driver and flow phases.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- `go test ./sync ./monitor -count=1` passed after the routing and regression-test changes.
- `go test ./... -count=1` still fails in pre-existing `core` SSH-config-dependent tests (`TestVFS_SSHConfig`, `TestVFS_SSHConfigWithCover`) unrelated to FTP routing, matching the known baseline limitation documented before this plan.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 2 can replace the deferred FTP constructor errors with real FTP driver-backed behavior without changing factory call sites.
- FTP source and destination endpoint classification is now fully wired through the sync and monitor selection seams.
- Full-repo verification remains partially environment-dependent until the existing SSH-config test assumptions in `core` are normalized.

## Self-Check: PASSED

- Found summary file: `.planning/phases/01-ftp-endpoint-contract-routing/01-02-SUMMARY.md`
- Found commit: `dc0e370`
- Found commit: `3db0d5f`
- Verified plan output file exists and both task commits are present in git history.

---
*Phase: 01-ftp-endpoint-contract-routing*
*Completed: 2026-04-23*
