---
phase: 02-ftp-driver-backend
plan: "02"
subsystem: testing
tags: [ftp, sync, driver, go, routing, regression]

# Dependency graph
requires:
  - phase: 02-01
    provides: FTP driver implementation and bounded reconnect behavior behind driver.Driver
provides:
  - FTP push and pull sync constructors wired to the real driver-backed sync flows
  - FTP pull metadata comparison routed through driver Stat and GetFileTime behavior
  - Regression tests proving FTP paths no longer stop at deferred backend errors
affects: [sync, ftp-driver, testing, roadmap]

# Tech tracking
tech-stack:
  added: []
  patterns: ["FTP sync entry points follow the same generic driverPushClientSync and driverPullClientSync constructor pattern already used by SFTP and MinIO, with package-local factory seams for constructor-only tests"]

key-files:
  created: [.planning/phases/02-ftp-driver-backend/02-02-SUMMARY.md]
  modified: [sync/ftp_push_client_sync.go, sync/ftp_pull_client_sync.go, sync/sync_test.go]

key-decisions:
  - "FTP sync constructors now instantiate ftp.NewFTPDriver through thin package-local factories so production wiring stays direct while tests can assert routing without opening network connections."
  - "FTP pull sync resets sourceAbsPath, statFn, and getFileTimeFn after startup so generic pull comparison logic uses the FTP driver metadata policy instead of disk defaults."

patterns-established:
  - "Protocol-specific constructor tests can use unexported driver factory seams to validate routing without introducing broad integration harnesses."
  - "FTP sync behavior reuses existing generic driver-backed push and pull sync structs instead of adding protocol-specific sync semantics."

requirements-completed: [FTPD-01, FTPD-02, FTPD-03, FTPD-04, FTPD-05, FTPD-06, FTPD-07, FTPD-08, FTPD-09]

# Metrics
duration: 12min
completed: 2026-04-24
---

# Phase 2 Plan 02: FTP Driver Backend Summary

**FTP disk↔endpoint sync constructors now route through the real driver-backed sync path with deterministic regression coverage instead of Phase 1 deferred placeholder errors.**

## Performance

- **Duration:** 12 min
- **Started:** 2026-04-24T02:18:09Z
- **Completed:** 2026-04-24T02:30:13Z
- **Tasks:** 2
- **Files modified:** 4

## Accomplishments
- Replaced the FTP push and pull placeholder constructors with real driver-backed constructors that mirror the existing SFTP and MinIO wiring shape.
- Wired FTP pull sync metadata comparison to the FTP driver's `Stat` and `GetFileTime` behavior so the generic pull path uses the backend's conservative metadata policy.
- Rewrote sync routing regression tests to prove FTP routes into real constructor types without requiring a live FTP server and kept unsupported non-FTP combinations on the existing unsupported fallback.

## Task Commits

Each task was committed atomically:

1. **Task 1: Replace FTP deferred constructors with real driver-backed sync constructors** - `cca3845` (feat)
2. **Task 2: Update sync regression tests to prove FTP routes into the real backend path** - `07a681e` (test)

**Plan metadata:** Pending

## Files Created/Modified
- `sync/ftp_push_client_sync.go` - Replaces the deferred FTP push constructor with `driverPushClientSync` wiring and direct `ftp.NewFTPDriver` startup.
- `sync/ftp_pull_client_sync.go` - Replaces the deferred FTP pull constructor with `driverPullClientSync` wiring plus FTP-specific metadata hook resets.
- `sync/sync_test.go` - Replaces deferred-error assertions with constructor-routing tests backed by package-local fake driver factories.
- `.planning/phases/02-ftp-driver-backend/02-02-SUMMARY.md` - Records plan outcomes, decisions, verification, and readiness context.

## Decisions Made
- Added package-local unexported driver factory seams in the FTP sync constructors so tests can assert routing deterministically without changing production driver selection.
- Kept FTP sync integration strictly inside existing generic driver-backed push and pull flows rather than adding FTP-specific sync logic.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- The environment did not expose `gofmt` on `PATH`, so formatting had to use `$(go env GOROOT)/bin/gofmt` before verification.

## TDD Gate Compliance
- RED gate: Not produced as a separate failing-test commit. The plan's TDD-marked task was scoped to rewriting regression tests after constructor implementation, so execution completed as a single test-only task commit.
- GREEN gate: Not applicable within Task 2 because implementation shipped in Task 1 per the plan order.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 2 is complete: FTP now has a real driver package and sync-layer constructor wiring.
- Phase 3 can focus on one-way FTP sync behavior and verification of actual file transfer semantics without revisiting constructor routing.
- Existing environment-specific baseline noise in unrelated files (`.gitignore`, `AGENTS.md`, `encrypt/testdata/`) remains untouched and out of scope for this plan.

## Self-Check: PASSED

- Found summary file: `.planning/phases/02-ftp-driver-backend/02-02-SUMMARY.md`
- Found commit: `cca3845`
- Found commit: `07a681e`
- Verified `go test ./sync -count=1` and `go test ./driver/ftp ./sync -count=1` pass.
