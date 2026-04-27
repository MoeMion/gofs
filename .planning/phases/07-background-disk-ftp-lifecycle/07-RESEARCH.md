# Phase 7: Background Disk→FTP Lifecycle - Research

**Date:** 2026-04-27
**Status:** Complete

## Research Question

How should `ftpsync.FTPSyncService.StartBackground(ctx)` implement persistent `disk→FTP` synchronization with initial catch-up, debounced local change handling, observable lifecycle control, and deterministic shutdown while keeping `FTP→disk` background mode unsupported in v2.0?

## Inputs Read

- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/STATE.md`
- `.planning/phases/07-background-disk-ftp-lifecycle/07-CONTEXT.md`
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-01-SUMMARY.md`
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-02-SUMMARY.md`
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-03-SUMMARY.md`
- `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-01-SUMMARY.md`
- `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-02-SUMMARY.md`
- `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-03-SUMMARY.md`
- `ftpsync/service.go`
- `ftpsync/oneshot.go`
- `ftpsync/errors.go`
- `ftpsync/hooks.go`
- `ftpsync/context_test.go`
- `ftpsync/oneshot_test.go`
- `monitor/fsnotify_monitor.go`
- `monitor/base_monitor.go`
- `monitor/option.go`
- `sync/sync.go`
- `sync/option.go`
- `sync/driver_push_client_sync.go`
- `sync/ftp_push_client_sync.go`
- `driver/ftp/ftp.go`
- `retry/retry.go`

## Key Constraints From Context

### Locked Decisions

- **D-01:** Background `disk→FTP` performs an initial sync before watch mode.
- **D-02:** Prioritize remote catch-up correctness over fastest watch startup.
- **D-03:** Single sync failures are reported but do not terminate the background run by default.
- **D-04:** Background mode behaves like a long-running service, not a fail-fast batch job.
- **D-05:** Local filesystem events use debounce/coalescing.
- **D-06:** Avoid duplicate uploads during bursts; do not react independently to every raw event.
- **D-07:** Public entry remains `StartBackground(ctx)`.
- **D-08:** Extra lifecycle control lives on the returned handle.
- **D-09:** Handle must support active stop, wait-for-exit, and final/current error access.
- **D-10:** Handle semantics must work for long-lived embedders.
- **D-11:** `disk→FTP` only; no `FTP→disk` background polling.
- **D-12:** Public API remains typed Go options in package `ftpsync`.
- **D-13:** Phase 6 one-shot result/error semantics stay compact.

## Existing Patterns Worth Reusing

### 1. Public API boundary in `ftpsync`

- `FTPSyncService` already owns typed options, hooks, validation ordering, and public error classification.
- `SyncOnce(ctx)` already performs validation and context checks before running library-local orchestration in `ftpsync/oneshot.go`.
- `StartBackground(ctx)` must follow the same pattern: validate, guard unsupported directions, then dispatch to a library-local implementation.

### 2. One-shot orchestration as the initial catch-up engine

- `executeSyncOnce(ctx, svc)` already builds the approved legacy adapter and performs local→FTP work.
- Reusing `SyncOnce` or a close sibling helper for initial catch-up is the lowest-risk way to satisfy D-01 and D-02.
- Reusing one-shot orchestration also preserves the compact hook and error behavior established in Phase 6.

### 3. Existing monitor package gives debounce and fsnotify ideas, but is too broad to import wholesale

- `monitor/fsnotify_monitor.go` already shows the legacy project pattern for translating fsnotify events into `Create`, `Write`, `Remove`, `Rename`, and `Chmod` calls.
- `monitor/base_monitor.go` already demonstrates burst coalescing via delayed write processing and retry-aware write handling.
- However, `monitor.Option` depends on runtime-only packages (`auth`, `conf`, `report`, legacy logger/report/eventlog surfaces), which conflicts with the focused Phase 7 library boundary.
- Conclusion: reuse the conceptual patterns, not the `monitor` package directly.

### 4. Legacy FTP push sync already has reconnect-aware write semantics

- `sync.NewSync` dispatches `disk→FTP` to `NewFTPPushClientSync`.
- `driverPushClientSync.Write` and `ftpDriver.Connect/reconnectIfLost` already own FTP write behavior, reconnects, and passive-mode enforcement.
- Phase 7 should not implement raw FTP transfer logic or connection management itself; it should keep delegating to the existing sync engine.

