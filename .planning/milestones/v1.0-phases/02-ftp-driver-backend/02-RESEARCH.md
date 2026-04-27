# Phase 2: FTP Driver Backend — Research

**Date:** 2026-04-24
**Status:** Complete
**Phase:** 02-ftp-driver-backend

## Research Question

What does gofs need in order to implement FTP as a real `driver.Driver` backend with minimal architectural change, conservative reconnect behavior, and metadata handling that prefers extra transfers over missed changes?

## Recommendation

Use `github.com/jlaffaye/ftp` as the FTP client library for Phase 2.

Why this fits the phase:
- It is a focused Go FTP client with the exact primitives the current `driver.Driver` contract needs: `Dial`, `Login`, `Walk`, `Retr`, `Stor`, `MakeDir`, `RemoveDirRecur`, `Rename`, `GetEntry`, `GetTime`, `SetTime`, and `NoOp`.
- It maps naturally onto the existing `driver/sftp` and `driver/minio` pattern where transport-specific connection lifecycle and reconnect logic stay inside the driver layer (per D-01, D-02, D-06, D-07).
- It exposes timeout configuration via `DialWithTimeout`, which fits the existing `ftp_timeout` VFS contract from Phase 1.
- It exposes time capability checks (`IsGetTimeSupported`, `IsSetTimeSupported`, `IsTimePreciseInList`) that let the implementation be conservative about metadata fidelity instead of making optimistic assumptions (per D-03, D-04, D-12, D-13, D-14).

## Library Evidence

Source reviewed:
- `https://pkg.go.dev/github.com/jlaffaye/ftp`

Relevant API surface from the docs:
- Connection/auth: `ftp.Dial`, `ftp.DialWithTimeout`, `(*ServerConn).Login`, `(*ServerConn).NoOp`, `(*ServerConn).Quit`
- Traversal/metadata: `(*ServerConn).Walk`, `(*ServerConn).List`, `(*ServerConn).GetEntry`, `(*ServerConn).GetTime`, `(*ServerConn).IsGetTimeSupported`, `(*ServerConn).IsTimePreciseInList`
- Mutation: `(*ServerConn).MakeDir`, `(*ServerConn).Stor`, `(*ServerConn).Retr`, `(*ServerConn).Delete`, `(*ServerConn).RemoveDirRecur`, `(*ServerConn).Rename`, `(*ServerConn).SetTime`

Important constraint from the docs:
- `ServerConn` is **not safe to be called concurrently**. The FTP driver must serialize access to the live connection with a mutex and keep retry/reconnect logic conservative.

## Architectural Fit

### Existing gofs pattern to preserve

The existing codebase already has the right seams:
- `core.VFS` carries transport-specific endpoint config.
- `driver.Driver` normalizes remote storage access.
- `sync/driver_push_client_sync.go` and `sync/driver_pull_client_sync.go` already implement the generic disk↔driver sync orchestration.
- Phase 1 already routed disk→FTP and FTP→disk into explicit FTP sync constructors.

Therefore Phase 2 should:
1. add `driver/ftp/` implementing `driver.Driver`
2. replace Phase 1 deferred FTP sync constructors with real driver-backed constructors
3. keep reconnect, metadata caveats, and operational errors inside the FTP driver

This directly matches D-01 and D-02 and avoids inventing an FTP-specific sync engine.

## Capability Mapping to Requirements

| Requirement | Research conclusion |
|-------------|---------------------|
| FTPD-01 connect/auth | Use `ftp.Dial(addr, ftp.DialWithTimeout(...))` + `Login(user, pass)` |
| FTPD-02 recursive traverse | Use `Walk(root)` and convert entries into `fs.WalkDirFunc` calls |
| FTPD-03 upload | Use `Stor(remotePath, reader)` |
| FTPD-04 download | Use `Retr(remotePath)` wrapped as an `http.File` |
| FTPD-05 mkdir | Use recursive helper over `MakeDir` to emulate `MkdirAll` |
| FTPD-06 delete | Use `Delete` for files and `RemoveDirRecur` for directories |
| FTPD-07 rename | Use `Rename(from, to)` |
| FTPD-08 size + mtime compare | Use `GetEntry`/`Walk` size, `GetTime` when supported, and document coarse fallback |
| FTPD-09 transient failures | Reuse existing `retry.Retry` pattern with bounded reconnect on connection-loss-style errors |

## Recommended Driver Shape

Add a new package under `driver/ftp/` with a shape parallel to `driver/sftp` and `driver/minio`:

- `ftp.go` — driver struct, connect/reconnect logic, CRUD methods, walk/stat/time methods
- `file.go` — adapter that wraps `*ftp.Response` to satisfy `http.File`
- `file_info.go` — adapter converting FTP entries/metadata to `fs.FileInfo`

Recommended constructor signature:

```go
func NewFTPDriver(remoteAddr string, ftpConfig core.FTPConfig, autoReconnect bool, r retry.Retry, maxTranRate int64, logger *logger.Logger) driver.Driver
```

Why this signature:
- Mirrors the SFTP/MinIO constructor pattern already used in `sync/`
- Lets `sync/ftp_push_client_sync.go` and `sync/ftp_pull_client_sync.go` stay thin
- Keeps timeout/passive-mode policy inside the driver where connection behavior belongs

