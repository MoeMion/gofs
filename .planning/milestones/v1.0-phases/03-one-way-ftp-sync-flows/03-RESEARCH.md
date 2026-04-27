# Phase 3: One-Way FTP Sync Flows — Research

**Date:** 2026-04-24
**Phase:** 03-one-way-ftp-sync-flows
**Goal:** Make `disk→FTP` and `FTP→disk` usable as real one-way sync flows without changing gofs sync semantics.

## Executive Summary

Phase 3 should stay minimal-change and build on Phase 2's real `driver/ftp` backend instead of adding a new FTP-specific sync engine. The biggest functional gap is not transfer code — generic driver-backed push/pull sync paths are already wired — but sustained `FTP→disk` operation: `monitor/ftp_pull_client_monitor.go` is still a placeholder, so FTP source mode cannot currently run truthfully in long-running mode.

Recommended Phase 3 approach:

1. Replace the FTP pull monitor placeholder with a thin wrapper around `driverPullClientMonitor`, following the SFTP/MinIO monitor pattern.
2. Make FTP source long-running behavior explicit and truthful: FTP has no event stream, so long-running mode must use polling (`sync_cron`) or `sync_once`; if neither is configured, fail clearly instead of idling successfully.
3. Preserve one-way semantics in the existing generic sync helpers (`driverPushClientSync`, `driverPullClientSync`); add targeted flow tests first, then make only the minimal sync-layer adjustments required by those tests.
4. Validate no-op behavior conservatively: use supported metadata paths (`GetTime`, precise listing time when available), and accept extra transfers when metadata is ambiguous rather than risking missed changes.

## What Already Exists

### Existing implementation surface

- `sync/ftp_push_client_sync.go` already routes `disk→FTP` through `driverPushClientSync` with `ftp.NewFTPDriver(...)`.
- `sync/ftp_pull_client_sync.go` already routes `FTP→disk` through `driverPullClientSync`, resetting `sourceAbsPath`, `statFn`, and `getFileTimeFn` to FTP driver behavior.
- `driver/ftp/ftp.go` already implements connect, traversal, read/write, mkdir, delete, rename, stat, `GetFileTime`, and bounded reconnect behavior.
- `monitor/driver_pull_client_monitor.go` already provides a polling-oriented source monitor with `sync_once` and `sync_cron` support.

### Current functional gap

- `monitor/ftp_pull_client_monitor.go` still returns a deferred placeholder error.
- `monitor.NewMonitor(...)` routes FTP sources into that placeholder, so sustained `FTP→disk` operation is blocked even though the driver and pull sync path exist.

## Codebase Pattern Findings

### Monitor pattern for polled remote sources

`monitor/sftp_pull_client_monitor.go` and `monitor/minio_pull_client_monitor.go` are intentionally thin:

```go
type sftpPullClientMonitor struct {
	driverPullClientMonitor
}

func NewSftpPullClientMonitor(opt Option) (m Monitor, err error) {
	m = &sftpPullClientMonitor{
		driverPullClientMonitor: driverPullClientMonitor{
			baseMonitor: newBaseMonitor(opt),
		},
	}
	return m, nil
}
```

This is the closest analog for FTP source monitoring.

### What `driverPullClientMonitor` actually does

- `sync_once=true` → immediately runs one `SyncOnce(...)` and shuts down.
- `sync_cron` set → registers cron-based polling and keeps running.
- neither set → starts successfully but performs no sync work.

That last behavior is acceptable for some backends but violates Phase 3's locked decisions for FTP because it would look like long-running FTP source mode started successfully while doing nothing.

## External Library Notes

From `github.com/jlaffaye/ftp` docs (`pkg.go.dev/github.com/jlaffaye/ftp`):

- `Dial(..., ftp.DialWithTimeout(...))` is the correct timeout configuration path.
- `ServerConn` is **not safe to be called concurrently**, which validates Phase 2's mutex/serialized access design.
- `GetTime`, `IsGetTimeSupported`, `IsSetTimeSupported`, and `IsTimePreciseInList` are the relevant metadata capability checks.
- `Walk`, `Retr`, `Stor`, `RemoveDirRecur`, and `Rename` map directly to the existing driver abstraction.

Implication for Phase 3: metadata-sensitive no-op tests should assume different fidelity depending on `GetTime` and precise listing support, not perfect timestamp behavior across all FTP servers.

## Recommendations

### 1. Implement FTP long-running source behavior in monitor/, not sync/

Per D-02, do not add a new orchestration model. Reuse `driverPullClientMonitor` and keep FTP-specific logic at the monitor boundary.

### 2. Fail explicitly when FTP source is configured without a polling mode

Per D-09, D-10, and D-11, FTP long-running mode should not appear healthy when no sustained mechanism exists. Recommended behavior:

