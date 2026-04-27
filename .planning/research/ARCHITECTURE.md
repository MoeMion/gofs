# Architecture Research: v2.0 FTP Sync Library

**Project:** gofs  
**Domain:** Go library extraction for disk↔FTP synchronization  
**Researched:** 2026-04-27  
**Confidence:** HIGH for internal package disposition; MEDIUM for final API naming until implementation discussion locks it.

## Recommendation

Replace the current CLI/server-centered product shape with a small library facade around the existing FTP-capable sync core. The v2.0 architecture should make `FTPSyncService` the only supported public entry point and demote existing package-level factories (`sync.NewSync`, `monitor.NewMonitor`, `core.NewVFS`) to internal implementation details.

Do **not** rewrite the sync algorithm first. Keep the FTP driver, disk sync, generic driver sync adapters, ignore/retry/logger support, and the local/FTP pull monitor mechanics; delete or isolate HTTP, gRPC, task, SFTP, MinIO, auth/session, daemon, and CLI runtime packages. This gets the package small while preserving the highest-risk behavior that already works: disk↔FTP copy, metadata comparison, remote path mapping, delete/rename semantics, and scheduled/persistent sync.

## Target Shape

```text
consumer app
  → ftpsync.FTPSyncService             public API
      → internal/config normalization  typed options, no CLI flags required
      → internal/core endpoint model    disk + ftp only
      → internal/engine sync factory    disk↔ftp only
      → internal/monitor background     fsnotify for disk source, cron/poll for FTP source
      → internal/driver/ftp             FTP protocol operations
```

### Public API Boundary

Expose one package, preferably module root or `ftpsync`, and keep all brownfield internals unexported or explicitly marked internal.

```go
type FTPSyncService interface {
    SyncOnce(ctx context.Context, req SyncRequest) (*SyncResult, error)
    Start(ctx context.Context, req SyncRequest) (BackgroundSync, error)
}

type BackgroundSync interface {
    Done() <-chan error
    Stop(ctx context.Context) error
}
```

`SyncRequest` should be typed, not URL-string-only:

```go
type SyncRequest struct {
    Direction Direction // LocalToFTP or FTPToLocal
    LocalPath string
    FTP FTPConfig
    RemotePath string

    DeleteExtraneous bool
    ForceChecksum bool
    ChecksumAlgorithm string
    ChunkSize int64
    MaxTransferRate int64
    DryRun bool
    Ignore []string

    Background BackgroundOptions // debounce, workers, cron/poll interval
    Logger Logger                // optional adapter, default no-op
}
```

Keep `core.VFS` compatibility behind the facade only while migrating. Consumers should never need to build `ftp://...?ftp_user=...` strings.

## Package Disposition