## Metadata and Precision Policy

This is the hardest constraint in the phase because false “no-op” decisions would silently skip required transfers.

Recommended policy:
- Prefer `GetTime(path)` when `IsGetTimeSupported()` is true.
- Use `GetEntry(path)` / walker entry metadata for file size and directory detection.
- Treat list-derived timestamps as coarse unless `IsTimePreciseInList()` is true.
- When timestamp fidelity is uncertain, return the best available modification time but **do not add optimistic normalization** that could suppress legitimate transfers.
- If setting times is unsupported, return an explicit error from `Chtimes` rather than pretending success (per D-10, D-14). The current push sync path logs `Chtimes` failures without aborting the write, which is compatible with this policy.

This aligns with D-03, D-04, D-05, D-12, D-13, and D-14.

## Reconnect Strategy

Mirror the existing `driver/sftp` and `driver/minio` structure:
- Keep `online` state in the driver.
- Protect connection state with a mutex.
- Wrap each operation in `reconnectIfLost(func() error)`.
- Use the existing `retry.Retry` instance for reconnect attempts instead of inventing new retry state.
- Only retry on connection-loss / transport failures, not on semantic FTP errors such as bad credentials or missing files.

Recommended reconnect probe behavior:
- treat `NoOp` / operation-level transport failures as the signal that the control connection is stale
- set `online=false`
- call `retry.Do(d.Connect, "ftp reconnect")`
- retry the original operation once after successful reconnect

Do **not** add parallel transfer logic in this phase. The upstream library is single-connection, non-concurrent, and D-07 explicitly calls for bounded, non-aggressive recovery.

## Passive Mode Decision

Phase 1 exposed `ftp_passive` as a boolean, but the v1 project scope does **not** require active FTP mode; the requirements and roadmap only require passive-mode-compatible behavior, and active FTP is explicitly deferred in `REQUIREMENTS.md` (`FTPV2-02`).

Planning consequence:
- Phase 2 should preserve the config field and keep the implementation compatible with passive-mode operation.
- The driver should not attempt to invent active-mode handling.
- If the chosen client library cannot meaningfully switch to active mode, the implementation should fail clearly when the config demands unsupported behavior rather than silently misrepresenting transport semantics.

This preserves project truthfulness and avoids hiding a v2 feature behind a misleading v1 toggle.

## Testing Recommendation

Phase 4 owns realistic protocol-flow verification, so Phase 2 should focus on package-level automated coverage that locks down driver logic and sync wiring without needing a full external FTP environment.

Recommended approach:
- Add package tests under `driver/ftp/` using a narrow internal seam or fakeable FTP-connection adapter so driver behaviors can be tested deterministically.
- Cover connect/auth validation, traversal translation, upload/download invocation, recursive delete, rename, timestamp fallback, and reconnect-on-transport-error behavior.
- Add sync constructor tests proving the FTP push/pull constructors instantiate the real driver-backed path and no longer return Phase 1 deferred errors.

Why this split works:
- It gives immediate automated coverage for Phase 2 requirements.
- It keeps realistic end-to-end FTP server verification in Phase 4 where the roadmap already assigns it.

## Files the implementation will likely touch

- `go.mod`
- `go.sum`
- `driver/ftp/ftp.go`
- `driver/ftp/file.go`
- `driver/ftp/file_info.go`
- `driver/ftp/ftp_test.go`
- `sync/ftp_push_client_sync.go`
- `sync/ftp_pull_client_sync.go`
- `sync/sync_test.go`

## Common Pitfalls to Avoid

1. **Optimistic timestamp handling**
   - Bad: assuming LIST timestamps are always precise enough for no-op detection
   - Good: use capability checks and bias toward retransferring when uncertain

2. **Silent capability degradation**
   - Bad: returning nil for unsupported `Chtimes` or symlink behavior
   - Good: return explicit operational errors so higher layers can log the limitation clearly

3. **Concurrent connection use**
   - Bad: sharing one `*ftp.ServerConn` across simultaneous operations without locking
   - Good: serialize driver operations and keep reconnect state guarded by a mutex

4. **Pushing reconnect logic into sync code**
   - Bad: adding FTP-specific retry branches in `sync/driver_*`
   - Good: keep transport instability handling inside `driver/ftp`

## Validation Architecture

**Framework:** Go standard testing (`go test`)

**Fast feedback commands:**
- `go test ./driver/ftp ./sync -count=1`

**Broader confidence commands:**
- `go test ./driver/... ./sync ./monitor -count=1`

**Validation focus for this phase:**
- The FTP driver can satisfy the `driver.Driver` contract without Phase 1 deferred errors.
- Metadata handling is conservative and explicit.
- Reconnect behavior is bounded and test-covered.
- Sync constructors wire FTP through the same driver-backed abstractions as SFTP/MinIO.

## Final Research Verdict

Phase 2 should be planned as **two execution plans**:
1. implement the FTP driver package with conservative metadata and reconnect behavior
2. wire the FTP push/pull sync constructors to the new driver and replace deferred-error routing with real constructor coverage

This covers the phase goal with minimal architecture change and keeps realistic FTP flow verification where the roadmap already placed it: Phase 4.
