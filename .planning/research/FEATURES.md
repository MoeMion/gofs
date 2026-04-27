# Feature Landscape: v2.0 FTP Sync Library API

**Domain:** Go package API for FTP file synchronization  
**Project:** gofs  
**Researched:** 2026-04-27  
**Scope:** Capabilities `FTPSyncService` should expose for one-shot and persistent disk↔FTP sync after the CLI/server runtime is removed or isolated.  
**Overall confidence:** HIGH for project/API needs from current planning artifacts; MEDIUM for exact implementation cost until code-level extraction planning.

## Recommendation Summary

v2.0 should expose a focused **library-first FTP sync service**, not a thin wrapper around the old CLI. The public API should preserve the current FTP CLI semantics users already validated in v1.0, but present them as explicit Go options and lifecycle methods.

The table-stakes API is: construct an `FTPSyncService`, pass source/destination disk/FTP endpoints, run **one-shot push/pull**, or start a **background monitor** with a returned handle that supports readiness, error reporting, and shutdown. Keep the package intentionally FTP-only: disk↔FTP is supported; SFTP, MinIO, HTTP/gRPC server, task mode, file server, daemon mode, and CLI-specific behavior should not appear in the public surface.

## Table Stakes for v2.0

Features consumers will expect from a usable Go FTP sync package. These should define the v2.0 roadmap.

| Feature | Why Expected | Complexity | Implementation Notes |
|---------|--------------|------------|----------------------|
| Public `FTPSyncService` type | Milestone goal names this as the primary API | Medium | Provide constructor(s) that validate options before runtime starts. Avoid requiring CLI flag parsing. |
| One-shot disk → FTP sync | Required push path; v1.0 already validated FTP as destination | Medium | API should expose `SyncOnce`/`PushOnce` style call that blocks until completion and returns a structured result/error. |
| One-shot FTP → disk sync | Required pull path; v1.0 already validated FTP as source | Medium | Same semantics as push, but with FTP endpoint as source and disk endpoint as destination. |
| Background disk → FTP monitoring | Library consumers need persistent upload sync without running the CLI | Medium-High | Reuse local filesystem monitor behavior where possible; return a lifecycle handle. |
| Background FTP → disk polling/monitoring | Project explicitly requires persistent monitoring/sync; FTP source cannot emit native events | Medium-High | Use existing driver-backed polling/cron-style monitor semantics; reject configs with no polling/cron interval. |
| Endpoint option shape equivalent to CLI semantics | Prevents behavior drift from shipped v1.0 | Medium | Cover disk path plus FTP host, port, username, password, remote path, passive mode, timeout, and path encoding controls. |
| Direction validation | Library users need immediate feedback for unsupported combinations | Low | Support only disk↔FTP. Reject FTP↔FTP, disk↔disk if outside scope, non-FTP URLs, missing endpoint paths, and ambiguous direction. |
| Config struct plus functional options | Go consumers expect typed configuration and ergonomic overrides | Medium | Prefer `FTPSyncOptions` as canonical; optional `WithLogger`, `WithRetry`, `WithReportHook`, etc. Keep YAML/JSON loading optional. |
| Context-aware execution | Go APIs should integrate with cancellation and timeouts | Medium | `SyncOnce(ctx, opts)` and `Start(ctx, opts)` should respect cancellation; background handle should also expose `Stop/Shutdown`. |
| Lifecycle handle for persistent runs | Existing architecture already uses wait/result handles; library needs this without CLI signals | Medium | Return handle with `WaitReady`, `Wait`, `Shutdown`, `Err`, and possibly `Report`/`Stats`. Do not rely on OS signal registration. |
| Structured errors | Package consumers need inspectable failures, not log-only behavior | Medium | Return validation, connection, auth, transfer, unsupported-combination, and cancellation errors. Preserve `%w` wrapping. |
| Progress/reporting callbacks | Existing reporting is a core operational concern; library users need observability without HTTP manage UI | Medium | Expose optional callbacks/events for file started/completed/skipped/deleted/error and aggregate stats. |
| Logging injection | Libraries should not force console/file logging side effects | Low | Accept a small logger interface or callback; default to no-op logger. |
| Ignore/filter support | Existing sync workflows depend on ignore rules; package should not regress practical sync behavior | Medium | Expose ignore paths/patterns or a predicate callback if current ignore package can be reused cleanly. |
| Retry and timeout controls | FTP is brittle; v1.0 already requires timeout compatibility | Medium | Provide sane defaults while allowing retry count/delay and FTP operation timeout customization. |
| Documentation with minimal examples | A Go package needs discoverable import and usage patterns | Low | Include examples for push once, pull once, persistent push, persistent pull, cancellation, and reporting callbacks. |
| Unit and integration tests for library API | Extraction can silently break previously validated CLI behavior | Medium | Test public API directly, not only internal sync constructors. Preserve real-server FTP integration coverage. |

