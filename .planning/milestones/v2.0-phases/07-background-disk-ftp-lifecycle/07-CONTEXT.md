# Phase 7: Background Disk→FTP Lifecycle - Context

**Gathered:** 2026-04-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Implement persistent `disk→FTP` background synchronization behind `FTPSyncService.StartBackground(ctx)`. This phase defines how the long-running lifecycle behaves, how local change events are coalesced, how failures are surfaced, and how callers control stop/wait semantics. It does not add `FTP→disk` background polling, package-pruning work, or release/documentation hardening.

</domain>

<decisions>
## Implementation Decisions

### Startup behavior
- **D-01:** Background `disk→FTP` should perform an initial sync before entering watch mode.
- **D-02:** The library should prioritize remote catch-up correctness over the fastest possible watch startup.

### Error lifecycle
- **D-03:** Single sync failures during background operation should be reported to the caller but should not terminate the background run by default.
- **D-04:** Background mode should behave like a long-running service, not a fail-fast batch job.

### Event handling
- **D-05:** Local filesystem events should use debounce/coalescing rather than immediate per-event upload.
- **D-06:** Avoiding duplicate/redundant uploads during bursty local changes is more important than reacting to every fsnotify event independently.

### Public lifecycle shape
- **D-07:** Public background entry remains `StartBackground(ctx)`; do not introduce a two-step builder/start API in Phase 7.
- **D-08:** Additional lifecycle control should live on the returned handle rather than expanding the public entrypoint shape.

### Handle contract
- **D-09:** The returned background handle should support active stop, wait-for-exit, and final/current error access.
- **D-10:** Phase 7 handle semantics should be strong enough for embedders running the library inside long-lived processes or service wrappers.

### Carried-forward constraints
- **D-11:** Background support remains `disk→FTP` only; `FTP→disk` polling stays out of scope for v2.0.
- **D-12:** Public API remains typed Go options only under package `ftpsync` and service `FTPSyncService`.
- **D-13:** One-shot result/error semantics from Phase 6 stay compact and summary-oriented; Phase 7 should not re-open the one-shot API surface.

### the agent's Discretion
- Exact debounce/coalescing mechanism and default timing window.
- Internal worker/channel structure used to coordinate fsnotify input and sync execution.
- Exact `Handle` concrete type and whether `Wait()` returns error directly or works alongside `Err()`.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone and phase requirements
- `.planning/ROADMAP.md` — Phase 7 goal, requirements, and success criteria for background disk→FTP lifecycle behavior.
- `.planning/REQUIREMENTS.md` — `WATCH-01` through `WATCH-05`, especially lifecycle, deterministic shutdown, and explicit lack of FTP→disk background support.
- `.planning/PROJECT.md` — Locked milestone scope: `ftpsync`, `FTPSyncService`, typed options only, disk→FTP-only background support.

### Prior public API decisions
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-01-SUMMARY.md` — typed options and constructor boundary.
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-02-SUMMARY.md` — `StartBackground(ctx)`, `Handle`, validation, and structured error contracts.
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-03-SUMMARY.md` — logging/progress/event hook surface and no-op behavior.
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-VERIFICATION.md` — verified public API contract for background entrypoint and unsupported capability semantics.

### Existing sync behavior to preserve or extend
- `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-CONTEXT.md` — locked one-shot semantics carried into initial sync and background-triggered sync work.
- `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-01-SUMMARY.md` — compact `Result` and typed partial-failure patterns.
- `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-02-SUMMARY.md` — library-side best-effort local→FTP orchestration pattern.
- `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-VERIFICATION.md` — verifies one-shot library contract already holds before background mode is added.

No external spec or ADR was referenced during this discussion.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `ftpsync/service.go` — already exposes `StartBackground(ctx)` and `Handle` contract stubs from Phase 5.
- `ftpsync/oneshot.go` — already contains the public one-shot execution path Phase 7 should reuse for initial sync and event-triggered uploads.
- `monitor/fsnotify_monitor.go` and `monitor/base_monitor.go` — existing local filesystem watch/coalescing logic worth studying for reuse or simplification.
- `sync/ftp_push_client_sync.go` and `sync/driver_push_client_sync.go` — existing FTP push behavior that background runs should continue to delegate toward rather than re-implementing transfer semantics.

### Established Patterns
- Public API shape is already locked: context-first methods, typed errors, compact result semantics, library-local hooks, no CLI/server runtime imports.
- One-shot execution already uses a library-side orchestration layer over legacy sync methods; Phase 7 should extend that pattern instead of inventing a separate public execution model.
- Existing project already has fsnotify-based monitoring concepts, but Phase 7 may simplify them for a focused library API rather than importing full old runtime behavior wholesale.

### Integration Points
- `FTPSyncService.StartBackground(ctx)` is the only public lifecycle entrypoint for this phase.
- The returned `Handle` should become the caller-owned control surface for stop/wait/error observation.
- Background startup should likely call into the existing `SyncOnce` path first, then keep using the same local→FTP orchestration for coalesced event batches.

</code_context>

<specifics>
## Specific Ideas

- Treat background mode as “initial sync + sustained event-driven push loop”, not as a completely separate sync semantic.
- Keep the API small: no second public starter function, no background FTP→disk mode, and no explosion of background policy knobs in this phase.
- The returned handle should feel suitable for embedding inside another Go service, not just for fire-and-forget scripting.

</specifics>

<deferred>
## Deferred Ideas

- Background `FTP→disk` polling lifecycle — explicitly deferred beyond v2.0 scope.
- Richer public policy switches for startup mode, failure mode, or debounce mode — defer unless Phase 7 planning proves they are necessary.
- Package-pruning of old monitor/runtime surfaces — belongs to Phase 8.

</deferred>

---

*Phase: 07-background-disk-ftp-lifecycle*
*Context gathered: 2026-04-27*