## Recommended Architecture

## Summary

Implement a library-local background runner in `ftpsync` rather than importing the legacy `monitor` package. The runner should:

1. Reject any direction except `DirectionLocalToFTP`.
2. Run an initial `SyncOnce`/shared local→FTP sync path before starting the watcher.
3. Create an `fsnotify.Watcher` scoped to the configured local source root.
4. Recursively watch existing directories and newly created directories.
5. Coalesce incoming file events into a debounce-triggered sync request queue.
6. Serialize actual sync runs so only one upload cycle is active at a time.
7. Reuse the existing local→FTP execution helper for each coalesced sync cycle.
8. Surface lifecycle state through a returned handle with `Done()`, `Stop(ctx)`, `Wait()`, and `Err()`/`CurrentError()` style access.
9. Shut down watcher, goroutines, timers, and any in-flight/retry work when context is canceled or the handle is stopped.

## Why this is the best fit

- Honors D-07/D-08 by keeping `StartBackground(ctx)` as the only public entry and pushing lifecycle control to the handle.
- Honors D-01/D-02 by performing initial catch-up before watch mode.
- Honors D-03/D-04 by recording sync failures and continuing operation instead of terminating the run.
- Honors D-05/D-06 by batching noisy fsnotify events into a delayed sync request instead of uploading every raw event.
- Preserves the library boundary because the new code lives in `ftpsync` and only depends on `fsnotify`, existing sync internals, and standard library concurrency primitives.

## Recommended Internal Types

### Public handle evolution

Current public interface in `ftpsync/service.go`:

```go
type Handle interface {
	Done() <-chan struct{}
	Err() error
	Stop(context.Context) error
}
```

This is too weak for D-09 because it lacks wait semantics and only exposes one undifferentiated error accessor.

Recommended Phase 7 contract:

```go
type Handle interface {
	Done() <-chan struct{}
	Err() error
	Wait() error
	Stop(context.Context) error
}
```

Implementation detail can provide more methods on the concrete type if helpful, but `Wait()` should be added to the public interface because D-09 explicitly requires wait-for-exit semantics.

### Internal runner shape

Recommended internal `backgroundHandle` / runner state:

- `ctx context.Context`
- `cancel context.CancelFunc`
- `done chan struct{}`
- `ready chan struct{}` or internal once/flag for readiness signaling
- `watcher *fsnotify.Watcher`
- `sourceRoot string`
- `debounce time.Duration`
- `trigger chan struct{}` with capacity 1 for coalesced sync requests
- `lastErr atomic/pointer or mutex-protected error`
- `waitErr error` mutex-protected final error
- `wg sync.WaitGroup`
- `stopOnce sync.Once`

## Event Handling Strategy

### Recommended debounce model

Use a single buffered trigger channel plus a resettable timer:

1. Any fsnotify create/write/remove/rename-worthy event marks the run as dirty.
2. A debounce timer is reset on each new event within the window.
3. When the timer fires, enqueue a single sync request if one is not already pending.
4. The sync worker performs one full local→FTP sync pass over the configured source root.

This is more appropriate than trying to replay individual create/write/remove events because:

- The approved public library abstraction already has a robust one-shot local→FTP pass.
- Rename and burst correctness are easier to preserve with a single best-effort rescan than with per-event remote mutation logic.
- D-06 favors avoiding duplicates over reacting to every raw event.

### Why not mirror the legacy monitor event-by-event engine exactly?

- The library does not need the full old runtime's eventlog/reporter/task plumbing.
- Raw fsnotify event streams are platform-noisy and often redundant.
- A coalesced "resync root" loop is simpler, easier to verify, and aligned with D-05/D-06.

## Initial Sync and Subsequent Sync Relationship

Recommended execution model:

- **Initial run:** call the same local→FTP sync helper used by `SyncOnce`, not a different path.
- **Steady-state runs:** call that same helper again after each coalesced burst.

This keeps background mode as “initial sync + repeated one-shot local→FTP passes,” exactly matching the intent captured in context.

## Error Model Recommendation

### Non-terminal sync errors

For D-03/D-04:

- A sync pass failure should be stored as the handle’s current/latest error.
- The error should be reported through hooks (`log`/`reportEvent`) with `ErrTransfer`, `ErrConnection`, or `ErrAuthentication` as appropriate.
- The runner should continue listening for future events unless the root context is canceled or the handle is explicitly stopped.

### Terminal errors

The background run should terminate only on setup/lifecycle failures such as:

- invalid service or unsupported direction
- nil/canceled context before startup
- failure to compute source root path
- failure to construct the fsnotify watcher
- unrecoverable watcher setup failure before the runner becomes active

Once running, transient sync failures should be non-terminal.

## Shutdown Design Recommendation

To satisfy WATCH-04, shutdown must explicitly cover:

- watcher close
- all goroutine exits
- timer stop
- trigger channel unblocking semantics
- in-flight sync loop cancellation
- retry sleep interruption where possible by using `context.Context` in the library-level loop
- final `done` channel close and `Wait()` unblocking

### Important nuance about retry sleeps

The legacy retry interface supports `DoWithContext(ctx, f, desc)`.
If the Phase 7 background path needs retry behavior around repeated sync cycles or watcher setup, it should prefer `DoWithContext` over `Do`, because WATCH-04 explicitly requires retry sleeps to stop when context is canceled.

## Unsupported Capability Guard

Phase 7 must continue to reject:

- `DirectionFTPToLocal` in `StartBackground`
- any bidirectional mode
- any implied background polling API separate from `StartBackground(ctx)`

Tests should assert that `StartBackground(context.Background())` for `DirectionFTPToLocal` still returns `ErrUnsupportedCapability` with method/direction context.

## Testing Implications

High-value automated coverage for this phase:

1. `StartBackground` local→FTP performs initial sync before watcher readiness.
2. local write/create/remove/rename burst produces at least one subsequent sync cycle after debounce.
3. repeated sync failures update handle error but do not terminate the run.
4. `Stop(ctx)` causes `Done()` to close and `Wait()` to return.
5. canceling root context causes deterministic shutdown.
6. newly created directories are added to watcher coverage.
7. `FTP→local` `StartBackground` remains unsupported.

Tests do not need a real FTP server if seams are added around the background sync invocation and watcher/event source. Use the same testing style already present in `ftpsync/oneshot_test.go`.

## Recommended File-Level Plan Shape

Likely implementation files:

- `ftpsync/service.go` — update `Handle` interface and route `StartBackground` into implementation.
- `ftpsync/background.go` — runner, watcher setup, debounce loop, lifecycle handle.
- `ftpsync/background_test.go` — background lifecycle, debounce, shutdown, and unsupported-direction tests.
- `ftpsync/context_test.go` — update public contract expectations now that local→FTP background is implemented.

## Common Pitfalls To Avoid

1. **Importing `monitor.Option` or `monitor.NewMonitor` directly**
   - Pulls in config/runtime concerns the library is supposed to avoid.

2. **Per-event FTP mutation logic as the primary background path**
   - Harder to make portable and deduplicated than a coalesced resync loop.

3. **Terminating the runner on first sync error**
   - Violates D-03/D-04.

4. **Leaving `Wait()` out of the public handle**
   - Violates D-09.

5. **Using retry helpers without context cancellation**
   - Risks violating WATCH-04’s retry-sleep shutdown requirement.

6. **Starting watcher before initial catch-up completes**
   - Conflicts with D-01/D-02.

## Final Recommendation

Build a small `ftpsync/background.go` runner that reuses the approved one-shot local→FTP execution path, watches the source tree with `fsnotify`, debounces event bursts into a single follow-up sync request, keeps failures observable but non-fatal, and exposes a stronger handle contract with `Done`, `Err`, `Wait`, and `Stop`.

This is the smallest correct change set that satisfies WATCH-01 through WATCH-05 without reviving the legacy monitor runtime.

## Validation Architecture

- **Primary test framework:** `go test`
- **Fast verification target:** `go test ./ftpsync -run 'Test(StartBackground|BackgroundHandle|BackgroundLifecycle)' -count=1`
- **Full phase verification target:** `go test ./ftpsync -count=1`
- **Phase-specific sampling target:** background lifecycle tests should run after each task because the feature is concurrency-heavy and regressions are likely to be behavioral rather than compile-only.
