# Domain Pitfalls: v2.0 FTP Sync Library Extraction

**Domain:** Destructive extraction of a Go library from an existing CLI/server sync application  
**Project:** `gofs` v2.0 FTP Sync Library  
**Researched:** 2026-04-27  
**Overall confidence:** HIGH for project-specific risks; MEDIUM for external ecosystem guidance

## Critical Pitfalls

### Pitfall 1: Exporting the old runtime instead of a library API
**What goes wrong:** `FTPSyncService` becomes a thin wrapper around `cmd.RunWithConfig`, global flags, daemon/server bootstrap, or process-level signal handling.

**Why it happens:** The current application already has a working config-driven orchestration path, but `.planning/codebase/CONCERNS.md` identifies `cmd/gofs.go` as a large, tightly coupled startup function that mixes config mutation, auth setup, server startup, monitor startup, retry, logging, daemon behavior, and signal handling.

**Consequences:**
- Library calls inherit CLI/server side effects.
- Consumers cannot safely run multiple sync services in one process.
- Tests require process-level setup instead of isolated service instances.
- Deleted server/CLI code becomes hard to remove because the public API still depends on it.

**Prevention:**
- Design `FTPSyncService` as the root API, not `cmd` as the root API.
- Keep process concerns out of the package: no `os.Exit`, no global flag parsing, no daemonization, no signal registration, no implicit servers.
- Make one-shot and background sync methods accept explicit structs and `context.Context`.
- Treat old CLI config as an input-compatibility source only, not the internal execution model.

**Detection:** Public package imports `cmd`, `flag`, `server`, `api`, or task packages; service construction mutates global state; unit tests must change process flags or cwd.

### Pitfall 2: Breaking path semantics and reintroducing cwd leaks
**What goes wrong:** Files are written relative to the process working directory when a caller intended a configured local root or remote FTP root.

**Why it happens:** The resolved FTP v1 bug showed that an omitted FTP `path` query normalized to `.` and optional local mirroring copied files into cwd. Library extraction increases this risk because CLI-era defaults may be moved, simplified, or renamed.

**Consequences:**
- Sync writes outside caller-approved directories.
- Tests pass from repository root but fail in embedded/library use.
- Background sync creates unexpected local mirrors.
- Security issue: caller-provided FTP operations affect process cwd.

**Prevention:**
- Preserve the v1 distinction between an explicit local path and an implicit/omitted path.
- Require local roots to be explicit for disk writes; do not infer cwd as a destination.
- Normalize paths at service construction, store absolute local roots internally, and keep remote FTP paths separate from local filesystem paths.
- Add cwd-sentinel tests for every public operation: create/write/remove/rename, one-shot sync, and background sync.

**Detection:** Any use of `.` as a default write root; tests do not `chdir` to a temp directory; local mirror behavior depends only on `!LocalSyncDisabled()` rather than explicit path presence.

### Pitfall 3: Losing cancellation and lifecycle ownership
**What goes wrong:** Background sync starts goroutines that the caller cannot stop, wait for, or observe.

**Why it happens:** The old app owns the process lifecycle. A library must delegate lifecycle to callers. Go's `context` guidance is to pass cancellation signals and deadlines across API boundaries so goroutines exit promptly when work is canceled.

**Consequences:**
- Goroutine leaks after `Stop`, test completion, or caller timeout.
- FTP connections and file watchers remain open.
- Callers cannot distinguish clean cancellation from sync failure.
- Race tests become flaky because background workers survive between tests.

**Prevention:**
- API shape should make lifecycle explicit, for example `SyncOnce(ctx, opts)` and `Start(ctx, opts) (*Run, error)` with `Run.Wait()` / `Run.Stop()`.
- Every goroutine must select on `ctx.Done()` or a service-owned stop channel.
- Close FTP connections, local watchers, timers, tickers, and retry loops on cancellation.
- Return `context.Canceled` / `context.DeadlineExceeded` distinctly from protocol errors.

**Detection:** Background APIs return only `error`; no wait handle exists; `context.Background()` appears inside service internals; tests need sleeps instead of deterministic `Wait()`.

### Pitfall 4: Preserving monitor timing bugs while deleting their tests
**What goes wrong:** The refactor keeps stateful monitor/coalescing code but deletes integration tests or race coverage that previously exercised it.

**Why it happens:** Non-FTP runtime surfaces are being removed aggressively, and tests often live near deleted packages. `.planning/codebase/CONCERNS.md` flags `monitor/base_monitor.go` as timing-sensitive, worker-based, and fragile.

