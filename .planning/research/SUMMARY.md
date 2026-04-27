# Project Research Summary

**Project:** gofs
**Domain:** Go library extraction for disk ↔ FTP synchronization
**Researched:** 2026-04-27
**Confidence:** HIGH

## Executive Summary

gofs v2.0 should stop being a CLI/server-centered sync application and become a focused Go package for plain FTP client synchronization. The research is consistent across stack, feature, architecture, and pitfall analysis: experts would preserve the already-working disk↔FTP transfer behavior, wrap it behind a small `FTPSyncService` facade, and remove or isolate unrelated runtime surfaces rather than adding a new framework or rewriting the sync algorithm.

The recommended approach is dependency and surface-area reduction. Keep Go 1.24.4, standard library `context`/filesystem/network primitives, `github.com/jlaffaye/ftp`, `github.com/fsnotify/fsnotify`, and only the internal sync pieces needed for one-shot and background disk↔FTP. Expose typed config structs, context-first methods, lifecycle handles, structured errors, optional logging/report callbacks, and explicit validation for supported endpoint combinations. Do not expose CLI flags, `cmd.RunWithConfig`, HTTP/gRPC/file-server/task behavior, SFTP, MinIO, daemon mode, process signals, or global logging as part of the public API.

The biggest risks are accidentally exporting the old runtime, breaking path semantics during extraction, and losing lifecycle/cancellation control for background sync. Mitigate these by defining the public API first, adding facade tests before deleting packages, requiring explicit local roots, keeping FTP remote paths distinct from local paths, preserving v1 FTP regression coverage, and making every persistent run return a stop/wait handle that closes watchers, tickers, FTP connections, and retry loops on context cancellation.

## Key Findings

### Recommended Stack

v2.0 should mostly remove dependencies, not add them. The package should use the existing Go module/toolchain and proven FTP implementation, while pruning server, multi-protocol, task, UI, and CLI-only infrastructure from the library build.

**Core technologies:**
- Go `1.24.4` — current module target and sufficient runtime for a library-first implementation.
- `context.Context` — first parameter for `SyncOnce`/`Start`; controls cancellation, deadlines, retry sleeps, monitor shutdown, and FTP dial behavior.
- Go standard library `io/fs`, `os`, `path/filepath`, `time`, `net` — enough for traversal, local file operations, timers, and endpoint handling.
- `github.com/jlaffaye/ftp` `v0.2.0` — keep as the plain FTP client; use `DialWithContext`, `DialWithTimeout`, `Login`, `Retr`, `Stor`, `Walk`, `GetTime`, `SetTime`, `Rename`, and `Delete`. Treat `ServerConn` as not concurrency-safe.
- `github.com/fsnotify/fsnotify` — keep for local disk event monitoring in disk→FTP background mode; recursive watches and debounce/coalescing remain library responsibilities.
- Minimal internal helpers — keep ignore, retry, rate limiting, metadata/path encoding, and lightweight logging hooks only where directly needed.

**Remove or isolate:** Gin, gRPC/protobuf/OAuth, SFTP/SSH, MinIO/S3, QUIC, Redis/cache/task loaders, daemon/process startup, progress bars, HTTP report UI, Docker/binary release assumptions, and cron unless legacy cron compatibility is deliberately kept behind an adapter. Use `time.Ticker`/poll intervals as the default background FTP-source scheduler.

### Expected Features

The table-stakes product is a Go package API that preserves validated FTP v1 semantics without preserving the old application surface. Consumers should be able to construct a service, pass typed local/FTP endpoints, run a one-shot push or pull, or start a persistent sync with deterministic shutdown and observable results.

