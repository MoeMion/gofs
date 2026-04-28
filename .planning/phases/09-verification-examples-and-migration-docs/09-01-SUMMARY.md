---
phase: 09-verification-examples-and-migration-docs
plan: 01
subsystem: module-contract-verification
tags: [verification, module-path, dependency-boundary]
requires: []
provides: [module-ftpsync, executable-coverage-checklist, dependency-boundary-guard]
affects: [go.mod, ftpsync-tests]
tech-stack:
  added: []
  patterns: [go-test, external-package-tests, dependency-boundary]
key-files:
  created:
    - ftpsync/coverage_checklist_test.go
  modified:
    - go.mod
    - ftpsync/validation_test.go
    - ftpsync/context_test.go
    - ftpsync/options_test.go
    - ftpsync/dependency_boundary_test.go
key-decisions:
  - Kept package source in ftpsync/ and documented/verified the resulting import path as ftpsync/ftpsync under module ftpsync.
requirements-completed: [VERIFY-01, VERIFY-03]
metrics:
  duration: 20 min
  completed: 2026-04-28
---

# Phase 09 Plan 01: Module Contract and Verification Checklist Summary

Updated the local module contract to `module ftpsync` while preserving the current `ftpsync/` subdirectory package layout, then added executable guards for final verification coverage and dependency boundaries.

## What Changed

- Changed the module path to `module ftpsync`.
- Updated external package tests to import `ftpsync/ftpsync`, which is the correct Go import path for package source in the `ftpsync/` subdirectory.
- Added `TestVerificationCoverageChecklist` to ensure required public API, one-shot, background, cwd-safety, encoding, passive-mode, cancellation, and shutdown coverage targets remain discoverable.
- Strengthened dependency boundary tests to assert `go list ./...` exposes exactly `ftpsync/ftpsync` and that old runtime/non-FTP dependencies stay absent.

## Tasks Completed

| Task | Description | Commit |
|------|-------------|--------|
| 1 | Aligned module path, external imports, and Phase 9 plans after resolving the layout/import contradiction. | 4c9a64e |
| 2-3 | Added coverage checklist and dependency boundary assertions. | c5a2e4a |

## Verification

| Command | Result |
|---------|--------|
| `go test ./ftpsync -run 'TestValidateAcceptsSupportedDirections\|TestContextAwarePublicContractsCompile' -count=1` | PASS |
| `go test ./ftpsync -run TestVerificationCoverageChecklist -count=1` | PASS |
| `go test ./ftpsync -run 'TestPackageDependencyBoundaryRejectsOldRuntime' -count=1` | PASS |
| `go test ./...` | PASS |
| `go list ./...` | PASS — printed `ftpsync/ftpsync` |

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 4 - Architectural Decision] Resolved impossible import-path requirements**
- **Found during:** Task 1
- **Issue:** Original plan required `module ftpsync`, package source in `ftpsync/`, and external imports as `import "ftpsync"`. Go resolves the subdirectory package as `ftpsync/ftpsync`, so these requirements were mutually incompatible.
- **Fix:** User selected "adjust plan then continue". Phase 9 plans were updated to use `ftpsync/ftpsync` unless the package is later moved to module root.
- **Files modified:** 09-01/09-03/09-04 PLAN files, external tests.
- **Commit:** 4c9a64e

**2. [Rule 3 - Blocking] Updated additional external test import**
- **Found during:** Task 1
- **Issue:** `go mod tidy` found `ftpsync/options_test.go` still imported `github.com/no-src/gofs/ftpsync`.
- **Fix:** Updated the import to `ftpsync/ftpsync` with the other external tests.
- **Files modified:** `ftpsync/options_test.go`
- **Commit:** 4c9a64e

**Total deviations:** 2.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

- Created file exists: `ftpsync/coverage_checklist_test.go`.
- Commit exists: `4c9a64e`.
- Commit exists: `c5a2e4a`.
- Verification commands passed.