**Consequences:**
- Background sync silently drops events.
- Shutdown hangs under burst traffic.
- CPU churn persists from polling loops.
- Library consumers see nondeterministic behavior that the old CLI masked.

**Prevention:**
- Before deletion, classify tests as protocol-specific, engine-specific, monitor-specific, or server-only.
- Move engine/monitor tests into the retained FTP library package before removing server/CLI packages.
- Add leak tests around start/stop loops and run `go test -race` on retained packages.
- Prefer channel/ticker ownership with explicit shutdown over sleep-based polling.

**Detection:** Large test deletions without replacement; background sync tests only cover happy-path startup; race CI no longer exercises monitor code.

### Pitfall 5: Treating CLI configuration compatibility as API stability
**What goes wrong:** The public Go API mirrors every old flag/config field, including obsolete server, auth, TLS, SFTP, MinIO, task, daemon, and web options.

**Why it happens:** The project requires `FTPSyncService` parameters equivalent to current CLI configuration semantics, but v2.0 also explicitly removes non-FTP runtime surfaces. Equivalence means FTP behavior compatibility, not preserving the entire old config object.

**Consequences:**
- Public API is bloated on day one.
- Removed runtime concepts leak into docs and type names.
- Future cleanup becomes a breaking change to library consumers.
- Users cannot tell which fields actually affect FTP sync.

**Prevention:**
- Define a new narrow config surface: FTP endpoint config, local path config, sync behavior, monitor behavior, retry/timeouts, logging hooks.
- Provide a compatibility adapter from old config/URLs to the new structs if needed, but keep it separate from the core API.
- Mark unsupported legacy fields as rejected with clear errors, not silently ignored.
- Add golden tests for legacy FTP URL/config examples from v1 docs.

**Detection:** Public types expose `EnableFileServer`, gRPC, task, SFTP, MinIO, session, token, or daemon fields; docs list old CLI flags as library options.

### Pitfall 6: Incorrect Go module/versioning plan for a destructive v2
**What goes wrong:** The repository ships a breaking library as `v2.0.0` without aligning the module path, import path, docs, and release notes.

**Why it happens:** Go modules treat major versions above v1 as separate module paths with `/v2` suffixes, and major releases signal backward-incompatible public API changes. This milestone is intentionally breaking.

**Consequences:**
- Consumers cannot `go get` the intended version cleanly.
- Package docs and examples import the wrong path.
- Old CLI users upgrade unexpectedly into a removed runtime.
- Downstream builds break in confusing ways.

**Prevention:**
- Decide explicitly whether this is a true Go module `v2` (`module github.com/no-src/gofs/v2`) or a pre-v1/library reset under a different module/package path.
- Update README, package docs, examples, CI, release workflow, and tags together.
- Publish migration notes that state CLI/server/SFTP/MinIO removal clearly.
- If import path churn is not desired yet, ship pre-release tags first and avoid claiming stable v2.

**Detection:** `go.mod` still uses `github.com/no-src/gofs` while release notes call it v2; examples omit `/v2`; CI only validates local module builds.

### Pitfall 7: Deleting transitive dependencies before proving compile boundaries
**What goes wrong:** Removing Gin/gRPC/SFTP/MinIO/task/server packages breaks retained sync code through hidden imports, tests, or shared utility types.

**Why it happens:** The old architecture has clean conceptual layers, but practical imports cross layers: sync, monitor, server, api, driver, report, retry, logger, and config are composed through option structs and factories.

**Consequences:**
- Refactor turns into broad package surgery.
- Non-FTP dependencies stay in `go.mod` because retained packages still import them.
- Public API accidentally depends on deleted packages.
- Build tags or stubs hide broken design instead of simplifying it.

**Prevention:**
- Work from retained package graph inward: `go list` target packages, then delete only what is outside the graph.
- Split shared utilities from runtime-specific packages before deleting surfaces.
- Keep `go mod tidy` as a release gate; verify removed dependencies disappear from `go.mod` and `go.sum`.
- Avoid compatibility stubs for removed runtime features unless they are clearly deprecated adapters.

**Detection:** Library package imports `server/`, `api/`, `cmd/`, SFTP, or MinIO; `go mod tidy` keeps old runtime dependencies; compile fixes add empty placeholder packages.

## Moderate Pitfalls

### Pitfall 8: Hiding errors behind CLI-style logging
**What goes wrong:** Library methods log failures but return nil, generic errors, or process-oriented result handles.

**Prevention:** Return structured errors with operation, local path, remote path, and retry/cancellation context. Allow caller-provided logging hooks, but never require log scraping for correctness.

