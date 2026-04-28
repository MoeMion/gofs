---
phase: 08-ftp-only-package-reduction
plan: 01
subsystem: testing
tags: [go, ftpsync, dependency-graph, public-api, package-reduction]

# Dependency graph
requires:
  - phase: 07-background-disk-ftp-lifecycle
    provides: Background disk-to-FTP lifecycle behavior that Phase 8 must preserve while pruning dependencies.
provides:
  - Dependency blacklist regression test for old runtime and non-FTP protocol imports.
  - Public API reflection guard proving FTPSyncService remains typed-options-only.
affects: [phase-08, phase-09, ftpsync]

# Tech tracking
tech-stack:
  added: []
  patterns: [go-list-dependency-blacklist, reflection-public-api-guard]

key-files:
  created: [ftpsync/dependency_boundary_test.go, ftpsync/public_api_test.go]
  modified: []

key-decisions:
  - "Added dependency and public API tests before extraction so later pruning work has executable safety rails."
  - "Kept the dependency blacklist intentionally failing until Phase 8 removes legacy runtime imports."

patterns-established:
  - "Package boundary guard: tests run `go list -deps ./ftpsync` and report matched forbidden import fragments in a table."
  - "Public API guard: reflection verifies FTPSyncService signatures while go doc scanning rejects legacy public type markers."

requirements-completed: [PRUNE-01, PRUNE-03, PRUNE-04]

# Metrics
duration: 8min
completed: 2026-04-27
---

# Phase 08 Plan 01: Dependency and Public API Guard Summary

**Executable package-reduction guardrails for ftpsync dependency pruning and typed-options-only public API contracts**

## Performance

- **Duration:** 8 min
- **Started:** 2026-04-27T09:56:29Z
- **Completed:** 2026-04-27T10:04:38Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments

- Added `TestPackageDependencyBoundaryRejectsOldRuntime`, a strict `go list -deps ./ftpsync` blacklist test covering old internal runtime packages and third-party runtime modules.
- Added `TestPublicAPIStaysTypedOptionsOnly`, a reflection and public-doc regression test for the approved `FTPSyncService` typed options API.
- Verified the public API guard passes now, while the dependency guard fails with explicit legacy dependency matches that later Phase 8 plans must eliminate.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add dependency blacklist regression test** - `84fe4c4` (test)
2. **Task 2: Add public API leak regression test** - `5a0897e` (test)

**Plan metadata:** pending final metadata commit

_Note: Both plan tasks were TDD guard tasks. The dependency test is intentionally red because this plan creates the safety net before extraction._

## Files Created/Modified

- `ftpsync/dependency_boundary_test.go` - Runs `go list -deps ./ftpsync`, matches old runtime and non-FTP third-party dependency fragments, and reports offending dependency lines in a table.
- `ftpsync/public_api_test.go` - Uses `reflect.TypeOf(NewFTPSyncService)` and method reflection to pin public signatures, plus a `go doc -all` scan to reject legacy public markers such as `Config`, `VFS`, `Server`, and `Task`.

## Decisions Made

- Added executable Phase 8 package reduction guards before extraction: dependency blacklist currently fails on legacy runtime imports, while public API reflection guard passes typed-options-only contracts.
- Allowed the plan-level verification command to remain red for the dependency test because the plan objective explicitly states this guard should expose dependency leaks until later plans remove them.

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

- The shell PATH did not expose `gofmt` by command name in one call, so formatting used the absolute `/usr/local/go/bin/gofmt` path. No code or plan scope changed.
- `go test ./ftpsync -run 'Test(PackageDependencyBoundaryRejectsOldRuntime|PublicAPIStaysTypedOptionsOnly)' -count=1` fails as expected because `ftpsync` still imports legacy runtime chains through the current one-shot adapter. The failure output lists Gin, gRPC, SFTP, MinIO, `core`, `sync`, `logger`, `server`, `api`, and related dependencies for later Phase 8 plans to remove.

## Verification

- `go test ./ftpsync -run TestPackageDependencyBoundaryRejectsOldRuntime -count=1` — expected failure; reports current forbidden dependency leaks.
- `go test ./ftpsync -run TestPublicAPIStaysTypedOptionsOnly -count=1` — passed.
- `go test ./ftpsync -run 'Test(PackageDependencyBoundaryRejectsOldRuntime|PublicAPIStaysTypedOptionsOnly)' -count=1` — expected failure due to dependency blacklist guard.

## Known Stubs

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Phase 8 Plan 02 can now replace the legacy one-shot adapter with package-local FTP internals and use `TestPackageDependencyBoundaryRejectsOldRuntime` as the dependency reduction target.
- The typed-options public API guard should continue passing during import rewrites and deletion work.

## Self-Check: PASSED

- Found `ftpsync/dependency_boundary_test.go`.
- Found `ftpsync/public_api_test.go`.
- Found task commit `84fe4c4`.
- Found task commit `5a0897e`.

---
*Phase: 08-ftp-only-package-reduction*
*Completed: 2026-04-27*
