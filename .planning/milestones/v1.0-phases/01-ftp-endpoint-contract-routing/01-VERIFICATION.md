---
phase: 01-ftp-endpoint-contract-routing
verified: 2026-04-23T09:47:17Z
status: passed
score: 3/3 must-haves verified
overrides_applied: 0
---

# Phase 1: FTP Endpoint Contract & Routing Verification Report

**Phase Goal:** Users can define FTP endpoints in config and have gofs recognize them as valid source or destination sync targets.
**Verified:** 2026-04-23T09:47:17Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | User can configure an FTP endpoint as either the source or destination of a sync using host, port, username, password, and remote path. | ✓ VERIFIED | `core/vfs.go` adds `ftp://` recognition in `NewVFS`, FTP-specific query keys (`ftp_user`, `ftp_pass`, `ftp_timeout`, `ftp_passive`), `FTPConfig`, and default port `21`; `core/vfs_test.go` covers source and destination fixtures plus round-trip/default-port assertions. |
| 2 | User can set FTP timeout and passive-mode-compatible behavior per endpoint in configuration. | ✓ VERIFIED | `core/vfs.go` parses `ftp_timeout` and `ftp_passive` into `FTPConfig`; getters `FTPTimeout()` and `FTPPassiveMode()` expose them; `core/vfs_test.go` verifies both populated and omitted timeout cases and passive-mode true/false behavior. |
| 3 | A configured `ftp://` endpoint is accepted by gofs and routed into the existing sync and monitor selection flow instead of being treated as unsupported. | ✓ VERIFIED | `sync/sync.go` has explicit disk→FTP and FTP→disk branches to `NewFTPPushClientSync` / `NewFTPPullClientSync`; `monitor/monitor.go` routes FTP sources to `NewFTPPullClientMonitor`; `sync/sync_test.go` and `monitor/monitor_test.go` assert FTP paths return FTP-deferred errors rather than generic unsupported errors; `go test ./sync ./monitor -count=1` passes. |

