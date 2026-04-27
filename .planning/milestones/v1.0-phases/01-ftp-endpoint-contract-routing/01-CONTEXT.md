# Phase 1: FTP Endpoint Contract & Routing - Context

**Gathered:** 2026-04-23
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase defines how FTP endpoints are expressed in configuration and how gofs recognizes and routes them through the existing VFS, sync, and monitor factory paths. It does not implement the FTP backend behavior itself; protocol operations such as list, upload, download, or reconnect belong to Phase 2.

</domain>

<decisions>
## Implementation Decisions

### FTP endpoint contract
- **D-01:** FTP endpoints should use the same query-parameter-oriented VFS contract style already used by `sftp://` and `minio://` endpoints.
- **D-02:** Phase 1 should preserve the current `core.VFS` parsing model rather than introducing a second URL grammar centered on `ftp://user:pass@host/path` semantics.

### Credential mapping
- **D-03:** FTP credentials should use FTP-specific parameter names rather than reusing SSH-oriented parameter names.
- **D-04:** Phase 1 should keep FTP authentication fields semantically explicit so downstream driver work can consume them without SSH-specific ambiguity.

### Passive mode configuration
- **D-05:** Passive mode should be exposed in Phase 1 as a single boolean configuration switch.
- **D-06:** Phase 1 should not introduce additional FTP compatibility toggles such as EPSV-specific controls; those remain future-scope unless Phase 2 research proves they are required.

### Defaulting behavior
- **D-07:** FTP should follow the existing remote-endpoint convention of having a default port, with `21` used when the endpoint omits an explicit port.
- **D-08:** FTP timeout should be an endpoint-level optional parameter rather than a required setting.

### the agent's Discretion
- Exact parameter names for FTP-specific fields, as long as they remain FTP-specific, internally consistent, and documented.
- Whether timeout is represented as a general remote timeout field or an FTP-named timeout field, provided the exposed contract is endpoint-level and optional.
- Internal factoring between `core.VFS`, config parsing, and factory wiring, provided the external behavior matches the decisions above.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project scope and constraints
- `.planning/PROJECT.md` — Overall project goal, locked constraints, and explicit v1 exclusions for FTP support.
- `.planning/REQUIREMENTS.md` — Phase-mapped FTP endpoint requirements `FTP-01` through `FTP-04` and the out-of-scope list.
- `.planning/ROADMAP.md` §Phase 1 — Phase goal, requirement mapping, and success criteria for FTP endpoint contract and routing.
- `.planning/STATE.md` — Current phase focus and cross-phase concerns already noted for later FTP work.

### Existing architecture and codebase patterns
- `.planning/codebase/ARCHITECTURE.md` — Current VFS/driver/sync/monitor layering and where protocol extensions fit.
- `.planning/codebase/STRUCTURE.md` — File locations and extension points for `core/`, `sync/`, `monitor/`, and `driver/`.
- `.planning/codebase/CONVENTIONS.md` — Naming, config, and code-organization patterns to preserve when adding FTP support.

### Existing implementation anchors
- `core/vfs_type.go` — Existing `FTP` enum value confirms the repo already reserves FTP as a first-class VFS type.
- `core/vfs.go` — Current remote endpoint parsing contract and default-port handling for `rs://`, `sftp://`, and `minio://`.
- `conf/config.go` — Central configuration model through which `core.VFS` enters runtime config.
- `sync/sync.go` — Sync factory routing that Phase 1 must extend so FTP endpoints are no longer unsupported.
- `monitor/monitor.go` — Source-side monitor routing that Phase 1 must extend for future FTP source support.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `core/vfs.go`: Existing remote endpoint parsing flow, query parameter extraction, and protocol-specific default port handling can be extended for FTP.
- `core/vfs_type.go`: `FTP` already exists in the enum, so Phase 1 does not need a new top-level type classification.
- `sync/sync.go`: Current source/destination dispatch pattern already branches by `core.VFSType`; FTP can slot into this same factory style.
- `monitor/monitor.go`: Existing pull-monitor dispatch pattern already branches by source VFS type and can accept a future FTP branch.
- `conf/config.go`: `Source` and `Dest` already use `core.VFS`, so no new top-level config object is required to introduce FTP endpoints.

### Established Patterns
- Remote backends are recognized by scheme in `core.VFS`, then routed through factory methods in `sync/` and `monitor/`.
- Existing protocols use protocol-specific defaults rather than forcing every endpoint to specify all connection details explicitly.
- Brownfield changes are expected to preserve current abstractions instead of introducing a separate orchestration path per protocol.

### Integration Points
- `core/vfs.go`: add `ftp://` recognition, FTP-specific query parsing, and FTP default-port behavior.
- `sync/sync.go`: add FTP-aware source/destination routing in line with existing SFTP and MinIO branches.
- `monitor/monitor.go`: add FTP source routing so the phase goal of accepted `ftp://` endpoints is fully wired for later phases.
- Potential follow-on callers that depend on `core.VFS` fields: sync constructors, future FTP driver constructor, and any docs/examples that serialize endpoint strings.

</code_context>

<specifics>
## Specific Ideas

- Keep FTP endpoint syntax aligned with current gofs remote endpoint style rather than introducing a more standard-but-different FTP URL grammar.
- Keep Phase 1 narrow: expressibility and routing now, backend protocol behavior later.
- Preserve minimal-change symmetry with existing remote backends by using default port behavior and endpoint-local optional timeout.

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 01-ftp-endpoint-contract-routing*
*Context gathered: 2026-04-23*
