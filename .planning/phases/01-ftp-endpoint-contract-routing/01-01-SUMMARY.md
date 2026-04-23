---
phase: 01-ftp-endpoint-contract-routing
plan: "01"
subsystem: config
tags: [ftp, vfs, parsing, go, testing]

# Dependency graph
requires: []
provides:
  - FTP-aware VFS scheme parsing for ftp:// endpoints
  - FTP-specific endpoint fields for username, password, timeout, and passive mode
  - Automated FTP VFS round-trip and default-port test coverage
affects: [sync, monitor, ftp-driver, roadmap]

# Tech tracking
tech-stack:
  added: []
  patterns: ["Extend core.VFS with scheme-specific endpoint config while preserving existing query-parameter grammar"]

key-files:
  created: [.planning/phases/01-ftp-endpoint-contract-routing/01-01-SUMMARY.md]
  modified: [core/vfs.go, core/vfs_test.go]

key-decisions:
  - "FTP endpoints use FTP-specific query keys instead of reusing SSH field names."
  - "Omitted FTP ports default to 21 inside core.VFS parsing to match existing remote backend defaults."

patterns-established:
  - "VFS remote backend support is introduced by adding a scheme branch in NewVFS and backend-specific query parsing in parse()."
  - "FTP tests reuse the existing canonical fixture plus table-driven round-trip pattern used by other VFS backends."

requirements-completed: [FTP-01, FTP-02, FTP-03, FTP-04]

# Metrics
duration: 34min
completed: 2026-04-23
---

# Phase 1 Plan 01: FTP VFS Contract Summary

**FTP endpoint parsing in core.VFS with dedicated ftp query fields, passive-mode/timeout support, and default-port round-trip coverage.**

## Performance

- **Duration:** 34 min
- **Started:** 2026-04-23T08:38:20Z
- **Completed:** 2026-04-23T09:12:52Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Added `ftp://` recognition to `core.NewVFS` so FTP endpoints classify as `core.FTP` instead of falling through to disk handling.
- Added FTP-specific endpoint storage and getters for username, password, timeout, and passive-mode behavior without reusing SSH config fields.
- Extended VFS tests with canonical FTP source/destination fixtures, round-trip coverage, and omitted-port assertions for default port `21`.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add FTP-specific VFS contract fields and parsing** - `a0f7000` (feat)
2. **Task 2: Add FTP VFS round-trip and default-port tests** - `34dc223` (test)

**Plan metadata:** `4c8e75e` (docs)

## Files Created/Modified
- `core/vfs.go` - Adds FTP scheme parsing, FTP-specific query constants, default port handling, and FTP getters/config storage.
- `core/vfs_test.go` - Adds FTP fixtures and verifies marshal/unmarshal, flag usage, default port, and FTP config isolation from SSH config.
- `.planning/phases/01-ftp-endpoint-contract-routing/01-01-SUMMARY.md` - Records execution results, baseline test limitation, and readiness for the next plan.

## Decisions Made
- Used FTP-specific query keys (`ftp_user`, `ftp_pass`, `ftp_timeout`, `ftp_passive`) so FTP credentials remain semantically distinct from SFTP/SSH settings.
- Kept FTP timeout optional and stored as an endpoint-level string contract rather than introducing a new top-level config object.
- Left `conf.Config` unchanged because `Source` and `Dest` already carry `core.VFS` endpoints directly.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Continued despite baseline core test failures caused by existing SSH-config expectations**
- **Found during:** Task 2 (Add FTP VFS round-trip and default-port tests)
- **Issue:** `go test ./core -count=1` fails in pre-existing `TestVFS_SSHConfig` cases because the environment does not provide the SSH config assumptions those tests expect; the failures are unrelated to the new FTP code path.
- **Fix:** Continued plan execution, completed FTP-specific implementation and coverage, and documented the baseline verification limitation here instead of changing unrelated SSH behavior.
- **Files modified:** `.planning/phases/01-ftp-endpoint-contract-routing/01-01-SUMMARY.md`
- **Verification:** FTP-specific test code was added and `go test ./core -count=1` was re-run, showing failures only in existing SSH-config tests rather than FTP assertions.
- **Committed in:** `4c8e75e` (plan metadata commit)

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** FTP contract and test coverage were completed as planned. Full package-green verification remains limited by unrelated baseline SSH-config expectations in this environment.

## Issues Encountered
- `go test ./core -count=1` is not fully reliable in this environment because existing SSH-config tests expect local SSH config host mappings that are not present. This affected baseline validation before and after the FTP changes.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- `core.VFS` can now carry FTP endpoint contract data for future sync and monitor routing work.
- Phase 01 plan 02 can build on `core.FTP` classification and FTP-specific getters without changing the top-level config model.
- Before broad package verification is treated as authoritative, the existing SSH-config test environment should be normalized or made deterministic.

## Self-Check: PASSED

- Found summary file: `.planning/phases/01-ftp-endpoint-contract-routing/01-01-SUMMARY.md`
- Found commit: `a0f7000`
- Found commit: `34dc223`
- Verified plan output file exists and both task commits are present in git history.

---
*Phase: 01-ftp-endpoint-contract-routing*
*Completed: 2026-04-23*
