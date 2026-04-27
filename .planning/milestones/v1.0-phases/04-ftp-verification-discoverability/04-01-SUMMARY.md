---
phase: 04-ftp-verification-discoverability
plan: "01"
subsystem: testing
tags: [ftp, integration, pyftpdlib, fixtures, go, regression]

# Dependency graph
requires:
  - phase: 03-02
    provides: FTP one-way sync semantics and conservative metadata handling were already locked down in package tests
provides:
  - Repo-owned FTP bootstrap scripts now provision a deterministic local plain-FTP test server with inspectable PID/log artifacts
  - Tagged FTP integration fixtures now cover real disk→FTP and FTP→disk flows with nested paths plus delete/rename-relevant assertions
  - Real FTP integration work exposed and fixed pull-path protocol handling gaps so the suite passes against a live FTP server
affects: [ftp, integration, testing, verification, roadmap]

# Tech tracking
tech-stack:
  added: [pyftpdlib]
  patterns: ["FTP integration follows the existing integration/testdata harness pattern, while repo-owned scripts provision a passive-only local FTP service for deterministic verification"]

key-files:
  created: [scripts/ftp/init-ftp.sh, scripts/ftp/server.py, integration/integration_ftp_test.go, integration/testdata/conf/run-gofs-ftp-push-client.yaml, integration/testdata/conf/run-gofs-ftp-pull-client.yaml, integration/testdata/test/test-gofs-ftp-push.yaml, integration/testdata/test/test-gofs-ftp-pull.yaml, .planning/phases/04-ftp-verification-discoverability/04-01-SUMMARY.md]
  modified: [driver/ftp/ftp.go, driver/ftp/file.go, sync/driver_pull_client_sync.go, sync/disk_sync.go]

key-decisions:
  - "FTP integration setup uses repo-owned pyftpdlib scripts on 127.0.0.1:2121 with explicit passive ports, keeping Phase 4 inside the plain-FTP-only scope."
  - "FTP push and pull scenarios use sync_once fixtures in the existing integration harness so the YAML can deterministically seed nested, rename, and delete assertions before gofs starts."
  - "Phase 4 fixed live-server pull-path bugs in-place rather than weakening the fixture, because realistic verification exposed real FTP protocol handling defects in WalkDir, pull reads, and directory timestamping."

patterns-established:
  - "Repo-owned backend verification scripts should emit PID/log artifacts and bind test services to localhost only."
  - "Real protocol fixtures should preload deterministic inputs in init, then reserve actions for wait-and-assert phases when using sync_once integration flows."

requirements-completed: [TEST-01, TEST-02, TEST-03]

# Metrics
duration: 59min
completed: 2026-04-24
---

# Phase 4 Plan 01: FTP Integration Foundation Summary

**Repo-owned pyftpdlib bootstrap and tagged FTP fixtures now verify real disk→FTP and FTP→disk flows, including nested paths and delete/rename-relevant behavior, against a live plain-FTP server.**

## Performance

- **Duration:** 59 min
- **Started:** 2026-04-24T07:32:32Z
- **Completed:** 2026-04-24T08:29:51Z
- **Tasks:** 2
- **Files modified:** 12

## Accomplishments
- Added `scripts/ftp/init-ftp.sh` and `scripts/ftp/server.py` so the repository can launch its own deterministic passive-only FTP test server on `127.0.0.1:2121`.
- Added a new tagged integration suite plus FTP push/pull config and scenario YAML files that assert nested paths, content updates, deletes, and rename-relevant outcomes through the existing harness.
- Fixed several live-server FTP pull issues uncovered by the new suite: WalkDir lock re-entry, FTP success-response normalization during reads, pull stat/open ordering, and directory timestamp handling.

## Task Commits

Each task was committed atomically:

1. **Task 1: Add repository-owned FTP integration bootstrap scripts** - `a38ca01` (feat)
2. **Task 2: Add FTP integration runtime configs and scenario fixtures** - `8eff292` (feat)

**Plan metadata:** Pending / not committed in this session because `.planning/` already contains unrelated pre-existing changes.

## Files Created/Modified
- `scripts/ftp/init-ftp.sh` - Installs `pyftpdlib`, prepares the integration workspace, restarts the local FTP server, and writes PID/log artifacts.
- `scripts/ftp/server.py` - Starts a localhost-only writable FTP server user with explicit passive port range and full read/write/delete/rename permissions.
- `integration/integration_ftp_test.go` - Adds the FTP-tagged integration entrypoint using the existing thin test-case table pattern.
- `integration/testdata/conf/run-gofs-ftp-push-client.yaml` - Defines deterministic `disk→FTP` runtime config using the shipped FTP endpoint grammar.
- `integration/testdata/conf/run-gofs-ftp-pull-client.yaml` - Defines deterministic `FTP→disk` runtime config using sync-once pull semantics.
- `integration/testdata/test/test-gofs-ftp-push.yaml` - Verifies nested push paths, rename-relevant movement, delete propagation, and content hashing.
- `integration/testdata/test/test-gofs-ftp-pull.yaml` - Verifies nested pull paths, rename-relevant movement, delete propagation, and content hashing.
- `driver/ftp/ftp.go` - Makes FTP walk callbacks non-reentrant so real pull traversal can safely inspect a live server tree.
- `driver/ftp/file.go` - Normalizes successful FTP end-of-transfer/listing responses that the library surfaced as errors.
- `sync/driver_pull_client_sync.go` - Gets FTP metadata before opening the remote stream and tolerates protocol success responses during copy.
- `sync/disk_sync.go` - Skips directory chtimes for directory sources so FTP pull startup does not fail on directory-only time lookups.
- `.planning/phases/04-ftp-verification-discoverability/04-01-SUMMARY.md` - Records plan outcome, deviations, and verification status.

