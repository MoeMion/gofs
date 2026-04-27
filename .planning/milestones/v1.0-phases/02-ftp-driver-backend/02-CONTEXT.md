# Phase 2: FTP Driver Backend - Context

**Gathered:** 2026-04-23
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase implements the real FTP backend behind the existing `driver.Driver` contract so FTP endpoints can perform connect, traversal, read, write, create, delete, rename, and metadata-based comparison support. It does not yet deliver end-user `disk→FTP` / `FTP→disk` workflow verification as a complete feature; Phase 3 owns sync-flow behavior on top of this backend, and Phase 4 owns broader test/documentation completion.

</domain>

<decisions>
## Implementation Decisions

### Driver integration shape
- **D-01:** Phase 2 must integrate FTP by implementing the existing `driver.Driver` boundary rather than inventing a protocol-specific sync engine.
- **D-02:** The FTP backend should plug into the generic `driverPushClientSync` and `driverPullClientSync` patterns already used for remote backends where possible.

### File state comparison
- **D-03:** FTP file-change detection should be conservative.
- **D-04:** The implementation should primarily use size and modification time when available, but when FTP metadata is unreliable or incomplete it should prefer extra transfers over risking missed changes.
- **D-05:** Phase 2 should not escalate into a checksum-heavy design by default just to compensate for FTP metadata limitations.

### Connection and reconnect behavior
- **D-06:** FTP should follow a conservative automatic reconnect strategy similar in spirit to existing SFTP and MinIO drivers.
- **D-07:** Retries and reconnects should stay bounded and non-aggressive; do not introduce complex self-healing state machines or high-concurrency recovery logic in this phase.
- **D-08:** When recovery fails, return explicit errors instead of hiding backend instability.

### Directory, delete, and rename semantics
- **D-09:** Phase 2 should implement the full `driver.Driver` capability surface needed by the current sync layer: traversal, mkdir, file create/read/write, recursive delete, and rename.
- **D-10:** If a specific FTP server cannot support an operation robustly, the driver should return a clear operational error rather than silently degrading behavior.
- **D-11:** The goal is interface compatibility with the existing driver-backed sync path, not a reduced FTP-only subset.

### Timestamp precision policy
- **D-12:** FTP timestamp precision should be treated as coarse and potentially inconsistent across servers.
- **D-13:** Comparison logic should explicitly accept a conservative precision policy: when timestamp fidelity is questionable, bias toward retransferring rather than skipping.
- **D-14:** This precision limitation should become part of the expected implementation and future testing/documentation story rather than something hidden behind optimistic assumptions.

### the agent's Discretion
- Which Go FTP client library best fits the repo's minimal-change requirements, as long as it supports the required `driver.Driver` operations and bounded reconnect handling.
- The exact internal shape of FTP-specific metadata wrappers and helper methods, provided they serve the conservative comparison and full-driver-capability decisions above.
- The exact retry thresholds and reconnect triggers, provided they remain conservative and explicit.

</decisions>

<specifics>
## Specific Ideas

- Reuse the current remote-driver architecture rather than building an FTP-specific execution path.
- Prefer a “safe and slightly chatty” FTP backend over an “optimized but risky” backend.
- Treat FTP server variance as expected reality: the driver should fail clearly when a server cannot support a required operation, not pretend the operation succeeded.

</specifics>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project and phase scope
- `.planning/PROJECT.md` — Global constraints for FTP support, including minimal-change architecture and plain-FTP-only scope.
- `.planning/REQUIREMENTS.md` — Phase 2 requirement set `FTPD-01` through `FTPD-09`.
- `.planning/ROADMAP.md` §Phase 2 — Phase goal, success criteria, and dependency on Phase 1.
- `.planning/STATE.md` — Active concerns already recorded for timestamp fidelity and conservative retry behavior.

### Prior locked decisions
- `.planning/phases/01-ftp-endpoint-contract-routing/01-CONTEXT.md` — Phase 1 decisions that Phase 2 must honor: query-style FTP endpoint grammar, dedicated `ftp_*` fields, passive-mode boolean only, default port `21`, optional timeout.
- `.planning/phases/01-ftp-endpoint-contract-routing/01-01-SUMMARY.md` — Actual VFS contract now implemented for FTP endpoints.
- `.planning/phases/01-ftp-endpoint-contract-routing/01-02-SUMMARY.md` — Actual sync/monitor routing now implemented and intentionally deferred to Phase 2 backend behavior.

### Existing code and patterns
- `.planning/codebase/ARCHITECTURE.md` — How drivers, sync, and monitor layers connect.
- `.planning/codebase/STRUCTURE.md` — Where backend drivers and sync helpers live.
- `.planning/codebase/CONVENTIONS.md` — Naming, config, and Go code conventions to preserve.
- `driver/driver.go` — The concrete driver contract Phase 2 must satisfy.
- `driver/sftp/sftp.go` — Main analog for connection lifecycle, reconnect shape, and remote filesystem semantics.
- `driver/minio/minio.go` — Analog for reconnect handling and metadata-backed remote driver behavior.
- `sync/driver_push_client_sync.go` — Generic push path that Phase 2 FTP driver should support.
- `sync/driver_pull_client_sync.go` — Generic pull path that Phase 2 FTP driver should support.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `driver/driver.go`: Existing interface already defines the operations FTP must implement, including traversal, stat/lstat, read/write, rename, delete, and time updates.
- `driver/sftp/sftp.go`: Strong analog for a stateful remote filesystem driver with bounded reconnect behavior, recursive delete, and file abstractions.
- `driver/minio/minio.go`: Analog for metadata-backed remote driver tradeoffs, especially around reconnect and partial capability support.
- `sync/driver_push_client_sync.go`: Generic local-to-remote sync path already handles write, mkdir, rename/delete orchestration, and post-write time propagation through `driver.Chtimes`.
- `sync/driver_pull_client_sync.go`: Generic remote-to-local sync path already handles read, stat-driven comparison, and traversal through `driver.WalkDir`.

### Established Patterns
- Remote backend support is expected to come from implementing `driver.Driver`, then reusing generic sync wrappers rather than bypassing them.
- Existing drivers keep reconnect logic inside the driver layer instead of spreading it through sync code.
- The current sync layer already assumes metadata-based no-op detection, which means FTP precision and reliability choices belong in the driver behavior and test expectations.

### Integration Points
- New FTP backend should live under `driver/ftp/` and implement `driver.Driver`.
- Existing FTP sync placeholders in `sync/ftp_push_client_sync.go` and `sync/ftp_pull_client_sync.go` are the handoff points where Phase 2 will swap deferred errors for real driver-backed behavior.
- `monitor/ftp_pull_client_monitor.go` is the handoff point for real FTP source monitoring behavior to become possible in later phases.
- Comparison and mutation behavior must align with what `sync/driver_push_client_sync.go` and `sync/driver_pull_client_sync.go` already expect from remote drivers.

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 02-ftp-driver-backend*
*Context gathered: 2026-04-23*