### Pitfall 9: Making retries unsafe for library consumers
**What goes wrong:** Existing retry behavior duplicates FTP side effects after partial success. This is already identified as a codebase concern.

**Prevention:** Restrict retries to idempotent operations or make mutation retries prove remote state before replay. Document retry defaults in the public API.

### Pitfall 10: Silent behavior changes in delete/rename/no-op sync
**What goes wrong:** One-shot sync still copies files, but delete propagation, rename behavior, and second-run no-op behavior drift during simplification.

**Prevention:** Preserve v1 FTP regression tests for disk→FTP and FTP→disk flows, including nested paths, delete/rename-relevant behavior, and no-op checks.

### Pitfall 11: Documentation describes the old product, not the new library
**What goes wrong:** README and examples continue to focus on CLI flags, web/file server, gRPC, SFTP, or MinIO, while the package API is under-documented.

**Prevention:** Rewrite docs around import path, `FTPSyncService`, one-shot example, background example, cancellation, path semantics, and limitations. Keep a removed-features migration section.

### Pitfall 12: Security messaging gets worse after removing servers
**What goes wrong:** Because HTTP/gRPC auth is deleted, docs understate remaining FTP credential and plaintext transport risks.

**Prevention:** State plainly that v2.0 supports plain FTP only, credentials travel according to FTP behavior, FTPS is out of scope, and callers must choose trusted networks or separate security controls.

## Minor Pitfalls

### Pitfall 13: Package names and examples remain CLI-shaped
**What goes wrong:** Consumers import packages with names like `cmd`, `conf`, or `sync` and cannot discover the intended library surface.

**Prevention:** Provide a small, obvious public package and keep old/internal packages unexported or internal where possible.

### Pitfall 14: Test fixtures become too heavyweight for a library
**What goes wrong:** Library tests require the old full application runtime or server setup.

**Prevention:** Keep unit tests pure and isolate real FTP server tests behind existing integration tags/fixtures.

### Pitfall 15: Release artifacts still imply CLI support
**What goes wrong:** Docker images, binary release workflows, badges, or install instructions promise a runtime that v2 removed.

**Prevention:** Update or remove binary/container release workflows in the same milestone as code deletion.

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|----------------|------------|
| Public API design | Wrapping `cmd.RunWithConfig` instead of extracting a service | Define `FTPSyncService` first; make old config adaptation optional |
| Path model | Omitted paths become cwd writes | Require explicit local roots; preserve `HasLocalPath`-style checks; add cwd sentinel tests |
| Background sync | Goroutines survive caller cancellation | Context-first API, wait handles, deterministic stop tests, race tests |
| Code deletion | Removing tests with removed packages | Move retained FTP/sync/monitor tests before deleting surfaces |
| Config compatibility | Public structs mirror obsolete CLI flags | Narrow FTP config; reject unsupported legacy fields explicitly |
| Module release | Breaking v2 without `/v2` module path or migration docs | Align `go.mod`, tags, import examples, README, release notes |
| Dependency pruning | Old runtime dependencies remain hidden in retained packages | Verify package graph and run `go mod tidy` as a gate |
| Docs | Users follow old CLI/server instructions | Replace docs with library examples and removed-feature notes |

## Minimum Acceptance Checklist

- [ ] Public API does not import or depend on CLI/server/task packages.
- [ ] One-shot sync accepts `context.Context` and returns operation errors directly.
- [ ] Background sync exposes deterministic stop/wait behavior.
- [ ] No retained code writes to cwd unless caller explicitly configured cwd as the local root.
- [ ] FTP source and destination semantics match v1 regression coverage.
- [ ] Race-enabled tests cover retained background sync paths.
- [ ] Legacy FTP config examples either work through an adapter or fail with clear unsupported-field errors.
- [ ] README, examples, package docs, `go.mod`, tags, and release workflows agree on the new library/import path.
- [ ] `go mod tidy` removes deleted runtime dependencies.

## Sources

- `.planning/PROJECT.md` — v2.0 goal, active requirements, removed surfaces, constraints.
- `.planning/MILESTONES.md` — v1.0 FTP behavior and regression coverage to preserve.
- `.planning/codebase/CONCERNS.md` — fragile startup, retry, monitor concurrency, dependency and test risks.
- `.planning/debug/sync-files-copied-to-cwd.md` — cwd leak root cause and regression requirement.
- Go documentation: module version numbering and major-version `/v2` path expectations — https://go.dev/doc/modules/version-numbers
- Go blog: context cancellation/lifecycle guidance — https://go.dev/blog/context
