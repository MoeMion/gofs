# Phase 6: One-Shot Disk↔FTP Sync Through Library API - Context

**Gathered:** 2026-04-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Implement the actual one-shot `disk→FTP` and `FTP→disk` execution behavior behind the already-defined `ftpsync.FTPSyncService` API. This phase delivers real transfer behavior, result semantics, and path safety for `SyncOnce`, while preserving the typed-options API boundary locked in Phase 5. Background lifecycle behavior, package pruning, and release/docs hardening remain separate phases.

</domain>

<decisions>
## Implementation Decisions

### Result contract
- **D-01:** `SyncOnce` should return a summary-style `Result`, not a full per-file report.
- **D-02:** The summary should stay compact and library-friendly rather than exposing detailed file-by-file execution records in Phase 6.

### Failure behavior
- **D-03:** One-shot sync should use best-effort execution rather than fail-fast.
- **D-04:** When some files succeed and some fail, `SyncOnce` must return both `Result` and a non-nil `error`.
- **D-05:** The returned `error` should express partial failure without losing the summary result that callers need for logging, orchestration, or retry decisions.

### Sync policy surface
- **D-06:** Phase 6 should not expand the public API with delete, overwrite, or compare-mode policy options.
- **D-07:** One-shot behavior should follow the existing FTP v1 sync semantics already proven in the legacy implementation and tests.

### Local path semantics
- **D-08:** For `FTP→disk`, the configured target local root may be auto-created if it does not exist.
- **D-09:** The local root must still be explicitly provided by the caller; the library must never fall back to the process working directory.
- **D-10:** CWD safety remains a non-negotiable regression boundary for this phase.

### Carried-forward constraints
- **D-11:** Public package remains `ftpsync` and public entrypoint remains `FTPSyncService`.
- **D-12:** Public configuration remains typed Go options only; do not reintroduce YAML, CLI flag, or URL parser requirements into the Phase 6 API.

### the agent's Discretion
- Exact summary fields inside `Result`, as long as they remain summary-oriented and sufficient for library callers.
- Internal adapter shape used to bridge `ftpsync` to existing `sync` and `driver/ftp` behavior.
- Exact partial-failure error type/wording, as long as it preserves the existing `ErrorKind` approach and supports `Result + error` semantics.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Milestone and phase requirements
- `.planning/ROADMAP.md` — Phase 6 goal, requirements, and success criteria for one-shot library sync.
- `.planning/REQUIREMENTS.md` — `ONCE-01` through `ONCE-05`, especially one-shot behavior preservation, cwd safety, and result summary expectations.
- `.planning/PROJECT.md` — Locked milestone direction: `ftpsync`, `FTPSyncService`, typed options only, CLI removal in progress.

### Prior phase decisions
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-01-SUMMARY.md` — Public options, constructor, and dependency-boundary decisions.
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-02-SUMMARY.md` — Validation, `context.Context`, and structured error contracts.
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-03-SUMMARY.md` — Hook contracts and no-op default behavior.
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-VERIFICATION.md` — Verified scope boundary showing `SyncOnce` contract exists but execution is deferred to Phase 6.

### Existing FTP behavior to preserve
- `.planning/milestones/v1.0-phases/02-ftp-driver-backend/02-02-SUMMARY.md` — Existing FTP push/pull constructor routing into real driver-backed sync paths.
- `.planning/milestones/v1.0-phases/03-one-way-ftp-sync-flows/03-02-SUMMARY.md` — Existing one-way FTP flow semantics, delete/rename behavior, and conservative no-op expectations.
- `.planning/debug/sync-files-copied-to-cwd.md` — Root cause and fix for cwd leakage; must remain protected in library API.

If additional docs are needed, they should be added by research/planning. No external product spec was referenced during this discussion.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `sync/ftp_push_client_sync.go` — Existing `disk→FTP` one-shot implementation entrypoint that already wires through the real FTP driver-backed sync path.
- `sync/ftp_pull_client_sync.go` — Existing `FTP→disk` one-shot implementation entrypoint with FTP-specific metadata handling and remote-path normalization.
- `driver/ftp/ftp.go` and `driver/ftp/encoding.go` — Existing FTP transport, reconnect, passive mode, timeout, path encoding, and metadata behavior to preserve.
- `ftpsync/service.go` — Existing public service shell with validated options, structured errors, and `SyncOnce` placeholder ready to be connected to real execution.

### Established Patterns
- Existing sync behavior is factory-led and driver-backed rather than protocol-specific business logic duplicated in multiple layers.
- Public Phase 5 API already established `Result + error` style method signatures and typed `ErrorKind` classification; Phase 6 should fill execution behind that boundary instead of changing the public shape.
- Existing Go code favors small constructors, explicit validation, immediate error checks, and table-driven tests.

### Integration Points
- `ftpsync/service.go` will become the public orchestration bridge into existing sync machinery.
- Existing `sync/ftp_push_client_sync.go` and `sync/ftp_pull_client_sync.go` are the most direct bridges for preserving v1 one-shot semantics.
- CWD safety and local-root handling must be verified through the same public `ftpsync.SyncOnce` boundary, not only by lower-level sync tests.

</code_context>

<specifics>
## Specific Ideas

- Keep the Phase 6 API intentionally narrow: real one-shot execution first, no policy explosion.
- Preserve current FTP v1 one-shot semantics instead of inventing a new library-only behavior model.
- Treat partial success as a first-class library outcome by returning summary data together with a non-nil error.

</specifics>

<deferred>
## Deferred Ideas

- Exposing delete/overwrite/compare behavior as explicit public one-shot policy options — defer unless later phases need a broader surface.
- Returning full per-file execution reports from `SyncOnce` — defer beyond the summary-first Phase 6 contract.
- Background `FTP→disk` polling behavior — explicitly deferred to future requirements, not part of v2.0.

</deferred>

---

*Phase: 06-one-shot-disk-ftp-sync-through-library-api*
*Context gathered: 2026-04-27*
