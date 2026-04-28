---
phase: 06-one-shot-disk-ftp-sync-through-library-api
plan: "01"
subsystem: api
tags: [ftp, ftpsync, oneshot, result-summary, legacy-adapter, go, testing]

# Dependency graph
requires:
  - phase: 05-02
    provides: Structured public errors, validation ordering, and context-aware SyncOnce contract
  - phase: 05-03
    provides: Library-local hook dispatch and no-op default observability surface
provides:
  - Compact public SyncOnce result summary fields for one-shot library execution
  - Typed transfer-error path that preserves Result data for partial failures
  - Internal adapter that builds legacy disk/FTP sync dependencies from typed ftpsync options
affects: [phase-06, phase-07, public-api, oneshot-execution, ftp-library]

# Tech tracking
tech-stack:
  added: []
  patterns: ["SyncOnce remains public and typed while legacy sync construction stays internal to ftpsync", "Partial one-shot failures return Result plus typed ErrTransfer without exposing credentials"]

key-files:
  created: [ftpsync/oneshot.go, ftpsync/oneshot_test.go, .planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-01-SUMMARY.md]
  modified: [ftpsync/service.go, ftpsync/errors.go, ftpsync/context_test.go, ftpsync/hooks_test.go, ftpsync/options_test.go]

key-decisions:
  - "Expanded Result with compact summary counters and timestamps instead of introducing per-file report slices, preserving the summary-first Phase 6 contract."
  - "Kept FTP URL/query construction as an internal adapter detail in ftpsync so the public API remains typed-options-only."
  - "Classified partial one-shot failures as ErrTransfer while preserving a populated Result for caller logging and retry decisions."

patterns-established:
  - "FTPSyncService.SyncOnce validates and checks context first, then delegates to a library-local helper instead of reviving CLI/runtime entrypoints."
  - "Typed IgnoreRule values are adapted in-package to ignore.PathIgnore rather than by writing config files or invoking parser-driven runtime paths."

requirements-completed: [ONCE-05]

# Metrics
duration: 12min
completed: 2026-04-27
---

# Phase 06 Plan 01: One-Shot Result and Legacy Adapter Summary

**Compact SyncOnce result summaries with partial-failure transfer errors and an internal typed-options adapter into the legacy FTP sync engine.**

## Performance

- **Duration:** 12 min
- **Started:** 2026-04-27T06:09:31Z
- **Completed:** 2026-04-27T06:21:31Z
- **Tasks:** 2
- **Files modified:** 8

## Accomplishments

- Replaced the `SyncOnce` unsupported placeholder with real library-local dispatch that returns compact one-shot summary data.
- Added typed transfer error construction so partial failures can return both `Result` and a non-nil `ErrTransfer` without leaking FTP secrets.
- Added an internal adapter that converts typed `ftpsync.Options` into legacy `core.VFS`, retry, ignore, and `sync.NewSync` dependencies without changing the public API.

## Task Commits

Plan implementation was committed atomically:

1. **Task 1 + Task 2: one-shot result scaffolding and legacy adapter** - `e2e4bc3` (feat)

**Plan metadata:** Pending

## Files Created/Modified

- `ftpsync/service.go` - Expands `Result` and routes `SyncOnce` into library-local execution.
- `ftpsync/errors.go` - Adds transfer-error helpers for password-safe one-shot failure reporting.
- `ftpsync/oneshot.go` - Implements compact result construction, internal adapter building, ignore-rule adaptation, and SyncOnce execution seam/scaffold.
- `ftpsync/oneshot_test.go` - Covers summary results, partial-failure behavior, compactness, and legacy-adapter construction.
- `ftpsync/context_test.go` - Updates public-context expectations now that `SyncOnce` no longer returns unsupported capability.
- `ftpsync/hooks_test.go` - Adjusts hook dependency checks to match the new Phase 6 legacy-adapter boundary.
- `ftpsync/options_test.go` - Adjusts package dependency checks to match the new Phase 6 legacy-adapter boundary.

## Decisions Made

- Used summary counters plus start/end timestamps for `Result` so callers get useful observability without a file-by-file API surface.
- Kept adapter construction internal and string-based only at the `core.NewVFS("ftp://...")` boundary, avoiding any typed API expansion.
- Preserved `StartBackground` as unsupported so Phase 7 remains the sole owner of background lifecycle behavior.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Relaxed prior-phase dependency-boundary tests to permit the approved Phase 6 legacy sync adapter**
- **Found during:** Task 2
- **Issue:** Existing package-boundary tests assumed `ftpsync` still excluded all legacy runtime-adjacent imports, but this plan explicitly introduces an internal adapter that routes through the existing sync engine.
- **Fix:** Updated boundary tests to stop rejecting dependencies that are now intentionally reachable through the approved Phase 6 adapter path.
- **Files modified:** `ftpsync/hooks_test.go`, `ftpsync/options_test.go`
- **Verification:** `go test ./ftpsync -count=1`
- **Committed in:** `e2e4bc3`

---

**Total deviations:** 1 auto-fixed (1 blocking)
**Impact on plan:** Necessary to align earlier dependency assertions with the approved Phase 6 architecture. No public API scope creep was added.

## Known Stubs

- `ftpsync/oneshot.go` — `runSyncOnceScaffold` currently provides result/orchestration scaffolding and adapter wiring only; real direction-specific transfer walking remains intentionally deferred to Plans 06-02 and 06-03.

## Threat Flags

None. The plan adds only the expected `ftpsync` → legacy sync-engine adapter at the planned trust boundary and does not introduce new public network endpoints or auth surfaces.

## Issues Encountered

- Existing dependency-boundary tests from Phase 5 conflicted with the planned reuse of legacy sync construction in Phase 6. They were updated to reflect the now-approved internal adapter boundary.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness

- Plan 06-02 can replace the current execution scaffold with real local→FTP one-shot walking while preserving the new result/error contract.
- Plan 06-03 can wire FTP→local one-shot execution and cwd-safety regression coverage on top of the same adapter/result structure.

## Self-Check: PASSED

- Found summary file: `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-01-SUMMARY.md`
- Found created/modified files: `ftpsync/service.go`, `ftpsync/errors.go`, `ftpsync/oneshot.go`, `ftpsync/oneshot_test.go`, `ftpsync/context_test.go`, `ftpsync/hooks_test.go`, `ftpsync/options_test.go`
- Found implementation commit: `e2e4bc3`
- Verified plan commands passed: `go test ./ftpsync -run 'TestSyncOnce(Result|Partial|Hooks)' -count=1`, `go test ./ftpsync -run 'Test(NewFTPSyncService|SyncOnceBuildsLegacyAdapter)' -count=1`, `go test ./ftpsync -count=1`

---
*Phase: 06-one-shot-disk-ftp-sync-through-library-api*
*Completed: 2026-04-27*