**Must have (table stakes):**
- Public `FTPSyncService` type or interface with constructor validation.
- Typed endpoint/config model covering local path, FTP host, port, username, password, remote path, passive mode, timeout, and path encoding.
- One-shot disk→FTP sync and one-shot FTP→disk sync through context-aware public methods.
- Background disk→FTP monitoring using local filesystem events.
- Background FTP→disk synchronization using explicit polling/cron-style scheduling; FTP must not be described as event-driven.
- Direction validation: support only disk↔FTP; reject FTP↔FTP, local↔local if outside scope, non-FTP protocols, missing paths, and ambiguous directions.
- Lifecycle handle for background runs with readiness/wait/shutdown/error reporting.
- Structured errors for validation, unsupported combinations, connection/auth, transfer, retry, and cancellation.
- Optional progress/report callbacks and logging injection with no-op defaults.
- Ignore/filter support, retry/timeout controls, public API tests, real FTP integration tests, and package examples.

**Should have (useful after core boundary is stable):**
- Compatibility parser/adapter for legacy FTP CLI URL/config examples.
- Convenience methods such as `PushOnce`, `PullOnce`, `WatchPush`, and `WatchPull` if they do not fragment the API.
- File/byte-level progress counters after core callbacks exist.
- Advanced FTP compatibility knobs only when real users report specific server issues.

**Defer beyond v2.0:**
- FTPS/TLS, FTP server support, FTP↔FTP sync, bidirectional/conflict resolution, resumable transfers, Prometheus/OpenTelemetry adapters, progress bars/terminal UI, and recreating old CLI/server behavior as package APIs.

### Architecture Approach

Build a small public facade first, then move/narrow internals. `FTPSyncService` should become the only supported entry point, while current `conf`, `core.VFS`, `sync.NewSync`, and `monitor.NewMonitor` concepts are retained temporarily behind the facade and later moved under `internal/` or reduced to FTP-only cases.

**Major components:**
1. Public package (`ftpsync` or root `gofs`) — exposes `FTPSyncService`, `Config`/`SyncRequest`, endpoint structs, `FTPConfig`, direction/mode enums, `SyncResult`, errors, and `BackgroundSync`/`RunHandle`.
2. Internal config normalization — validates exactly one local endpoint and one FTP endpoint, normalizes absolute local roots, keeps remote FTP paths separate, applies defaults, and rejects unsupported legacy fields.
3. Internal sync engine — preserves disk sync plus generic driver push/pull behavior for Local→FTP and FTP→Local one-shot operations.
4. Internal background monitor — uses fsnotify/debounce for local-source sync and ticker/poll scheduling for FTP-source sync, with context-owned shutdown.
5. Internal FTP driver — owns `jlaffaye/ftp` operations, passive/timeout/encoding behavior, metadata handling, reconnect rules, and serialized access.
6. Observability layer — returns `SyncResult` and emits optional callbacks/events instead of web reports, progress bars, global logs, or process output.

Recommended public shape:

```go
func NewFTPSyncService(cfg Config) (*FTPSyncService, error)
func (s *FTPSyncService) SyncOnce(ctx context.Context) (*SyncResult, error)
func (s *FTPSyncService) Start(ctx context.Context) (*Run, error)
```

The exact names can change, but the contract should not: explicit structs, no stored context, no global process ownership, and deterministic shutdown.

### Critical Pitfalls

1. **Exporting the old runtime instead of a library API** — never wrap `cmd.RunWithConfig`; public package must not import `cmd`, `flag`, `server`, `api`, task, daemon, or process signal code.
2. **Breaking path semantics and cwd safety** — require explicit local roots for disk writes, store absolute local paths internally, preserve the distinction between omitted FTP remote paths and local destinations, and add cwd-sentinel tests.
3. **Losing cancellation and lifecycle ownership** — all public work accepts `context.Context`; every background run returns a handle; shutdown closes goroutines, watchers, tickers, retry sleeps, and FTP connections.
4. **Deleting tests before preserving behavior** — move retained FTP/sync/monitor tests before removing old packages; keep integration, race, no-op, delete/rename, and burst-write coverage.
5. **Treating CLI config compatibility as API stability** — map only FTP-relevant semantics into typed options; reject or isolate obsolete server/auth/SFTP/MinIO/task fields.
6. **Incorrect Go v2 module plan** — decide whether the module path becomes `github.com/no-src/gofs/v2`; align README, examples, CI, release notes, and tags before stable release.
7. **Pruning dependencies before proving compile boundaries** — facade and tests come first; then inspect package graph, delete outside it, run `go mod tidy`, and verify old runtime dependencies disappear.