**Score:** 3/3 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `core/vfs.go` | FTP scheme parsing, FTP query fields, and default port behavior | ✓ VERIFIED | Exists, substantive, and defines `ftp` scheme branch, FTP config storage/getters, and `ftpServerDefaultPort = 21`. |
| `core/vfs_test.go` | FTP parse, round-trip, and default-port coverage | ✓ VERIFIED | Exists, substantive, and includes canonical FTP source/destination fixtures plus marshal/unmarshal/default-port tests. |
| `conf/config.go` | `conf.Config` continues to carry `core.VFS` endpoints unchanged | ✓ VERIFIED | `Source core.VFS` and `Dest core.VFS` remain unchanged, preserving config wiring. |
| `sync/sync.go` | FTP source/destination factory routing | ✓ VERIFIED | Exists, substantive, and routes disk→FTP / FTP→disk combinations into FTP-specific constructors. |
| `monitor/monitor.go` | FTP source monitor routing | ✓ VERIFIED | Exists, substantive, and routes FTP source to `NewFTPPullClientMonitor`. |
| `sync/ftp_push_client_sync.go` | FTP-specific sync constructor entry point | ✓ VERIFIED | Thin constructor exists and returns explicit phase-2 deferred FTP error. |
| `sync/ftp_pull_client_sync.go` | FTP-specific sync constructor entry point | ✓ VERIFIED | Thin constructor exists and is wired from the sync factory. |
| `monitor/ftp_pull_client_monitor.go` | FTP-specific pull monitor entry point | ✓ VERIFIED | Thin constructor exists and is wired from the monitor factory. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `core/vfs.go` | `conf/config.go` | Source and Dest fields remain typed as `core.VFS` | ✓ WIRED | `conf/config.go` keeps `Source core.VFS` and `Dest core.VFS`; parsed FTP VFS values flow through existing config model. |
| `core/vfs.go` | `core/vfs_test.go` | FTP endpoint examples and parse assertions | ✓ WIRED | FTP fixtures in tests exercise parser behavior, round-trip handling, timeout/passive config, and default port logic. |
| `sync/sync.go` | `sync/ftp_push_client_sync.go` | `source.IsDisk() && dest.Is(core.FTP)` dispatch | ✓ WIRED | `newSync()` routes disk→FTP to `NewFTPPushClientSync`. |
| `sync/sync.go` | `sync/ftp_pull_client_sync.go` | `source.Is(core.FTP) && dest.IsDisk()` dispatch | ✓ WIRED | `newSync()` routes FTP→disk to `NewFTPPullClientSync`. |
| `monitor/monitor.go` | `monitor/ftp_pull_client_monitor.go` | source FTP monitor branch | ✓ WIRED | `NewMonitor()` routes FTP sources to `NewFTPPullClientMonitor`. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| --- | --- | --- | --- | --- |
| `core/vfs.go` | Parsed FTP endpoint fields | URL query params in `NewVFS`/`parse` | Yes | ✓ FLOWING |
| `sync/sync.go` | `source` / `dest` VFS type dispatch | `Option.Source` / `Option.Dest` | Yes | ✓ FLOWING |
| `monitor/monitor.go` | `source` VFS type dispatch | `opt.Syncer.Source()` | Yes | ✓ FLOWING |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| --- | --- | --- | --- |
| FTP sync routing avoids generic unsupported branch | `go test ./sync ./monitor -count=1` | `ok github.com/no-src/gofs/sync`, `ok github.com/no-src/gofs/monitor` | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| FTP-01 | `01-01-PLAN.md`, `01-02-PLAN.md` | User can configure an FTP endpoint as a sync source using host, port, username, password, and remote path | ✓ SATISFIED | `core/vfs.go` parses FTP source endpoints and stores FTP fields; `monitor/monitor.go` routes FTP sources; `core/vfs_test.go` and `monitor/monitor_test.go` cover source-side recognition. |
| FTP-02 | `01-01-PLAN.md`, `01-02-PLAN.md` | User can configure an FTP endpoint as a sync destination using host, port, username, password, and remote path | ✓ SATISFIED | `core/vfs.go` parses FTP destination endpoints and stores FTP fields; `sync/sync.go` routes disk→FTP; `core/vfs_test.go` and `sync/sync_test.go` cover destination-side recognition. |
| FTP-03 | `01-01-PLAN.md` | User can configure FTP connection timeout behavior for an endpoint | ✓ SATISFIED | `core/vfs.go` parses `ftp_timeout`; getter exposed via `FTPTimeout()`; tests verify populated and omitted timeout handling. |
| FTP-04 | `01-01-PLAN.md` | User can configure passive-mode-compatible FTP behavior for an endpoint | ✓ SATISFIED | `core/vfs.go` parses `ftp_passive` into `FTPConfig.PassiveMode`; tests verify true and false cases. |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| --- | --- | --- | --- | --- |
| `sync/ftp_push_client_sync.go` | 5 | Explicit deferred error placeholder | ℹ️ Info | Intentional Phase 1 boundary: FTP routing target exists while backend behavior is deferred to Phase 2. |
| `sync/ftp_pull_client_sync.go` | 5 | Explicit deferred error placeholder | ℹ️ Info | Same intentional boundary; not a stub relative to this phase goal because the requirement is recognition/routing, not protocol execution. |
| `monitor/ftp_pull_client_monitor.go` | 5 | Explicit deferred error placeholder | ℹ️ Info | Same intentional boundary for monitor path. |

### Gaps Summary

No goal-blocking gaps found. FTP endpoints are parsable as first-class `core.FTP` VFS values, preserve FTP-specific config fields, default the port to `21` when omitted, and are intentionally routed through sync/monitor factories instead of falling into generic unsupported-path handling. The remaining FTP runtime behavior is explicitly deferred to Phase 2 and aligns with the roadmap.

---

_Verified: 2026-04-23T09:47:17Z_
_Verifier: the agent (gsd-verifier)_