| Package / Area | v2.0 Action | Reason | Notes |
|---|---:|---|---|
| `driver/ftp` | **Keep, move under `internal/driver/ftp` or wrap** | Owns real FTP protocol behavior and current tests | Keep plain FTP only; keep passive mode, timeout, encoding, reconnect, metadata handling. |
| `driver/driver.go` | **Keep, narrow** | Generic driver contract is useful for FTP sync adapters | Remove methods not supportable/needed for FTP if simplification is safe; otherwise keep to avoid regression. |
| `sync/disk_sync.go` | **Keep, move under `internal/sync`** | Core comparison/copy engine | Preserve metadata/checksum behavior before refactoring. |
| `sync/driver_push_client_sync.go`, `sync/driver_pull_client_sync.go` | **Keep** | Existing disk↔remote adapter seam | These become the main engine for LocalToFTP and FTPToLocal. |
| `sync/ftp_push_client_sync.go`, `sync/ftp_pull_client_sync.go` | **Keep, rename internally** | Thin FTP-specific wiring already matches target scope | Rename later to `local_to_ftp.go` / `ftp_to_local.go` if helpful. |
| `sync/sync.go` | **Keep but narrow** | Factory becomes disk↔FTP-only | Delete branches for RemoteDisk, SFTP, MinIO. |
| `monitor/fsnotify_monitor.go`, `monitor/base_monitor.go` | **Keep, simplify** | Needed for persistent local-source sync | Remove task/remote-server assumptions and polling busy-wait where feasible. |
| `monitor/ftp_pull_client_monitor.go` | **Keep** | Needed for persistent FTP-source sync | FTP has no event stream; require cron/poll interval for background FTP→local. |
| `monitor/monitor.go` | **Keep but narrow** | Factory becomes disk source or FTP source only | Delete remote disk/task/SFTP/MinIO routing. |
| `core/vfs.go`, `core/path.go`, `core/size.go`, `core/duration.go` | **Keep internally, then reduce** | Useful existing endpoint/path parsing | Replace public VFS strings with typed API options; retain conversion internally during first extraction. |
| `conf/` | **Wrap then delete/reduce** | CLI/YAML config is not the public contract | Keep only if needed to map old semantics into `SyncRequest`; avoid exposing `conf.Config`. |
| `ignore/`, `retry/`, `logger/`, `result/`, `wait/`, `internal/rate` | **Keep/minimize** | Needed by sync and monitor lifecycle | Replace no-src-specific public leakage with tiny interfaces where possible. |
| `report/`, `eventlog/` | **Optional wrap or delete** | Library should return results/events, not run a web report subsystem | Keep only simple counters/events surfaced through `SyncResult` or callback. |
| `cmd/`, `flag/`, `daemon/` | **Delete or move to examples/cmd if retained** | No required CLI runtime in v2.0 | A demo CLI may exist but must import the library rather than being the architecture center. |
| `server/`, `api/`, `auth/` | **Delete** | HTTP/gRPC/file-server/task/auth surfaces are out of scope | Removes major dependency and security surface. |
| `driver/sftp`, `driver/minio`, `sync/sftp_*`, `sync/minio_*`, remote/push HTTP sync files | **Delete** | Non-FTP protocols are out of scope | Keep only test fixtures temporarily if needed for behavioral comparison. |
| `integration/` | **Keep, prune to FTP** | Regression safety | Convert CLI fixture tests into direct library tests. |

## Data Flow

### One-shot Local → FTP

```text
FTPSyncService.SyncOnce(ctx, req)
  → validate typed request: LocalToFTP, local path, FTP host/port/user/pass/passive/timeout, remote path
  → normalize to internal disk endpoint + FTP endpoint
  → build retry/logger/ignore options
  → construct FTP push syncer
  → Connect FTP
  → SyncOnce(local root or requested path)
  → Close FTP
  → return SyncResult{files, bytes, skipped, errors}
```

### One-shot FTP → Local

```text
FTPSyncService.SyncOnce(ctx, req)
  → validate typed request: FTPToLocal and local destination
  → construct FTP pull syncer
  → Connect FTP
  → WalkDir/Open/Stat remote tree through driver/ftp
  → materialize via diskSync
  → Close FTP
  → return SyncResult
```

### Background Local → FTP

```text
FTPSyncService.Start(ctx, req)
  → construct FTP push syncer
  → start fsnotify monitor on local source
  → debounce/coalesce writes in base monitor
  → apply create/write/remove/rename to FTP
  → Stop(ctx) closes monitor, syncer, driver
```

### Background FTP → Local

```text
FTPSyncService.Start(ctx, req)
  → require Background.PollInterval or CronSpec
  → construct FTP pull syncer
  → schedule repeated remote walk/sync
  → no event-stream promise; FTP is polling-only
  → Stop(ctx) cancels scheduler and closes FTP driver
```

## Integration Strategy

1. **Add facade before deleting packages.** Implement `FTPSyncService` using current `conf/core/sync/monitor` wiring internally. This proves the API without destabilizing behavior.
2. **Move internals behind `internal/`.** Once facade tests pass, move or hide packages that should not be imported by consumers.
3. **Narrow factories.** Remove non-FTP branches from sync/monitor factories after tests prove disk↔FTP paths still route correctly.
4. **Prune dependencies.** Delete HTTP/gRPC/SFTP/MinIO/task/CLI packages and run `go mod tidy`; expected large removals include Gin, gRPC, MinIO, SFTP, OAuth, QUIC, sessions, etc.
5. **Refactor config last.** Typed `SyncRequest` should be authoritative, but internal VFS conversion can survive until package deletion is safe.

## Build Order

1. **API contract phase**
   - Define `FTPSyncService`, `SyncRequest`, `FTPConfig`, `Direction`, `BackgroundSync`, `SyncResult`.
   - Add validation: exactly one FTP endpoint and one local endpoint; passive FTP only; timeout parse; remote path required.

