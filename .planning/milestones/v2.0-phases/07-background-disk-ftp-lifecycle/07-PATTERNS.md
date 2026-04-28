# Phase 7: Background Disk→FTP Lifecycle - Patterns

**Date:** 2026-04-27
**Status:** Complete

## Target Files

### `ftpsync/service.go`
- **Role:** public API boundary and method dispatch.
- **Closest analog:** existing `SyncOnce(ctx)` dispatch in the same file.
- **Pattern to follow:** validate service/options first, validate context second, then dispatch to a library-local implementation helper.

Relevant excerpt:

```go
func (s *FTPSyncService) SyncOnce(ctx context.Context) (Result, error) {
	if err := s.Validate(); err != nil {
		return Result{}, err
	}
	if err := validateContext(ctx, "SyncOnce"); err != nil {
		return Result{}, err
	}
	return executeSyncOnce(ctx, s)
}
```

### `ftpsync/background.go`
- **Role:** new library-local background runner.
- **Closest analogs:**
  - `ftpsync/oneshot.go` for library-owned orchestration around the legacy sync engine.
  - `monitor/fsnotify_monitor.go` for recursive watch setup and fsnotify event classes.
  - `monitor/base_monitor.go` for burst coalescing/debounce concepts.
- **Pattern to follow:** own the orchestration in `ftpsync`, reuse legacy sync internals for transfer behavior, avoid importing the old monitor runtime wholesale.

Relevant excerpts:

From `monitor/fsnotify_monitor.go`:

```go
func (m *fsNotifyMonitor) monitor(dir string) (err error) {
	dir, err = filepath.Abs(dir)
	if err != nil {
		return err
	}
	err = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			m.watcher.Remove(path)
			err = m.watcher.Add(path)
		}
		return err
	})
	return err
}
```

From `monitor/base_monitor.go`:

```go
if (wm.count <= 2 && now-wm.last <= time.Second.Nanoseconds()) || (wm.count > 2 && now-wm.last <= 3*time.Second.Nanoseconds()) {
	<-time.After(time.Second)
	go func() {
		m.writeNotify <- struct{}{}
	}()
	continue
}
```

### `ftpsync/background_test.go`
- **Role:** lifecycle and debounce regression coverage.
- **Closest analog:** `ftpsync/oneshot_test.go`.
- **Pattern to follow:** use package-local seams to avoid real FTP infrastructure and verify public behavior with deterministic fakes.

Relevant excerpt:

```go
defer withSyncBuilder(buildLegacySync)()
buildLegacySync = func(opt legacysync.Option) (legacysync.Sync, error) {
	captured = &recordingLegacySync{source: opt.Source, dest: opt.Dest}
	return captured, nil
}
```

## Interfaces Executors Should Reuse

### Existing public handle surface

```go
type Handle interface {
	Done() <-chan struct{}
	Err() error
	Stop(context.Context) error
}
```

Phase 7 should extend this with wait semantics rather than replacing the concept.

### Legacy sync interface

```go
type Sync interface {
	Create(path string) error
	Symlink(oldname, newname string) error
	Write(path string) error
	Remove(path string) error
	Rename(path string) error
	Chmod(path string) error
	IsDir(path string) (bool, error)
	SyncOnce(path string) error
	Source() core.VFS
	Dest() core.VFS
	Close()
}
```

### Retry interface

```go
type Retry interface {
	Do(f func() error, desc string) wait.Wait
	DoWithContext(ctx context.Context, f func() error, desc string) wait.Wait
	Count() int
	WaitTime() time.Duration
}
```

Phase 7 should prefer `DoWithContext` where background cancellation must interrupt retry waits.

## File Ownership Guidance

- `07-01-PLAN.md`: `ftpsync/service.go`, `ftpsync/background.go`, `ftpsync/background_test.go`, `ftpsync/context_test.go`
- `07-02-PLAN.md`: `ftpsync/background.go`, `ftpsync/background_test.go`
- `07-03-PLAN.md`: `ftpsync/background.go`, `ftpsync/background_test.go`, `ftpsync/context_test.go`

Because `background.go` is central, later plans should depend on earlier ones instead of running in parallel.