## Differentiators / Future Candidates

Useful capabilities after the v2.0 package boundary is stable.

| Feature | Value Proposition | Complexity | Defer Rationale |
|---------|-------------------|------------|-----------------|
| YAML/JSON config file compatibility helper | Helps migrate old CLI users | Medium | Useful, but canonical package API should be typed options first. |
| Convenience direction-specific methods (`PushOnce`, `PullOnce`, `WatchPush`, `WatchPull`) | Improves ergonomics over one generic method | Low-Medium | Add if it does not duplicate too much API; generic config can ship first. |
| Transfer progress byte counters | Better UX for large files | Medium | Requires careful instrumentation around current copy paths. File-level callbacks are enough for v2.0. |
| Resumable transfers | Valuable for large files and flaky servers | High | Requires partial-state handling and stronger FTP restart semantics. |
| FTPS support | Security-sensitive users will ask for it | High | Explicitly out of current project scope. |
| Advanced FTP compatibility knobs | Helps legacy servers | Medium | Add only after real package consumers report specific interop issues. |
| Bidirectional sync/conflict resolution | Higher-level product value | High | Not part of current one-way disk↔FTP semantics and would change sync guarantees. |
| FTP↔FTP sync | Some users may expect remote-to-remote support | Medium-High | Current milestone is disk↔FTP; remote-to-remote multiplies edge cases. |
| Metrics exporter | Useful in services | Medium | Callbacks/stats are sufficient; Prometheus/OpenTelemetry should stay adapter-level initially. |

## Anti-Features for v2.0

Features to explicitly keep out of the public library surface.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| CLI flag API as the primary interface | v2.0 intentionally drops CLI runtime; flag-shaped APIs make package use awkward | Translate CLI semantics into typed Go options. |
| HTTP/gRPC/file-server/task/daemon modes | These are non-FTP runtime surfaces slated for removal or isolation | Keep only internal code needed for FTP sync; expose no server APIs. |
| SFTP/MinIO/non-FTP protocols | Violates milestone goal of focused FTP package | Reject non-FTP endpoints at validation. |
| Global logging, process exit, or signal handling | Libraries must not own the host process lifecycle | Return errors/handles and let callers decide. |
| Hidden goroutines without shutdown | Persistent sync must be embeddable safely | Every background start returns a stop/wait handle. |
| FTPS in v2.0 | Adds TLS/cert scope beyond current milestone | Document as deferred. |
| Bidirectional conflict resolution | Much larger than exposing existing one-way sync | Keep explicit source→destination direction. |

## Recommended Public API Capabilities

Implementation can choose exact names, but roadmap phases should cover this shape.