2. **Facade adapter phase**
   - Map `SyncRequest` to current internal options.
   - Add direct one-shot Local→FTP and FTP→Local tests against fake driver and local temp dirs.

3. **Background lifecycle phase**
   - Wrap monitor start/stop into `BackgroundSync`.
   - Add context cancellation and deterministic shutdown assertions.

4. **Package pruning phase**
   - Delete/isolate CLI, HTTP/gRPC, task, SFTP, MinIO, auth/session, daemon.
   - Narrow `core.VFSType` to `Disk`, `FTP`, `Unknown`.
   - Narrow `sync.NewSync` and `monitor.NewMonitor` to FTP-only cases.

5. **Dependency and docs phase**
   - Run `go mod tidy` after pruning.
   - Update README to library usage examples, not CLI-first usage.

## Tests Needed to Avoid Regressions

### API and validation
- `NewFTPSyncService` works with zero-value optional dependencies.
- Reject missing local path, missing remote path, missing FTP host/user, invalid port, invalid timeout.
- Reject unsupported direction and FTP↔FTP or local↔local requests.
- Verify typed options produce the same semantics as current CLI/VFS FTP config: host, port, username, password, passive mode, timeout, encoding, local path, remote path.

### One-shot sync
- Local→FTP create/write/remove/rename path mapping preserves existing `ftpPushClientSync` behavior.
- FTP→Local remote walk/open/stat materializes files and directories correctly.
- Existing metadata skip behavior remains: size + precise modtime skips rewrite; ambiguous FTP time forces safe rewrite.
- Dry-run performs no remote/local mutations.
- Ignore rules still filter files before sync.

### Background sync
- Local source uses fsnotify and propagates create/write/remove/rename to FTP.
- FTP source requires poll/cron; starting persistent FTP→local without schedule returns a clear error.
- `Stop(ctx)` shuts down monitor goroutines and closes FTP connection.
- Race-enabled test for burst writes and shutdown during active transfer.

### FTP driver
- Connect requires username and passive mode.
- Timeout parse and dial option behavior.
- Reconnect-on-lost-transport retries idempotent operations only once after reconnect.
- Path encoding modes: auto, utf8, gbk, invalid fallback.
- Unsupported operations return stable errors: symlink/readlink/chtime if server lacks support.

### Pruning/build safety
- `go test ./...` passes after deleting non-FTP packages.
- `go list -deps ./...` no longer includes Gin, gRPC, MinIO, SFTP, QUIC, OAuth, session stores, or protobuf packages unless intentionally retained by examples.
- Public API compatibility check: only intended package exports are importable by external consumers.

## Key Pitfalls

- **Leaking old config/API types:** If `conf.Config`, `core.VFS`, or `sync.Option` remain public guidance, v2.0 will still feel like the old CLI app. Hide them behind typed requests.
- **Deleting before facade tests exist:** The existing sync behavior is valuable; create facade regression tests first, then prune.
- **Pretending FTP is event-driven:** Background FTP→local must be scheduled/polled. Only local→FTP can be fsnotify-driven.
- **Keeping server/auth packages “just in case”:** They dominate dependency and security surface and are explicitly out of scope.
- **Over-refactoring `diskSync` early:** It is large and proven. Move/narrow first; algorithmic cleanup later.

## Roadmap Implications

Suggested phase structure:

1. **Public service API + adapter** — establish `FTPSyncService` while old internals still compile.
2. **One-shot library sync** — validate Local→FTP and FTP→Local without CLI.
3. **Background library sync** — wrap monitor lifecycle and polling semantics.
4. **Aggressive package pruning** — remove non-FTP runtime surfaces and tidy dependencies.
5. **Regression hardening/docs** — real FTP integration tests, race tests, README library examples.

## Sources

- `.planning/PROJECT.md` — HIGH confidence milestone scope.
- `.planning/codebase/ARCHITECTURE.md` — HIGH confidence current layer map.
- `.planning/codebase/STRUCTURE.md` — HIGH confidence package inventory and hotspots.
- `.planning/codebase/CONCERNS.md` — HIGH confidence risks driving pruning and lifecycle tests.
- Current code inspection: `driver/ftp`, `sync/ftp_*`, `monitor/ftp_pull_client_monitor.go`, `core/vfs.go`, `sync/sync.go`, `monitor/monitor.go` — HIGH confidence.