- allow `sync_once`
- allow `sync_cron`
- if both are absent for FTP source monitor startup, return a clear error explaining that FTP source mode requires polling (`sync_cron`) or `sync_once`

### 3. Optionally perform an initial sync before cron-based polling begins

This is a good minimal-change improvement for FTP source UX because cron-only polling otherwise waits for the first tick before producing any files. If implemented, it should be FTP-source-specific or justified as a safe generic behavior change.

### 4. Use tests to prove semantics before changing sync code

The one-way behaviors required by Phase 3 already live in generic helpers:

- `driverPushClientSync` owns create/write/delete/rename flow for `disk→remote`
- `driverPullClientSync` owns create/write/delete/rename flow for `remote→disk`

So Phase 3 should add targeted tests for:

- `disk→FTP` create/update/delete/rename behavior
- `FTP→disk` create/update/delete behavior under polling
- no-op behavior on second run when metadata is sufficient
- explicit failure, not silent success, for unsupported FTP-source runtime shapes

Only after those tests exist should minimal code changes be made if a gap is proven.

## Likely Files for Phase 3

### High-confidence implementation files

- `monitor/ftp_pull_client_monitor.go`
- `monitor/monitor_test.go`

### High-confidence verification / flow files

- `sync/ftp_push_client_sync.go`
- `sync/ftp_pull_client_sync.go`
- `sync/driver_push_client_sync.go`
- `sync/driver_pull_client_sync.go`
- `sync/sync_test.go` or a new FTP-focused sync test file

## Common Pitfalls

1. **Silent idle startup for FTP source mode**
   - Risk: monitor starts with neither `sync_once` nor `sync_cron`, so user sees “running” but nothing happens.
   - Mitigation: explicit startup error for unsupported runtime shape.

2. **Adding FTP-specific orchestration in sync layer**
   - Risk: Phase 3 duplicates behavior already owned by generic driver sync helpers.
   - Mitigation: keep protocol-specific changes thin and local to FTP monitor/wiring unless tests prove a real sync-layer bug.

3. **Over-promising no-op guarantees**
   - Risk: FTP timestamps are coarse or inconsistent; strict “never retransfer” logic can miss real changes.
   - Mitigation: preserve conservative metadata policy and test “no unnecessary transfers under supported metadata conditions,” matching the roadmap wording.

4. **Accidentally expanding into Phase 4**
   - Risk: full realistic FTP-server integration suite and user docs belong to Phase 4.
   - Mitigation: Phase 3 verification should stay at package/targeted flow coverage unless a very small fixture addition is necessary to support implementation confidence.

## Validation Architecture

Phase 3 should use fast package-level feedback with focused FTP flow coverage:

- **Quick command:** `go test ./monitor ./sync -count=1`
- **Full command:** `go test ./driver/ftp ./monitor ./sync -count=1`

Recommended validation split:

- `03-01`: monitor startup and routing tests proving FTP source uses the real pull monitor path and rejects idle long-running configuration.
- `03-02`: flow-focused sync tests proving `disk→FTP` / `FTP→disk` one-way semantics and conservative no-op behavior remain intact.

Manual verification is optional rather than primary for this phase because the implementation target is backend/runtime behavior, not UI.

## Source Audit Inputs for Planning

| Source | ID | Item | Planning implication |
|---|---|---|---|
| GOAL | — | Real one-way `disk→FTP` and `FTP→disk` usability | Plans must cover both directions, not just FTP source polling |
| REQ | SYNC-01 | User can run sync from local disk to FTP destination | Need flow-focused coverage or fixes for push semantics |
| REQ | SYNC-02 | User can run sync from FTP source to local disk | Must replace FTP monitor placeholder |
| REQ | SYNC-03 | Preserve one-way semantics | Tests must cover delete/rename/no bidirectional behavior |
| REQ | SYNC-04 | Second sync run avoids unnecessary transfers when metadata supports it | Tests must encode supported-metadata no-op expectations |
| CONTEXT | D-03 / D-04 | FTP source must support long-running mode via polling if needed | FTP monitor plan must cover sustained mode explicitly |
| CONTEXT | D-09 / D-10 / D-11 | Capability gaps must fail explicitly, not silently degrade | Startup/runtime error behavior must be part of acceptance criteria |

## Planning Guidance

Use **two plans** unless discovery during planning proves a concrete sync-layer bug that warrants an extra isolated fix plan:

1. **FTP source monitor plan** — make long-running FTP source mode real and truthful.
2. **Flow semantics plan** — lock down one-way `disk↔FTP` semantics and no-op behavior with targeted tests, then patch minimally if required.

---

*Research completed for Phase 03 on 2026-04-24.*