## Implications for Roadmap

Based on research, suggested phase structure:

### Phase 1: Public Service API Contract
**Rationale:** The API boundary must be fixed before extraction, otherwise the old CLI/server runtime will leak into the library.  
**Delivers:** `FTPSyncService`, typed config/request structs, endpoint/direction model, `FTPConfig`, validation errors, context-first method contracts, and public docs skeleton.  
**Addresses:** Public service API, endpoint option shape, direction validation, context-aware execution, structured errors.  
**Avoids:** Exporting old runtime, CLI-shaped config bloat, package-name confusion.  
**Research flag:** Standard Go library API patterns; no deep research needed, but module path/version decision must be explicitly resolved.

### Phase 2: Facade Adapter for One-Shot Sync
**Rationale:** Prove the new public API can drive existing behavior before deleting or moving packages. This protects the validated FTP v1 sync semantics.  
**Delivers:** `SyncOnce(ctx)` for disk→FTP and FTP→disk using current FTP driver and sync adapters; direct public API tests against fake/temp fixtures; structured `SyncResult`.  
**Uses:** `jlaffaye/ftp`, existing driver push/pull sync, disk sync, ignore/retry helpers.  
**Avoids:** Rewriting proven algorithms too early, silent delete/rename/no-op behavior drift, hidden cwd writes.

### Phase 3: Background Lifecycle and Monitoring
**Rationale:** Persistent sync is table-stakes but has the highest concurrency/lifecycle risk; it should be isolated after one-shot API behavior is stable.  
**Delivers:** `Start(ctx)` plus `RunHandle`/`BackgroundSync`, fsnotify-driven disk→FTP, polling/ticker FTP→disk, readiness/wait/shutdown behavior, cancellation tests, race tests.  
**Addresses:** Background disk→FTP monitoring, background FTP→disk polling, lifecycle handle, deterministic shutdown, reporting callbacks.  
**Avoids:** Pretending FTP is event-driven, goroutine leaks, monitor timing regressions.

### Phase 4: Internalization and Package Pruning
**Rationale:** Once facade behavior is covered, remove obsolete surfaces and reduce dependencies without destabilizing the API.  
**Delivers:** Move/narrow internals under `internal/`, delete or isolate CLI/server/task/auth/SFTP/MinIO/daemon packages, narrow sync/monitor factories to disk↔FTP, simplify config/VFS to internal compatibility only, run `go mod tidy`.  
**Uses:** Package graph verification (`go list`), build/test gates, dependency checks.  
**Avoids:** Hidden imports retaining Gin/gRPC/MinIO/SFTP/QUIC/OAuth/protobuf, placeholder compatibility stubs, public leakage of old config types.  
**Research flag:** Needs phase-level codebase/package-graph research because deletion boundaries are implementation-specific and high blast-radius.

### Phase 5: Regression Hardening, Docs, and Release Alignment
**Rationale:** After pruning, the package needs proof that preserved FTP behavior still works and that consumers see the new library contract, not old CLI instructions.  
**Delivers:** Real FTP integration tests for one-shot and background modes, cwd-sentinel tests, path encoding tests, retry/cancellation tests, package examples, README rewrite, migration notes, security note for plain FTP, and Go module/release workflow alignment.  
**Addresses:** Documentation examples, integration coverage, migration compatibility, security messaging, v2 module/import expectations.  
**Avoids:** Docs describing removed product surfaces, release artifacts implying CLI/container support, unverified module path breaks.  
**Research flag:** Needs focused release/versioning research if stable `v2.0.0` is planned; otherwise standard docs/test patterns.

### Phase Ordering Rationale

