---
phase: 05-public-ftpsyncservice-api-contract
plan: "01"
subsystem: api
tags: [ftp, ftpsync, public-api, options, go, testing]

# Dependency graph
requires:
  - phase: v1.0-01-01
    provides: FTP endpoint contract semantics for host, credentials, timeout, passive mode, and encoding
  - phase: v1.0-02-01
    provides: FTP driver behavior that future ftpsync execution plans will wrap
provides:
  - Importable `github.com/no-src/gofs/ftpsync` package boundary
  - Typed public `Options`, endpoint, FTP, retry, direction, and ignore-rule configuration contract
  - `FTPSyncService` constructor with private option copy and dependency-boundary coverage
affects: [phase-06, phase-07, phase-08, public-api, docs]

# Tech tracking
tech-stack:
  added: []
  patterns: ["Public ftpsync API starts with typed Go values only and no legacy runtime package imports", "TDD RED/GREEN commits for public option and constructor contracts"]

key-files:
  created: [ftpsync/doc.go, ftpsync/options.go, ftpsync/service.go, ftpsync/options_test.go, .planning/phases/05-public-ftpsyncservice-api-contract/05-01-SUMMARY.md]
  modified: []

key-decisions:
  - "Represented sync direction as string-backed `Direction` constants for local→FTP and FTP→local only."
  - "Kept `FTPSyncService` state private and copied ignore-rule slices so callers cannot mutate service-local slice storage after construction."
  - "Constructor errors use a generic sentinel message to avoid leaking FTP passwords or other sensitive option values."

patterns-established:
  - "External-style `ftpsync_test` package tests prove library consumers can import and construct from public API only."
  - "Dependency-boundary tests use `go list -deps` to guard against CLI/server/daemon/API/report/SFTP/MinIO runtime imports."

requirements-completed: [API-01, API-02]

# Metrics
duration: 3min
completed: 2026-04-27
---

# Phase 05 Plan 01: Public FTPSyncService API Contract Summary

**Typed `ftpsync` package contract with public FTP options and a dependency-isolated `FTPSyncService` constructor.**

## Performance

- **Duration:** 3 min
- **Started:** 2026-04-27T03:43:36Z
- **Completed:** 2026-04-27T03:46:05Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Created the public `ftpsync` package with package documentation and typed option structs for local path, FTP connection, remote path, passive mode, timeout, path encoding, retry behavior, and ignore rules.
- Added `FTPSyncService` plus `NewFTPSyncService(opts Options)` without importing legacy CLI/config/server/daemon/API/monitor/report or SFTP/MinIO runtime packages.
- Added external-package tests covering typed option construction, service construction, password-safe errors, and dependency-boundary enforcement.

## Task Commits

Each TDD task was committed atomically:

1. **Task 1 RED: Define the public typed option surface tests** - `e86c1a9` (test)
2. **Task 1 GREEN: Define the public typed option surface** - `e020c9e` (feat)
3. **Task 2 RED: Add FTPSyncService construction tests** - `7676fda` (test)
4. **Task 2 GREEN: Add FTPSyncService construction without legacy runtime imports** - `352e66d` (feat)

**Plan metadata:** Pending

## Files Created/Modified

- `ftpsync/doc.go` - Declares and documents the importable public package boundary.
- `ftpsync/options.go` - Defines `Direction`, `Endpoint`, `FTPOptions`, `RetryOptions`, `IgnoreRule`, and `Options` using typed Go values only.
- `ftpsync/service.go` - Adds `FTPSyncService`, constructor validation, and private option-copy behavior.
- `ftpsync/options_test.go` - Covers external import style, typed option expressiveness, constructor behavior, password-safe errors, and dependency boundary.
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-01-SUMMARY.md` - Records execution results and verification status.

## Decisions Made

- Used string-backed direction and ignore-kind constants so the public API is readable and stable while remaining independent of `core.VFS` and parser grammar.
- Added minimal constructor validation for complete local→FTP and FTP→local typed endpoint sets; deeper structured validation/error contracts remain assigned to Plan 02.
- Returned generic constructor validation errors so configured FTP passwords are not included in error messages.

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None. Empty-string checks in `service.go` are validation guards, not UI/data stubs.

## Issues Encountered

- `gofmt` was not available on PATH in this environment, so formatting was performed with `go fmt ./ftpsync` instead.

## TDD Gate Compliance

- RED gate: Covered by `e86c1a9` and `7676fda` failing test commits.
- GREEN gate: Covered by `e020c9e` and `352e66d` implementation commits.
- REFACTOR gate: Not needed.

## Verification

- `go test ./ftpsync -run TestOptions -count=1` — passed.
- `go test ./ftpsync -run 'Test(NewFTPSyncService|PackageDependencyBoundary)' -count=1` — passed.
- `go test ./ftpsync -count=1` — passed.
- `go list -deps ./ftpsync` with forbidden dependency check for `/cmd`, `/conf`, `/flag`, `/server`, `/daemon`, `/api`, `/monitor`, `/report`, `/driver/sftp`, and `/driver/minio` — passed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 02 can build explicit validation, structured errors, and context-aware public methods on top of the private `FTPSyncService` option copy.
- Plan 03 can replace reserved hook slots with final no-op logging, progress, and sync-event contracts without introducing global runtime reporting dependencies.

## Self-Check: PASSED

- Found summary file: `.planning/phases/05-public-ftpsyncservice-api-contract/05-01-SUMMARY.md`
- Found created files: `ftpsync/doc.go`, `ftpsync/options.go`, `ftpsync/service.go`, `ftpsync/options_test.go`
- Found task commits: `e86c1a9`, `e020c9e`, `7676fda`, `352e66d`
- Verified task acceptance criteria and plan verification commands passed in this environment.

---
*Phase: 05-public-ftpsyncservice-api-contract*
*Completed: 2026-04-27*
