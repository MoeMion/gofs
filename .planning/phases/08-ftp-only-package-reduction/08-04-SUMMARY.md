---
phase: 08-ftp-only-package-reduction
plan: 04
subsystem: package-reduction
tags: [go-modules, ci, dependency-graph, ftpsync]

# Dependency graph
requires:
  - phase: 08-03
    provides: Deleted old CLI/server/protocol runtime packages and release surfaces
provides:
  - Tidied root Go module with only FTP library dependencies
  - CI workflow scoped to build and test the reduced library module
  - Root dependency blacklist proof using go list -deps ./...
affects: [phase-09-verification-docs, release-packaging, dependency-policy]

# Tech tracking
tech-stack:
  added: []
  patterns: [go mod tidy after runtime pruning, root dependency blacklist test, reduced CI build/test matrix]

key-files:
  created: []
  modified: [go.mod, go.sum, .github/workflows/go.yml, ftpsync/dependency_boundary_test.go, ftpsync/background_test.go]

key-decisions:
  - "Kept the reduced root module on the existing github.com/no-src/gofs module path while pruning the dependency graph to ftpsync and its FTP/watch dependencies."
  - "Scoped CI to default build and test commands for ./... because old integration, SFTP, MinIO, Docker, and CLI surfaces were removed in Phase 8."
  - "Expanded the dependency boundary proof from ./ftpsync to ./... so the final root module build graph is explicitly covered."

patterns-established:
  - "Dependency boundary tests should execute the same go list -deps ./... command used as release evidence."
  - "Background error-observation tests should wait for asynchronous handle state instead of racing worker bookkeeping."

requirements-completed: [PRUNE-01, PRUNE-02, PRUNE-03, PRUNE-04]

# Metrics
duration: 4min
completed: 2026-04-28
---

# Phase 08 Plan 04: Tidy Module Dependencies and Prove FTP-Only Graph Summary

**Tidied FTP-only root module with reduced CI and root-wide dependency blacklist proof for the final package graph**

## Performance

- **Duration:** 4 min
- **Started:** 2026-04-28T01:04:35Z
- **Completed:** 2026-04-28T01:08:22Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments

- Pruned `go.mod` and `go.sum` from the old runtime dependency set to the FTP library graph: `github.com/jlaffaye/ftp`, `github.com/fsnotify/fsnotify`, `golang.org/x/text`, and minimal test/transitive dependencies.
- Reduced `.github/workflows/go.yml` to cross-platform Go setup, `go build ./...`, and `go test ./...`, removing stale SFTP/MinIO/FTP integration setup and removed `integration` package jobs.
- Strengthened `TestPackageDependencyBoundaryRejectsOldRuntime` to prove the entire root module with `go list -deps ./...`, not only the `./ftpsync` package path.
- Stabilized the background runtime-error assertion discovered during full-suite verification so asynchronous handle error publication is observed deterministically.

## Task Commits

Each task was committed atomically:

1. **Task 1: Tidy module dependencies and update CI test scope** - `322981f` (chore)
2. **Task 2: Run final dependency blacklist and full test proof** - `79cd5d4` (test)

**Plan metadata:** pending final docs commit

## Files Created/Modified

- `go.mod` - Removed old CLI/server/runtime module requirements and retained only the reduced FTP library dependency set.
- `go.sum` - Removed stale checksums for Gin, gRPC, SFTP, MinIO, Redis/cache, QUIC, OAuth, protobuf, and other deleted runtime surfaces.
- `.github/workflows/go.yml` - Removed deleted integration package jobs and old SFTP/MinIO/FTP setup scripts from CI.
- `ftpsync/dependency_boundary_test.go` - Runs the blacklist proof against `go list -deps ./...` for root-module coverage.
- `ftpsync/background_test.go` - Waits for asynchronous background error publication before asserting the latest runtime transfer error.

## Decisions Made

- Kept the root module path unchanged (`github.com/no-src/gofs`) for this plan; Phase 9 can document the supported import path without a module rename.
- Treated CI integration jobs referencing removed package paths and protocol setup scripts as stale runtime artifacts and removed them rather than preserving broken references.
- Promoted the dependency blacklist command to root scope so the final proof matches the plan's `go list -deps ./...` success criterion.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Stabilized asynchronous background error assertion**
- **Found during:** Task 1 (Tidy module dependencies and update CI test scope)
- **Issue:** `go test ./... -count=1` intermittently observed `Handle.Err()` before the background worker had published the simulated transfer failure, causing `TestBackgroundWaitReturnsFinalError` to fail.
- **Fix:** Added a test helper that waits briefly for the expected `ErrTransfer` state before asserting, matching the asynchronous background lifecycle semantics.
- **Files modified:** `ftpsync/background_test.go`
- **Verification:** `go test ./... -count=1`
- **Committed in:** `322981f`

---

**Total deviations:** 1 auto-fixed (1 bug)
**Impact on plan:** The fix was required for the planned full-suite verification and did not change runtime behavior.

## Issues Encountered

- The shell did not have `gofmt` on `PATH`; used `/usr/local/go/bin/gofmt` for formatting the touched Go test files.

## User Setup Required

None - no external service configuration required.

## Verification

All required verification passed:

- `go mod tidy`
- `go test ./... -count=1`
- `go test ./ftpsync -run TestPackageDependencyBoundaryRejectsOldRuntime -count=1`
- `go list -deps ./...`
- Grep-tool scans found no forbidden module entries for `gin-gonic`, `google.golang.org/grpc`, `pkg/sftp`, `minio-go`, `nscache`, `quic-go`, `oauth2`, or `protobuf` in `go.mod` or `go.sum`.
- Grep-tool scan found no removed path/protocol references (`cmd/gofs`, `integration`, `sftp`, `minio`) in `.github/workflows/go.yml`.

## Known Stubs

None - no placeholder or unwired data stubs were introduced by this plan.

## Threat Flags

None - no new network endpoints, auth paths, file access patterns, or trust-boundary schema changes were introduced.

## Next Phase Readiness

- Phase 9 can document the final reduced import/package shape and migration path against a clean FTP-only dependency graph.
- Remaining Phase 9 work should focus on verification examples, README/package documentation, and release/migration notes rather than dependency pruning.

## Self-Check: PASSED

- Found `go.mod`, `go.sum`, `.github/workflows/go.yml`, `ftpsync/dependency_boundary_test.go`, and `ftpsync/background_test.go`.
- Found task commits `322981f` and `79cd5d4` in git history.

---
*Phase: 08-ftp-only-package-reduction*
*Completed: 2026-04-28*