## Decisions Made
- Used repo-owned Python + `pyftpdlib` setup instead of an OS daemon or Docker container to keep FTP verification deterministic and phase-scoped.
- Switched both FTP fixtures to `sync_once: true` so seeded init data is present before gofs starts, matching the current harness semantics and avoiding flaky self-modifying action timing.
- Kept the runtime fixes minimal and localized to existing FTP driver/sync seams rather than adding a new FTP-specific execution path.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Installed missing `python3-pip` so the repo-owned FTP bootstrap could install `pyftpdlib`**
- **Found during:** Task 1 (Add repository-owned FTP integration bootstrap scripts)
- **Issue:** The execution environment had `python3` but no `pip` or `ensurepip`, so `scripts/ftp/init-ftp.sh` could not install the planned runtime dependency.
- **Fix:** Installed `python3-pip` via `apt` and kept the script’s bootstrap check so future failures are explicit.
- **Files modified:** None in repo (environment-only fix)
- **Verification:** `bash scripts/ftp/init-ftp.sh`
- **Committed in:** `a38ca01` (task commit)

**2. [Rule 1 - Bug] Fixed FTP pull deadlock caused by WalkDir invoking callbacks while holding the driver lock**
- **Found during:** Task 2 (Add FTP integration runtime configs and scenario fixtures)
- **Issue:** Live FTP pull traversal deadlocked because `WalkDir` held the connection mutex while callbacks re-entered `Stat()` on the same driver.
- **Fix:** Collected walk entries under the lock, then invoked callbacks after the FTP walker completed.
- **Files modified:** `driver/ftp/ftp.go`
- **Verification:** `go test ./integration -tags=integration_test_ftp -run TestIntegration_FTP -count=1`
- **Committed in:** `8eff292` (task commit)

**3. [Rule 1 - Bug] Fixed FTP pull stream handling for successful protocol replies surfaced as errors**
- **Found during:** Task 2 (Add FTP integration runtime configs and scenario fixtures)
- **Issue:** Live FTP reads and closes surfaced `226 Transfer complete` / MLST-style success replies as errors, causing gofs startup and pull writes to abort.
- **Fix:** Normalized successful FTP reply text in the FTP file wrapper and pull write path, and reordered pull metadata lookup to stat before opening the remote stream.
- **Files modified:** `driver/ftp/file.go`, `sync/driver_pull_client_sync.go`
- **Verification:** `go test ./integration -tags=integration_test_ftp -run TestIntegration_FTP -count=1`
- **Committed in:** `8eff292` (task commit)

**4. [Rule 1 - Bug] Fixed directory timestamp handling that broke FTP pull on directory entries**
- **Found during:** Task 2 (Add FTP integration runtime configs and scenario fixtures)
- **Issue:** Pulling directory trees attempted file-style time reads for remote directories, which caused 550 errors on FTP directory entries and aborted startup.
- **Fix:** Skipped directory `chtimes` when both source and destination are directories.
- **Files modified:** `sync/disk_sync.go`
- **Verification:** `go test ./integration -tags=integration_test_ftp -run TestIntegration_FTP -count=1`
- **Committed in:** `8eff292` (task commit)

---

**Total deviations:** 4 auto-fixed (3 bugs, 1 blocking environment fix)
**Impact on plan:** All deviations were required to make the planned realistic FTP verification actually work against a live server. No scope creep beyond correctness and execution unblockers.

## Issues Encountered
- The first draft of the fixtures assumed action-time mutations during long-running push/pull monitoring, but the live harness proved `sync_once` was the deterministic fit for seeded FTP verification in this plan.
- Real FTP integration exposed several pre-existing FTP pull protocol handling issues that package-level fake tests had not covered.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 4 Plan 02 can now add CI execution around a working tagged FTP integration suite instead of inventing the verification surface itself.
- The repo now has a concrete local FTP bootstrap and reusable fixtures that later CI steps can call directly.
- Metadata docs commit should be attempted only after resolving unrelated pre-existing `.planning/` workspace edits.

## Self-Check: PASSED

- Found summary file: `.planning/phases/04-ftp-verification-discoverability/04-01-SUMMARY.md`
- Found commit: `a38ca01`
- Found commit: `8eff292`
- Verified `bash scripts/ftp/init-ftp.sh` passes.
- Verified `go test ./integration -tags=integration_test_ftp -run TestIntegration_FTP -count=1` passes.
