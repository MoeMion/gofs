# Phase 06 Pattern Map

## Target Files and Closest Analogs

### `ftpsync/service.go`
- **Role:** public API orchestration boundary
- **Closest analog:** existing `ftpsync/service.go` plus Phase 05 summaries
- **Pattern to preserve:** keep public methods context-aware, typed, and isolated from CLI/server runtime

Relevant excerpt:

```go
func (s *FTPSyncService) SyncOnce(ctx context.Context) (Result, error) {
	if err := s.Validate(); err != nil {
		return Result{}, err
	}
	if err := validateContext(ctx, "SyncOnce"); err != nil {
		return Result{}, err
	}
	return Result{}, unsupportedMethod("SyncOnce", s.opts.Direction)
}
```

### `ftpsync/oneshot.go` (new)
- **Role:** library-local adapter from typed public options into legacy sync execution
- **Closest analogs:** `sync/ftp_push_client_sync.go`, `sync/ftp_pull_client_sync.go`, `sync/sync.go`
- **Pattern to preserve:** build internal state once, then route by source/destination capability rather than creating a new protocol engine

Relevant excerpt from `sync/sync.go`:

```go
if source.IsDisk() && dest.Is(core.FTP) {
	return NewFTPPushClientSync(opt)
} else if source.Is(core.FTP) && dest.IsDisk() {
	return NewFTPPullClientSync(opt)
}
```

### `sync/ftp_push_client_sync.go`
- **Role:** existing local→FTP constructor path
- **Pattern:** thin FTP wrapper over generic driver-backed push sync

```go
s := &ftpPushClientSync{
	driverPushClientSync: newDriverPushClientSync(*ds, dest.RemotePath().Base()),
	remoteAddr:           dest.Addr(),
}
s.driver = newFTPPushDriver(s.remoteAddr, dest.FTPConfig(), true, r, maxTranRate, logger)
```

### `sync/ftp_pull_client_sync.go`
- **Role:** existing FTP→local constructor path
- **Pattern:** thin FTP wrapper over generic driver-backed pull sync with FTP metadata overrides

```go
s := &ftpPullClientSync{
	driverPullClientSync: newDriverPullClientSync(*ds),
	remoteAddr:           source.Addr(),
}
s.driver = newFTPPullDriver(s.remoteAddr, source.FTPConfig(), true, r, maxTranRate, logger)
s.diskSync.sourceAbsPath = source.RemotePath().Base()
s.diskSync.isDirFn = s.IsDir
s.diskSync.statFn = s.driver.Stat
s.diskSync.getFileTimeFn = s.driver.GetFileTime
```

### `sync/driver_push_client_sync.go`
- **Role:** existing per-path local→remote mutation logic
- **Pattern:** create/write/symlink/remove operations are already factored per path, which makes them suitable for a library-side best-effort walker

```go
func (s *driverPushClientSync) Create(path string) error { ... }
func (s *driverPushClientSync) Write(path string) error { ... }
func (s *driverPushClientSync) SyncOnce(path string) error {
	return filepath.WalkDir(absPath, func(currentPath string, d fs.DirEntry, err error) error {
		return s.syncWalk(currentPath, d, s, fsutil.Readlink)
	})
}
```

### `sync/driver_pull_client_sync.go`
- **Role:** existing per-path remote→local mutation logic
- **Pattern:** FTP pull semantics already prefer driver metadata and can also be used by a library-side best-effort walker

```go
func (s *driverPullClientSync) write(path, dest string) error {
	sourceModTime := sourceStat.ModTime()
	if _, _, mTime, err := s.getFileTimeFn(path); err == nil && !mTime.IsZero() {
		sourceModTime = mTime
	}
	if s.hash.QuickCompare(..., sourceModTime, destStat.ModTime()) {
		return nil
	}
}
```

### `core/vfs.go`
- **Role:** internal endpoint contract to reuse without exposing publicly
- **Pattern:** `core.NewDiskVFS` for local roots and `core.NewVFS("ftp://...")` for FTP endpoints

Relevant excerpt:

```go
func NewDiskVFS(path string) VFS { ... }
func NewVFS(path string) VFS { ... }
func (vfs *VFS) HasLocalPath() bool { ... }
```

## Planning Implications

1. Put library-only orchestration in `ftpsync`, not `sync`.
2. Reuse existing FTP driver-backed mutation methods rather than rewriting transfer logic.
3. Preserve the cwd-safety rule proven by `HasLocalPath()` and the debug regression note.
4. Keep tests consumer-facing in `ftpsync_test` where possible, with internal/package-local seams only where required for deterministic failure injection.
