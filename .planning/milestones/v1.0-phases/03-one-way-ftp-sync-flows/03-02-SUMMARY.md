---
phase: 03-one-way-ftp-sync-flows
plan: "02"
subsystem: testing
tags: [ftp, sync, no-op, metadata, go, regression]

# Dependency graph
requires:
  - phase: 03-01
    provides: FTP source monitoring now uses a real driver-backed polling monitor instead of a deferred placeholder
provides:
  - Targeted FTP flow tests now prove disk→FTP and FTP→disk one-way semantics stay on the generic driver-backed sync paths
  - FTP pull no-op comparisons now prefer FTP driver file times when available so ambiguous listing metadata cannot hide real changes
  - Regression coverage locks down conservative second-run behavior: precise metadata may no-op, ambiguous metadata must safely rewrite
affects: [sync, ftp-sync, testing, roadmap]

# Tech tracking
tech-stack:
  added: []
  patterns: ["FTP-specific sync behavior stays thin: targeted tests exercise generic driver push/pull helpers, and FTP pull comparison prefers driver metadata without adding a new orchestration type"]

key-files:
  created: [.planning/phases/03-one-way-ftp-sync-flows/03-02-SUMMARY.md]
  modified: [sync/sync_test.go, sync/driver_pull_client_sync.go]

key-decisions:
  - "FTP one-way flow regression coverage should stay in package-level sync tests with narrow fakes instead of introducing a live FTP test server in Phase 3."
  - "FTP pull quick-compare logic should prefer driver GetFileTime metadata when available so coarse listing times cannot cause a false no-op skip."

patterns-established:
  - "FTP semantic regression tests should assert routing plus delete/rename behavior through existing driverPushClientSync and driverPullClientSync seams."
  - "When FTP metadata fidelity is uncertain, the sync layer should accept extra transfer work rather than risk missing a real remote change."

requirements-completed: [SYNC-01, SYNC-02, SYNC-03, SYNC-04]

# Metrics
duration: 1min
completed: 2026-04-24
---

# Phase 3 Plan 02: FTP Flow Semantics Summary

**FTP one-way sync semantics are now locked down with regression tests for routing, delete/rename behavior, and conservative no-op checks that use precise FTP file times when available.**

## Performance

- **Duration:** 1 min
- **Started:** 2026-04-24T04:13:33Z
- **Completed:** 2026-04-24T04:14:32Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Added focused FTP sync tests that prove `disk→FTP` and `FTP→disk` stay on the existing driver-backed one-way semantics.
- Added regression coverage for second-run behavior so precise metadata can no-op while ambiguous metadata must still rewrite safely.
- Fixed the FTP pull comparison path to use `GetFileTime` when available, preventing false no-op skips caused by coarse listing timestamps.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add targeted FTP flow tests for one-way semantics and conservative no-op behavior** - `dcbaa3e` (test)
2. **Task 2: Apply only the minimal sync-layer fixes required by the new FTP flow tests** - `1bc1d0a` (feat)

**Plan metadata:** Pending

## Files Created/Modified
- `sync/sync_test.go` - Adds targeted FTP routing, delete/rename, supported-metadata no-op, and ambiguous-metadata rewrite regression tests using narrow fakes.
- `sync/driver_pull_client_sync.go` - Prefers FTP driver file times during quick comparison so remote metadata ambiguity cannot suppress required rewrites.
- `.planning/phases/03-one-way-ftp-sync-flows/03-02-SUMMARY.md` - Records plan outcome, decisions, deviations, and verification results.

## Decisions Made
- Kept Phase 3 verification inside `sync` package tests with lightweight fakes, preserving the plan boundary that reserves realistic FTP-server fixtures for Phase 4.
- Fixed the semantic gap inside the existing `driverPullClientSync` seam instead of adding FTP-specific orchestration, preserving the repository’s generic sync architecture.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed FTP pull false no-op decisions under ambiguous metadata**
- **Found during:** Task 2 (Apply only the minimal sync-layer fixes required by the new FTP flow tests)
- **Issue:** `driverPullClientSync.write` compared destination files against `sourceFile.Stat().ModTime()` only, so coarse FTP listing timestamps could falsely match local files and skip a required rewrite.
- **Fix:** Preferred `getFileTimeFn(path)` metadata before the quick-compare check, while keeping the existing generic driver-backed pull path intact.
- **Files modified:** `sync/driver_pull_client_sync.go`
- **Verification:** `go test ./sync -count=1` and `go test ./driver/ftp ./monitor ./sync -count=1`
- **Committed in:** `1bc1d0a` (task commit)

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** The fix was required for correctness and threat-model compliance. No scope creep and no new FTP-specific sync engine was introduced.

## Issues Encountered
- The new ambiguous-metadata regression test failed immediately, revealing a real sync-layer bug rather than just a missing assertion. The fix stayed limited to the existing FTP pull comparison seam.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 4 can now build realistic FTP-server verification on top of stable one-way semantics already enforced by package-level regression tests.
- State tracking should now mark all Phase 3 sync-flow requirements complete.

## Self-Check: PASSED

- Found summary file: `.planning/phases/03-one-way-ftp-sync-flows/03-02-SUMMARY.md`
- Found commit: `dcbaa3e`
- Found commit: `1bc1d0a`
- Verified `go test ./sync -count=1` and `go test ./driver/ftp ./monitor ./sync -count=1` pass.
