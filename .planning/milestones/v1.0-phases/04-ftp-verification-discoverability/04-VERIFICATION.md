---
phase: 04-ftp-verification-discoverability
verified: 2026-04-24T09:08:28Z
status: passed
score: 4/4 must-haves verified
overrides_applied: 0
---

# Phase 4: FTP Verification & Discoverability Verification Report

**Phase Goal:** Users and maintainers can trust and adopt the FTP path because it is tested against realistic flows and documented clearly.
**Verified:** 2026-04-24T09:08:28Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | Automated tests verify both `disk→FTP` and `FTP→disk` flows against an FTP test server. | ✓ VERIFIED | `scripts/ftp/init-ftp.sh` starts a localhost FTP server on `127.0.0.1:2121` with PID/log artifacts; `integration/integration_ftp_test.go:7-23` runs both push and pull cases; `go test -v -race -tags=integration_test_ftp ./integration -run TestIntegration_FTP -count=1` passed with both subtests green. |
| 2 | Automated coverage proves nested path handling plus delete or rename behavior on the FTP path. | ✓ VERIFIED | `integration/testdata/test/test-gofs-ftp-push.yaml:2-63` and `integration/testdata/test/test-gofs-ftp-pull.yaml:2-61` assert nested directories, content equality, renamed-file presence, old-name absence, and deleted-file absence. |
| 3 | User-facing documentation includes at least one working FTP configuration example. | ✓ VERIFIED | `README.md:476-491` and `README-CN.md:456-471` both contain FTP push and pull examples using shipped `ftp://...?...&ftp_user=...&ftp_pass=...&ftp_passive=true` grammar; `go test ./core -run 'TestNewVFS_FTPConfig|TestNewVFS_FTPConfigWithoutOptionalTimeout' -count=1` passed, confirming documented parameters match parser support. |
| 4 | User-facing documentation clearly states the v1 FTP limitations, including plain FTP only and no FTP↔FTP sync. | ✓ VERIFIED | `README.md:484-501` and `README-CN.md:464-481` state plain FTP only, no FTPS, passive-only/active-mode unsupported, no FTP↔FTP sync, and explicit backend capability failure behavior. |

