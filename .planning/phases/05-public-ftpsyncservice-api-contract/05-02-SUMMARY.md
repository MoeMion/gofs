---
phase: 05-public-ftpsyncservice-api-contract
plan: "02"
subsystem: api
tags: [ftp, ftpsync, public-api, validation, context, errors, go, testing]

# Dependency graph
requires:
  - phase: 05-01
    provides: Typed `ftpsync` package options and private `FTPSyncService` constructor boundary
provides:
  - Public structured `ftpsync.Error` taxonomy and `IsKind` classifier for library callers
  - Deterministic validation for local→FTP and FTP→local endpoint role combinations
  - Context-aware `SyncOnce` and `StartBackground` API contracts with library-local result and handle types
affects: [phase-06, phase-07, phase-08, public-api, validation, error-contracts]

# Tech tracking
tech-stack:
  added: []
  patterns: ["Public ftpsync errors use standard errors.As/errors.Is wrapping with explicit ErrorKind classification", "Public methods validate options and context before returning unsupported capability stubs for later implementation phases"]

key-files:
  created: [ftpsync/errors.go, ftpsync/context_test.go, ftpsync/validation_test.go, .planning/phases/05-public-ftpsyncservice-api-contract/05-02-SUMMARY.md]
  modified: [ftpsync/service.go]

key-decisions:
  - "Represented public failure classes with string-backed `ErrorKind` constants and a structured `*ftpsync.Error` wrapper compatible with `errors.As`, `errors.Is`, and `Unwrap`."
  - "Required positive FTP ports during validation so caller configuration is complete before transfer work can start."
  - "Kept `SyncOnce` and `StartBackground` as context-aware public contracts that currently return `ErrUnsupportedCapability` after validation until later execution phases provide transfer and lifecycle implementations."

patterns-established:
  - "Validation rejects ambiguous endpoint roles before method dispatch and never includes FTP password values in error messages."
  - "Cancellation is mapped to `ErrCanceled` while preserving `errors.Is(err, context.Canceled)` compatibility."

requirements-completed: [API-03, API-04]

# Metrics
duration: 3min
completed: 2026-04-27
---

# Phase 05 Plan 02: Public Validation and Context/Error Contracts Summary

**Structured `FTPSyncService` validation plus context-aware public methods with classified, password-safe errors.**

## Performance

- **Duration:** 3 min
- **Started:** 2026-04-27T03:48:12Z
- **Completed:** 2026-04-27T03:51:19Z
- **Tasks:** 3
- **Files modified:** 4

## Accomplishments

- Added public `ErrorKind`, `Error`, and `IsKind` contracts covering validation, cancellation, connection, authentication, transfer, and unsupported-capability failures while preserving wrapped causes.
- Expanded `FTPSyncService` validation to accept only local→FTP and FTP→local endpoint roles and reject unsupported, missing-field, invalid-port, ambiguous, and mismatched combinations before any transfer work.
- Added `SyncOnce(ctx)` and `StartBackground(ctx)` public method contracts with library-local `Result` and `Handle` types, cancellation mapping, and explicit unsupported-capability returns for later implementation phases.

## Task Commits

Each TDD task was committed atomically:

1. **Task 1 RED: Implement public error taxonomy and classifiers tests** - `54d6123` (test)
2. **Task 1 GREEN: Implement public error taxonomy and classifiers** - `7a91264` (feat)
3. **Task 2 RED: Validate supported directions and reject unsupported endpoint combinations tests** - `f6e0a7a` (test)
4. **Task 2 GREEN: Validate supported directions and reject unsupported endpoint combinations** - `16ef674` (feat)
5. **Task 3 RED: Add context-aware public sync method contract tests** - `c456bba` (test)
6. **Task 3 GREEN: Add context-aware public sync method contracts** - `79cafdd` (feat)

**Plan metadata:** Pending

## Files Created/Modified

- `ftpsync/errors.go` - Defines public error kinds, structured error wrapper, `Unwrap`, `Kind`, and `IsKind` classification helper.
- `ftpsync/service.go` - Adds `Validate`, `SyncOnce`, `StartBackground`, `Result`, `Handle`, endpoint-role validation helpers, context checks, and unsupported capability returns.
- `ftpsync/context_test.go` - Covers error classification/wrapping, cancellation classification, context-aware method ordering, and public contract compilation.
- `ftpsync/validation_test.go` - Covers accepted directions, unsupported combinations, missing fields, invalid ports, ambiguous endpoints, mismatched direction roles, and password-safe errors.

## Decisions Made

- Used string-backed error kinds to keep public error classification readable and stable for library consumers.
- Required `Port > 0` for FTP endpoints instead of silently defaulting during validation, because API-03 requires explicit invalid-port rejection before transfer work.
- Returned unsupported capability errors from `SyncOnce` and `StartBackground` after validation/cancellation checks so Phase 06/07 can fill implementation without changing signatures.

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

- `ftpsync/service.go` — `SyncOnce` returns `ErrUnsupportedCapability` until Phase 06 implements one-shot transfer dispatch; this is intentional per Plan 02.
- `ftpsync/service.go` — `StartBackground` returns `ErrUnsupportedCapability` until Phase 07 implements disk→FTP lifecycle dispatch; this is intentional per Plan 02.

## Issues Encountered

None.

## TDD Gate Compliance

- RED gate: Covered by `54d6123`, `f6e0a7a`, and `c456bba` failing test commits.
- GREEN gate: Covered by `7a91264`, `16ef674`, and `79cafdd` implementation commits.
- REFACTOR gate: Not needed.

## Verification

- `go test ./ftpsync -run 'Test(ErrorKind|ErrorWrapping|ContextCancellation)' -count=1` — passed.
- `go test ./ftpsync -run TestValidate -count=1` — passed.
- `go test ./ftpsync -run 'Test(SyncOnce|StartBackground|Context)' -count=1` — passed.
- `go test ./ftpsync -count=1` — passed.
- `go test ./... -run TestNonExistent -count=1` — passed.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 03 can add no-op logging, progress, and sync-event hooks on top of the validated service and public method contracts.
- Phase 06 can replace `SyncOnce` unsupported-capability returns with one-shot transfer dispatch while preserving the validation and cancellation ordering established here.
- Phase 07 can replace the local→FTP `StartBackground` unsupported-capability return with lifecycle implementation while keeping FTP→local background unsupported.

## Self-Check: PASSED

- Found summary file: `.planning/phases/05-public-ftpsyncservice-api-contract/05-02-SUMMARY.md`
- Found created/modified files: `ftpsync/errors.go`, `ftpsync/service.go`, `ftpsync/context_test.go`, `ftpsync/validation_test.go`
- Found task commits: `54d6123`, `7a91264`, `f6e0a7a`, `16ef674`, `c456bba`, `79cafdd`
- Verified task acceptance criteria and plan verification commands passed in this environment.

---
*Phase: 05-public-ftpsyncservice-api-contract*
*Completed: 2026-04-27*
