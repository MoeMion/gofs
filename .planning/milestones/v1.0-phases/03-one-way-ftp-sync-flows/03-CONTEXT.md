# Phase 3: One-Way FTP Sync Flows - Context

**Gathered:** 2026-04-24
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase turns the already-wired FTP endpoint contract and FTP backend driver into real one-way sync behavior that users can run. It covers both `disk→FTP` and `FTP→disk` flows using the existing sync semantics, including long-running FTP-source behavior, delete/rename propagation, and conservative no-op decisions. It does not yet own realistic FTP-server integration coverage or end-user documentation; those remain in Phase 4.

</domain>

<decisions>
## Implementation Decisions

### Flow scope
- **D-01:** Phase 3 must deliver both `disk→FTP` and `FTP→disk` as real user-runnable one-way sync flows, not just constructor wiring.
- **D-02:** Phase 3 should continue to reuse the existing sync architecture and semantics rather than introducing an FTP-specific orchestration model.

### FTP source trigger mode
- **D-03:** `FTP→disk` should work not only for `sync once`, but also in long-running mode.
- **D-04:** If long-running FTP source support requires polling or another non-event-driven strategy, that is in-scope for Phase 3 because the user expects sustained FTP-source operation rather than one-shot-only behavior.

### Delete and rename semantics
- **D-05:** FTP-backed one-way sync should align as closely as possible with the repository's existing one-way sync semantics for delete and rename handling.
- **D-06:** Delete and rename are not optional “nice to have” behaviors for this phase; they are part of preserving the expected sync contract when the backend supports them.

### No-op behavior
- **D-07:** Phase 3 should bias toward stable correctness rather than minimal transfer volume.
- **D-08:** The implementation may tolerate occasional extra transfers when FTP metadata is ambiguous, but it should avoid false no-op decisions that miss real changes.

### Capability failure behavior
- **D-09:** If an FTP server lacks or misbehaves on a capability that Phase 3 depends on, the sync flow should fail explicitly with a clear error.
- **D-10:** Phase 3 should not silently skip key operations or quietly downgrade behavior in order to appear successful.
- **D-11:** Partial-success/soft-warning semantics are not the default for this phase; truthful failure is preferred over ambiguous completion.

### the agent's Discretion
- The exact mechanism for long-running FTP source updates, as long as it preserves the current architecture and supports sustained `FTP→disk` operation.
- The precise polling cadence, batching, or change-detection heuristics, provided they align with the conservative metadata policy already locked in Phase 2.
- Internal factoring between `sync/` and `monitor/` needed to realize these flows without changing the user-visible decisions above.

</decisions>

<specifics>
## Specific Ideas

- Treat Phase 3 as the point where FTP becomes operationally usable, not just architecturally wired.
- Preserve the current one-way behavior contract first; do not optimize for fewer transfers at the expense of missed changes.
- For FTP source long-running mode, correctness and explicit failure matter more than pretending FTP behaves like an event-native source.

</specifics>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project and roadmap scope
- `.planning/PROJECT.md` — Global FTP constraints, including minimal-change architecture and plain-FTP-only scope.
- `.planning/REQUIREMENTS.md` — Phase 3 requirement set `SYNC-01` through `SYNC-04`.
- `.planning/ROADMAP.md` §Phase 3 — Phase goal, success criteria, and dependency on Phase 2.
- `.planning/STATE.md` — Current concerns, prior decisions, and deferred scope boundaries.

### Prior locked decisions
- `.planning/phases/01-ftp-endpoint-contract-routing/01-CONTEXT.md` — FTP endpoint grammar, `ftp_*` fields, default port `21`, passive-mode boolean only.
- `.planning/phases/02-ftp-driver-backend/02-CONTEXT.md` — Conservative metadata policy, bounded reconnects, full driver surface, and explicit capability failure semantics.
- `.planning/phases/02-ftp-driver-backend/02-01-SUMMARY.md` — Real FTP driver behavior now available under `driver/ftp`.
- `.planning/phases/02-ftp-driver-backend/02-02-SUMMARY.md` — FTP sync constructors now wired into the generic driver-backed sync path.

### Existing code and patterns
- `.planning/codebase/ARCHITECTURE.md` — Current monitor/sync layering and lifecycle flow.
- `.planning/codebase/STRUCTURE.md` — Where sync and monitor implementations live.
- `.planning/codebase/CONVENTIONS.md` — Go code and testing conventions to preserve.
- `sync/driver_push_client_sync.go` — Generic local-to-remote one-way sync behavior for create/write/delete/rename and no-op checks.
- `sync/driver_pull_client_sync.go` — Generic remote-to-local one-way sync behavior for traversal, file writes, and metadata-based comparisons.
- `sync/ftp_push_client_sync.go` — FTP destination constructor already wired to the generic push path.
- `sync/ftp_pull_client_sync.go` — FTP source constructor already wired to the generic pull path.
- `monitor/ftp_pull_client_monitor.go` — Current FTP source monitor placeholder that Phase 3 likely needs to replace for sustained FTP-source operation.
- `monitor/sftp_pull_client_monitor.go` — Closest existing pull-monitor wrapper pattern.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `sync/driver_push_client_sync.go`: Already contains most of the behavior for `disk→FTP`, including create/write/delete/rename flow and logical-delete integration.
- `sync/driver_pull_client_sync.go`: Already contains most of the behavior for `FTP→disk`, including traversal, pull writes, and metadata-based no-op checks.
- `sync/ftp_push_client_sync.go` and `sync/ftp_pull_client_sync.go`: Already connect FTP endpoints to the generic driver sync flows.
- `driver/ftp/ftp.go`: Phase 2 already provides the backend operations and conservative metadata/reconnect behavior Phase 3 should build on.
- `monitor/sftp_pull_client_monitor.go`: Simple wrapper pattern that can guide FTP-source long-running monitor integration.

### Established Patterns
- One-way sync semantics are owned by generic sync helpers; protocol-specific files should stay thin where possible.
- Long-running source behavior is mediated through `monitor/` rather than embedding watch loops inside sync constructors.
- Existing sync code already supports delete/rename propagation and logical delete behavior; Phase 3 should preserve those semantics unless the backend truthfully cannot support them.

### Integration Points
- `monitor/ftp_pull_client_monitor.go` is currently a deferred placeholder and is the clearest Phase 3 integration gap for long-running `FTP→disk` behavior.
- `sync/driver_push_client_sync.go` and `sync/driver_pull_client_sync.go` are the primary behavioral surfaces whose existing semantics Phase 3 must preserve in real FTP flows.
- Any Phase 3 implementation must respect Phase 2's explicit capability-failure rule rather than hiding backend limitations in the sync layer.

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 03-one-way-ftp-sync-flows*
*Context gathered: 2026-04-24*