**Score:** 4/4 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `scripts/ftp/init-ftp.sh` | repo-owned FTP bootstrap | ✓ VERIFIED | Exists, substantive (82 lines), installs `pyftpdlib`, binds `127.0.0.1:2121`, writes PID/log files, and starts `server.py`. |
| `scripts/ftp/server.py` | deterministic plain-FTP server | ✓ VERIFIED | Exists, substantive (51 lines), provisions writable user with `elradfmwMT` permissions and explicit passive ports. |
| `integration/testdata/conf/run-gofs-ftp-push-client.yaml` | FTP push fixture | ✓ VERIFIED | Uses `ftp://127.0.0.1:2121`, `ftp_user`, `ftp_pass`, `ftp_timeout`, and `ftp_passive=true`. |
| `integration/testdata/conf/run-gofs-ftp-pull-client.yaml` | FTP pull fixture | ✓ VERIFIED | Uses FTP source grammar with passive mode and `sync_once: true`. |
| `integration/testdata/test/test-gofs-ftp-push.yaml` | push assertions for realistic flow | ✓ VERIFIED | Covers nested directories, content update, rename-relevant assertion, delete assertion, and hash check. |
| `integration/testdata/test/test-gofs-ftp-pull.yaml` | pull assertions for realistic flow | ✓ VERIFIED | Covers nested directories, content update, rename-relevant assertion, delete assertion, and hash check. |
| `integration/integration_ftp_test.go` | tagged FTP suite entrypoint | ✓ VERIFIED | Exists, substantive, wired to the shared harness with explicit push/pull subtests. |
| `.github/workflows/go.yml` | CI execution of FTP integration coverage | ✓ VERIFIED | Contains Ubuntu-only `Init FTP` and `Test Integration FTP` steps. |
| `README.md` | English examples and limitations | ✓ VERIFIED | Contains FTP push/pull examples plus nearby and dedicated limitations text. |
| `README-CN.md` | Chinese examples and limitations | ✓ VERIFIED | Mirrors FTP push/pull examples and limitations in Chinese. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `scripts/ftp/init-ftp.sh` | `scripts/ftp/server.py` | background server startup | ✓ WIRED | `init-ftp.sh:43-50` launches `python3 "${SCRIPT_DIR}/server.py" ...`. |
| `integration/integration_ftp_test.go` | `integration/testdata/conf/run-gofs-ftp-push-client.yaml` | push test case | ✓ WIRED | `integration_ftp_test.go:14` references `run-gofs-ftp-push-client.yaml`. |
| `integration/integration_ftp_test.go` | `integration/testdata/conf/run-gofs-ftp-pull-client.yaml` | pull test case | ✓ WIRED | `integration_ftp_test.go:15` references `run-gofs-ftp-pull-client.yaml`. |
| `integration/integration_ftp_test.go` | `integration/testdata/test/test-gofs-ftp-push.yaml` | push assertions | ✓ WIRED | `integration_ftp_test.go:14` references `test-gofs-ftp-push.yaml`. |
| `integration/integration_ftp_test.go` | `integration/testdata/test/test-gofs-ftp-pull.yaml` | pull assertions | ✓ WIRED | `integration_ftp_test.go:15` references `test-gofs-ftp-pull.yaml`. |
| `.github/workflows/go.yml` | `scripts/ftp/init-ftp.sh` | Ubuntu setup step | ✓ WIRED | `go.yml:52-65` runs the init script before tagged FTP tests. |
| `README.md` | `core/vfs.go` | documented FTP endpoint grammar | ✓ WIRED | README examples use `ftp_user`, `ftp_pass`, and `ftp_passive`; `core/vfs.go:50-53, 200-203` parses those exact parameters. |
| `README-CN.md` | `README.md` | mirrored discoverability | ✓ WIRED | Chinese README mirrors the same push/pull examples and limitations set. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| --- | --- | --- | --- | --- |
| `integration/integration_ftp_test.go` | `testCases` entries | hardcoded fixture names consumed by `testIntegrationClientServer(...)` | Yes — each case points to real YAML fixture files executed by the shared harness | ✓ FLOWING |
| `README.md` FTP examples | FTP endpoint query parameters | `core/vfs.go` parser support and tests in `core/vfs_test.go` | Yes — parser recognizes documented `ftp_*` parameters and defaults | ✓ FLOWING |
| `README-CN.md` FTP examples | FTP endpoint query parameters | same shipped parser contract as English README | Yes — mirrored examples match the same parser-supported grammar | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| --- | --- | --- | --- |
| Repo-owned FTP server starts locally | `bash scripts/ftp/init-ftp.sh` | Server started on `127.0.0.1:2121`; PID/log paths printed | ✓ PASS |
| FTP parser accepts documented config grammar | `go test ./core -run 'TestNewVFS_FTPConfig|TestNewVFS_FTPConfigWithoutOptionalTimeout' -count=1` | `ok` | ✓ PASS |
| Tagged FTP suite exercises push and pull flows | `go test -v -race -tags=integration_test_ftp ./integration -run TestIntegration_FTP -count=1` | `PASS`; both push and pull real-server subtests passed | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| TEST-01 | 04-01, 04-02 | Automated tests verify `disk→FTP` sync behavior against an FTP test server | ✓ SATISFIED | `integration/integration_ftp_test.go:14`; `test-gofs-ftp-push.yaml`; tagged suite passed against live repo-owned FTP server. |
| TEST-02 | 04-01, 04-02 | Automated tests verify `FTP→disk` sync behavior against an FTP test server | ✓ SATISFIED | `integration/integration_ftp_test.go:15`; `test-gofs-ftp-pull.yaml`; tagged suite passed against live repo-owned FTP server. |
| TEST-03 | 04-01, 04-02 | Automated tests cover nested paths plus delete or rename behavior on the FTP path | ✓ SATISFIED | Both FTP YAML scenarios assert nested directories, renamed-file behavior, deleted-file absence, and content/hash checks. |
| DOC-01 | 04-03 | User-facing documentation includes at least one working FTP configuration example | ✓ SATISFIED | `README.md:476-491` and `README-CN.md:456-471` include working push and pull examples using shipped grammar. |
| DOC-02 | 04-03 | User-facing documentation states v1 FTP limitations, including plain FTP only and no FTP↔FTP sync | ✓ SATISFIED | `README.md:484-501` and `README-CN.md:464-481` explicitly state plain FTP only, no FTPS, passive-only, no FTP↔FTP, explicit capability failures. |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| --- | --- | --- | --- | --- |
| `sync/driver_pull_client_sync.go` | 68-70 | Successful FTP pull run still logs source-file close errors on protocol success replies | ⚠️ Warning | Does not block Phase 4 goal because the live suite passes and outputs are correct, but it adds misleading error noise to otherwise successful verification runs. |
| `sync/disk_sync.go` | 305-312 | Successful FTP pull run still logs file-time warnings when FTP metadata replies are not fully normalized | ⚠️ Warning | Non-blocking for this phase’s documented goals, but worth follow-up because it can reduce operator confidence in clean successful runs. |

### Human Verification Required

None.

### Gaps Summary

No blocking gaps found. Phase 4 achieved the roadmap contract: the repository can provision a real FTP server from repo-owned scripts, the tagged integration suite verifies both push and pull realistic flows in CI, and both README surfaces make FTP discoverable with explicit v1 limits. Non-blocking warning: successful FTP pull runs still emit noisy log messages from protocol-reply handling, but this did not prevent the verified behaviors required by Phase 4.

---

_Verified: 2026-04-24T09:08:28Z_
_Verifier: the agent (gsd-verifier)_
