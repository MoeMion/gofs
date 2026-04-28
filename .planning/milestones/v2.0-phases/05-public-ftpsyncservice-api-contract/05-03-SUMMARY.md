---
phase: 05-public-ftpsyncservice-api-contract
plan: "03"
subsystem: api
tags: [ftp, ftpsync, public-api, hooks, observability, go, testing]

# Dependency graph
requires:
  - phase: 05-01
    provides: Typed `ftpsync` options and `FTPSyncService` constructor boundary
  - phase: 05-02
    provides: Structured public error kinds plus context-aware public method contracts
provides:
  - Public library-local hook contracts for logging, progress snapshots, and sync events
  - Service-owned hook normalization with no-op defaults
  - Dependency-boundary coverage proving hooks do not import legacy logger/report/web runtime packages
affects: [phase-06, phase-07, phase-08, public-api, observability]

# Tech tracking
tech-stack:
  added: []
  patterns: ["Synchronous caller-provided hooks are normalized to no-op callbacks inside FTPSyncService", "Hook payloads are typed public ftpsync values and avoid FTP credentials"]

key-files:
  created: [ftpsync/hooks.go, ftpsync/hooks_test.go, .planning/phases/05-public-ftpsyncservice-api-contract/05-03-SUMMARY.md]
  modified: [ftpsync/options.go, ftpsync/service.go]

key-decisions:
  - "Added `HookSet` to public `Options` instead of retaining separate reserved hook function fields, so logging, progress, and event callbacks have one cohesive optional configuration surface."
  - "Kept hook execution synchronous and library-local with no goroutines, queues, global loggers, report.Reporter, eventlog.Event, or web report dependencies."
  - "Normalized omitted hooks to no-op callbacks and made service dispatch helpers safe even for zero-value service state."

patterns-established:
  - "Public hook types use only standard-library-compatible values plus ftpsync package types."
  - "Package-local tests exercise private dispatch helpers while dependency-boundary tests guard the public package import graph."

requirements-completed: [API-05]

# Metrics
duration: 3min
completed: 2026-04-27
---

# Phase 05 Plan 03: Public Hook Contracts Summary

**Library-local logging, progress, and sync-event hooks with service-owned no-op defaults.**

## Performance

- **Duration:** 3 min
- **Started:** 2026-04-27T03:53:19Z
- **Completed:** 2026-04-27T03:56:11Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Added `ftpsync/hooks.go` with public `Logger`, `LoggerFunc`, `Progress`, `ProgressHook`, `SyncEvent`, `EventHook`, and `HookSet` contracts.
- Replaced the reserved per-hook function fields in `Options` with a cohesive `Hooks HookSet` configuration field.
- Extended `FTPSyncService` to normalize omitted hooks to no-op callbacks and dispatch log/progress/event notifications through private service helpers.
- Added tests proving hook contracts compile, zero values normalize safely, custom hooks receive typed payloads, passwords are not emitted by helper dispatches, and forbidden legacy runtime dependencies remain outside `ftpsync`.

## Task Commits

Each TDD task was committed atomically:

1. **Task 1 RED: Define library-local hook contract tests** - `c6c39c2` (test)
2. **Task 1 GREEN: Define library-local hook contracts** - `31458de` (feat)
3. **Task 2 RED: Add zero-value no-op dispatch coverage** - `c452675` (test)
4. **Task 2 GREEN: Normalize service hook dispatch defaults** - `0a08e7a` (feat)

**Plan metadata:** Pending final metadata commit.

## Files Created/Modified

- `ftpsync/hooks.go` - Defines public hook contracts and no-op logger implementation using no legacy runtime imports.
- `ftpsync/hooks_test.go` - Covers public hook contracts, no-op/default behavior, custom callback dispatch, password-safe payloads, and dependency boundaries.
- `ftpsync/options.go` - Adds `Hooks HookSet` to public service options and removes previous reserved hook fields.
- `ftpsync/service.go` - Stores normalized hooks on the service and adds private dispatch helpers for log, progress, and sync events.
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-03-SUMMARY.md` - Records execution results and verification status.

## Decisions Made

- Used a small `Logger` interface plus `LoggerFunc` adapter rather than exposing the existing runtime `logger.Logger` type.
- Modeled progress as a typed `Progress` value with path, byte, and file counters so future Phase 06/07 implementations can report useful snapshots without web reports.
- Modeled sync events as `SyncEvent` with operation/path/status/error-kind metadata and no credential fields.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 2 - Missing critical functionality] Made dispatch helpers safe for zero-value service state**
- **Found during:** Task 2
- **Issue:** Dispatch helpers used the service's stored hooks directly; a zero-value service could panic if package-local helpers were called without constructor normalization.
- **Fix:** Added `normalizedHooks()` so helper dispatches use no-op defaults whenever stored hooks are omitted.
- **Files modified:** `ftpsync/service.go`, `ftpsync/hooks_test.go`
- **Commit:** `0a08e7a`

## Known Stubs

- `ftpsync/service.go` — `SyncOnce` still returns `ErrUnsupportedCapability` until Phase 06 implements one-shot transfer dispatch; inherited intentional stub from Plan 02.
- `ftpsync/service.go` — `StartBackground` still returns `ErrUnsupportedCapability` until Phase 07 implements disk→FTP lifecycle dispatch; inherited intentional stub from Plan 02.

## Threat Flags

None. This plan added caller hook callbacks at the planned `ftpsync→caller hooks` trust boundary only; no new network endpoints, auth paths, file access paths, or schema boundaries were introduced.

## Issues Encountered

None.

## TDD Gate Compliance

- RED gate: Covered by `c6c39c2` and `c452675` failing test commits.
- GREEN gate: Covered by `31458de` and `0a08e7a` implementation commits.
- REFACTOR gate: Not needed.

## Verification

- `go test ./ftpsync -run TestHookContracts -count=1` — passed.
- `go test ./ftpsync -run 'Test(HookDefaults|CustomHooks|PackageDependencyBoundary)' -count=1` — passed.
- `go test ./ftpsync -count=1` — passed.
- `go list -deps ./ftpsync` with forbidden dependency checks for `report`, `eventlog`, `server`, `logger`, `cmd`, `conf`, `flag`, `daemon`, `api`, `monitor`, `driver/sftp`, and `driver/minio` — passed.
- `go test ./... -run TestNonExistent -count=1` — passed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 06 can report one-shot sync log messages, progress snapshots, and sync events through the normalized hook dispatch helpers without changing public API shape.
- Phase 07 can reuse the same hook surface for background disk→FTP lifecycle observability while keeping hook execution synchronous and caller-owned.

## Self-Check: PASSED

- Found summary file: `.planning/phases/05-public-ftpsyncservice-api-contract/05-03-SUMMARY.md`
- Found created/modified files: `ftpsync/hooks.go`, `ftpsync/hooks_test.go`, `ftpsync/options.go`, `ftpsync/service.go`
- Found task commits: `c6c39c2`, `31458de`, `c452675`, `0a08e7a`
- Verified task acceptance criteria and plan verification commands passed in this environment.

---
*Phase: 05-public-ftpsyncservice-api-contract*
*Completed: 2026-04-27*