```go
type FTPSyncService struct { /* unexported deps */ }

type FTPSyncOptions struct {
    Source      Endpoint
    Destination Endpoint
    Mode        SyncMode // once or persistent
    FTP         FTPOptions
    Monitor     MonitorOptions
    Retry       RetryOptions
    Ignore      IgnoreOptions
    Logger      Logger
    Reporter    Reporter
}

type FTPOptions struct {
    Host         string
    Port         int
    Username     string
    Password     string
    RemotePath   string
    PassiveMode  bool
    Timeout      time.Duration
    PathEncoding string
}

type RunHandle interface {
    WaitReady(ctx context.Context) error
    Wait() error
    Shutdown(ctx context.Context) error
    Report() SyncReport
}
```

Required behavior:

- `SyncOnce(ctx, options)` blocks until convergence attempt completes.
- `Start(ctx, options)` starts persistent monitoring and returns a `RunHandle`.
- Validation fails before network transfer when source/destination/mode/options are invalid.
- All public methods are safe to call from host applications without process-global side effects.

## Feature Dependencies

```text
Typed endpoint/config model
  → validation for supported disk↔FTP combinations
  → FTPSyncService constructor
  → one-shot push/pull API
  → direct public API tests

One-shot API
  → reporting/error model
  → documentation examples
  → persistent monitoring API

Persistent monitoring API
  → lifecycle handle
  → cancellation/shutdown tests
  → real-server background push/pull integration tests

Runtime surface removal/isolation
  → package import does not expose CLI/server/task/SFTP/MinIO concepts
  → final docs describe package-only usage
```

## MVP Recommendation

Prioritize:

1. **Typed `FTPSyncService` API and options**
   - Equivalent to current FTP CLI semantics, but Go-native.
   - Validate disk↔FTP source/destination combinations early.

2. **One-shot push and pull**
   - `disk → FTP` and `FTP → disk` through direct Go calls.
   - Structured result/error and no CLI/server dependency.

3. **Persistent monitoring with lifecycle handle**
   - Local disk watcher for push.
   - FTP polling/cron-style monitor for pull.
   - Explicit readiness, shutdown, wait, and error behavior.

4. **Observability, docs, and tests**
   - No-op defaults for logger/reporting; optional callbacks for applications.
   - Package examples plus real FTP integration tests for once and background modes.

Defer:

- FTPS/TLS.
- Bidirectional conflict resolution.
- FTP↔FTP sync.
- Resumable transfers.
- Prometheus/OpenTelemetry adapters.
- Recreating CLI/server behaviors as package APIs.

## Documentation Requirements

Minimum docs for v2.0:

- Package overview: what is supported and intentionally unsupported.
- Example: one-shot disk → FTP push.
- Example: one-shot FTP → disk pull.
- Example: persistent disk → FTP monitoring with shutdown.
- Example: persistent FTP → disk polling with cancellation.
- Options reference mapping old CLI/config semantics to `FTPSyncOptions` fields.
- Error/lifecycle guide: validation errors, transfer errors, cancellation, and `RunHandle` use.
- Security note: plain FTP only; credentials are not encrypted in transit; FTPS deferred.

## Test Requirements

Minimum test coverage for v2.0:

- Unit tests for option validation, direction detection, default values, and unsupported combinations.
- Unit tests for lifecycle handle behavior: ready, wait, shutdown, cancellation, error propagation.
- Unit tests for reporter/logger defaults and callback invocation.
- Real-server integration tests for one-shot disk→FTP and FTP→disk.
- Real-server integration tests for persistent disk→FTP and FTP→disk polling.
- Regression tests proving non-FTP endpoints and removed runtime surfaces are not accepted by the public API.
- Race-enabled coverage should continue matching existing repository CI patterns.

## Sources

- `.planning/PROJECT.md` — v2.0 goals, active requirements, out-of-scope boundaries, FTP v1 validated capabilities — HIGH confidence.
- `.planning/MILESTONES.md` — shipped FTP v1 capabilities and integration coverage — HIGH confidence.
- `.planning/codebase/ARCHITECTURE.md` — existing sync, monitor, option, result, report, and runtime-surface boundaries — HIGH confidence.
- `.planning/codebase/TESTING.md` — repository testing conventions and integration-test expectations — HIGH confidence.