- API contract comes first because every later phase either implements, tests, or prunes around that boundary.
- One-shot sync precedes background sync because it validates endpoint normalization, driver wiring, path semantics, result/error handling, and basic transfer behavior without concurrency complexity.
- Background lifecycle comes before pruning because monitor code is fragile and tests must be moved/preserved before old packages disappear.
- Package pruning is delayed until facade tests define the retained behavior and package graph boundaries are known.
- Docs/release finish the sequence because public examples, module path, and removed-feature messaging must reflect the final package shape and dependency set.

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 3: Background Lifecycle and Monitoring** — verify current monitor/coalescing behavior, race risks, and deterministic shutdown strategy before implementation.
- **Phase 4: Internalization and Package Pruning** — requires package graph analysis to avoid deleting shared utilities or keeping hidden runtime dependencies.
- **Phase 5: Regression Hardening, Docs, and Release Alignment** — needs explicit Go module v2/import-path/release workflow decision.

Phases with standard patterns (skip research-phase unless implementation questions arise):
- **Phase 1: Public Service API Contract** — context-first Go APIs, typed options, no-op logger/callback patterns are well documented.
- **Phase 2: Facade Adapter for One-Shot Sync** — primarily maps existing validated code into the new API; deeper research should be code-level, not external-domain research.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | Strong agreement that no framework is needed; official docs cover `context`, `jlaffaye/ftp`, and fsnotify; dependency removals are directly tied to v2 scope. Public package naming remains MEDIUM until accepted. |
| Features | HIGH | Table-stakes features are anchored in current project goals and validated FTP v1 behavior. Exact implementation cost for extraction remains MEDIUM. |
| Architecture | HIGH | Current package boundaries and retained/deleted package disposition are well understood. Final names and internal move sequencing need implementation confirmation. |
| Pitfalls | HIGH | Risks are project-specific and supported by prior v1 bug/debug history, codebase concerns, and Go module/context guidance. External ecosystem guidance is MEDIUM but not central. |

**Overall confidence:** HIGH

### Gaps to Address

- **Module path and release strategy:** Decide whether stable v2 uses `module github.com/no-src/gofs/v2`, a different module/package path, or pre-release tags before publishing docs/examples.
- **Public package name:** Choose `ftpsync` versus root `gofs` based on import stability and desired discoverability.
- **Exact compatibility adapter scope:** Decide whether legacy FTP URL/YAML parsing ships in core, adapter, examples, or not at all.
- **Cron compatibility:** Decide whether to preserve legacy cron expressions or use only `Monitor.PollInterval`/`time.Ticker` for FTP-source background sync.
- **Concurrency model:** Start serialized because `jlaffaye/ftp.ServerConn` is not concurrency-safe; only add connection pooling/workers if benchmarks and tests justify it.
- **Retained helper dependencies:** After internalization, run `go mod tidy` and verify whether `github.com/no-src/log`, `nsgo`, `fsctl`, and `golang.org/x/time` remain necessary internally.

## Sources

### Primary (HIGH confidence)
- `.planning/research/STACK.md` — dependency strategy, public API shape, cancellation patterns, migration sequence.
- `.planning/research/FEATURES.md` — table-stakes API, differentiators, anti-features, docs/tests requirements.
- `.planning/research/ARCHITECTURE.md` — facade architecture, package disposition, data flow, build order, regression tests.
- `.planning/research/PITFALLS.md` — critical extraction risks, phase-specific warnings, acceptance checklist.
- `.planning/PROJECT.md` and `.planning/MILESTONES.md` — project goals, constraints, shipped FTP v1 behavior.
- `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/STACK.md`, `.planning/codebase/CONVENTIONS.md`, `.planning/codebase/CONCERNS.md` — current codebase structure and risks.

### Secondary (MEDIUM confidence)
- Official Go `context` guidance — cancellation/lifecycle API patterns.
- Official Go module versioning docs — major-version `/v2` import path expectations.
- `github.com/jlaffaye/ftp` package docs — FTP client APIs and connection behavior.
- `github.com/fsnotify/fsnotify` package docs — local filesystem monitoring capabilities and limitations.

---
*Research completed: 2026-04-27*
*Ready for roadmap: yes*
