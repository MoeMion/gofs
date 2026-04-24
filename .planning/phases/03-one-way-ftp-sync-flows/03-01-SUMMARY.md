---
phase: 03-one-way-ftp-sync-flows
plan: "01"
subsystem: testing
tags: [ftp, monitor, polling, go, regression]

# Dependency graph
requires:
  - phase: 02-02
    provides: FTP sync constructors wired to the real driver-backed pull and push paths
provides:
  - FTP source monitoring now uses a real driver-backed pull monitor instead of a deferred placeholder
  - FTP source startup now fails explicitly unless sync_once or sync_cron is configured
  - Regression tests cover FTP monitor routing plus truthful sync_once, sync_cron, and idle-start behavior
affects: [monitor, ftp-sync, roadmap, testing]

# Tech tracking
tech-stack:
  added: []
  patterns: ["FTP source monitors stay as thin wrappers over driverPullClientMonitor, with protocol-specific startup gating at the monitor boundary"]

key-files:
  created: [.planning/phases/03-one-way-ftp-sync-flows/03-01-SUMMARY.md]
  modified: [monitor/ftp_pull_client_monitor.go, monitor/monitor_test.go]

key-decisions:
  - "FTP source startup is rejected unless sync_once or sync_cron is configured, so long-running FTP mode never appears healthy while doing nothing."
  - "FTP source monitoring reuses driverPullClientMonitor directly and adds only protocol-specific startup gating instead of introducing new FTP orchestration logic."

patterns-established:
  - "Protocol-specific pull monitors should remain thin wrappers that delegate sync execution to driverPullClientMonitor."
  - "Backends without event streams should enforce truthful startup checks at Start() rather than silently idling."

requirements-completed: [SYNC-02, SYNC-03]

# Metrics
duration: 2min
completed: 2026-04-24
---

# Phase 3 Plan 01: FTP Source Polling Monitor Summary

**FTP→disk source monitoring now runs through a real driver-backed polling monitor with explicit startup rejection when neither sync_once nor sync_cron is configured.**

## Performance

- **Duration:** 2 min
- **Started:** 2026-04-24T04:04:54Z
- **Completed:** 2026-04-24T04:06:40Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Replaced the FTP monitor placeholder with a real `ftpPullClientMonitor` that embeds `driverPullClientMonitor` and preserves the existing minimal-change monitor architecture.
- Added truthful startup gating so FTP source mode only starts in supported runtime shapes: `sync_once` or cron polling.
- Expanded monitor regression coverage to prove FTP routing works and that supported and unsupported startup paths behave explicitly.

## Task Commits

Each task was committed atomically:

1. **Task 1: Replace the FTP pull monitor placeholder with a real polling monitor** - `7404301` (feat)
2. **Task 2: Add monitor regression tests for FTP source routing and truthful startup behavior** - `0c7035c` (test)

**Plan metadata:** Pending

## Files Created/Modified
- `monitor/ftp_pull_client_monitor.go` - Replaces the deferred FTP monitor with a real thin wrapper over `driverPullClientMonitor` and explicit unsupported-startup rejection.
- `monitor/monitor_test.go` - Asserts FTP routing returns `*ftpPullClientMonitor` and covers `sync_once`, `sync_cron`, and idle-start rejection behavior.
- `.planning/phases/03-one-way-ftp-sync-flows/03-01-SUMMARY.md` - Captures execution outcome, decisions, deviations, and verification for this plan.

## Decisions Made
- Rejected FTP source startup without `-sync_once` or `-sync_cron` to satisfy the phase requirement for truthful long-running behavior.
- Kept FTP-specific behavior at the monitor boundary by overriding `Start()` and delegating supported execution paths to `driverPullClientMonitor.Start()`.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Updated stale FTP monitor regression while implementing Task 1**
- **Found during:** Task 1 (Replace the FTP pull monitor placeholder with a real polling monitor)
- **Issue:** Existing `monitor/monitor_test.go` still asserted the removed deferred-placeholder error, which blocked package verification immediately after the monitor implementation landed.
- **Fix:** Rewrote the routing assertion to expect successful FTP routing into `*ftpPullClientMonitor` instead of the old deferred error.
- **Files modified:** `monitor/monitor_test.go`
- **Verification:** `go test ./monitor -count=1`
- **Committed in:** `7404301` (part of task commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** The deviation was required to keep verification aligned with the real Phase 3 contract. No scope creep.

## Issues Encountered
- Task 2 was marked `tdd="true"`, but the plan order implemented the FTP monitor behavior in Task 1 first. As a result, the newly added Task 2 tests passed on their first execution instead of producing a natural RED failure.

## TDD Gate Compliance
- RED gate: Not achieved as a failing-test commit, because the feature was intentionally implemented in Task 1 before the Task 2 test task began.
- GREEN gate: Behavior was already present in `7404301` before the test-only commit `0c7035c`.
- Impact: No functional gap remains, but this task sequence does not represent a strict RED → GREEN TDD cycle.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 3 Plan 2 can now focus on flow semantics and conservative no-op behavior without revisiting FTP source monitor startup truthfulness.
- FTP source mode now has an explicit failure path for unsupported idle runtime configuration, reducing the risk of silent false-success behavior.

## Self-Check: PASSED

- Found summary file: `.planning/phases/03-one-way-ftp-sync-flows/03-01-SUMMARY.md`
- Found commit: `7404301`
- Found commit: `0c7035c`
- Verified `go test ./monitor -count=1` and `go test ./driver/ftp ./monitor ./sync -count=1` pass.
