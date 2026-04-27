# Requirements: gofs

**Defined:** 2026-04-27
**Milestone:** v2.0 FTP Sync Library
**Core Value:** Provide a focused Go package for reliable FTP file synchronization without requiring the existing CLI/server runtime.

## v2.0 Requirements

Requirements for converting the existing FTP synchronization capability into a focused Go library. Each requirement maps to exactly one roadmap phase.

### Public Library API

- [ ] **API-01**: Developer can import a `ftpsync` package and construct an `FTPSyncService` without invoking CLI, process, server, or daemon code
- [ ] **API-02**: Developer can configure `FTPSyncService` through typed Go options covering local path, FTP host, port, username, password, remote path, passive mode, timeout, path encoding, retry behavior, and ignore rules
- [ ] **API-03**: `FTPSyncService` validates supported directions explicitly and rejects unsupported local↔local, FTP↔FTP, non-FTP, missing-path, and ambiguous endpoint combinations
- [ ] **API-04**: Public sync methods accept `context.Context` and return structured errors for validation, cancellation, connection/authentication, transfer, and unsupported-capability failures
- [ ] **API-05**: Public API exposes optional no-op-by-default hooks for logging, progress, and sync event reporting without requiring global loggers or web reports

### One-Shot Sync

- [ ] **ONCE-01**: Developer can run one-shot local disk→FTP synchronization through `FTPSyncService`
- [ ] **ONCE-02**: Developer can run one-shot FTP→local disk synchronization through `FTPSyncService`
- [ ] **ONCE-03**: One-shot sync preserves existing FTP v1 behavior for create, update, delete, rename, nested paths, passive mode, timeout, and path encoding
- [ ] **ONCE-04**: One-shot sync never writes files to the process working directory unless that directory is explicitly configured as the local endpoint
- [ ] **ONCE-05**: One-shot sync returns a useful result summary suitable for library callers instead of relying on CLI output

### Background Disk to FTP Sync

- [ ] **WATCH-01**: Developer can start persistent local disk→FTP synchronization through `FTPSyncService` without invoking the CLI monitor runtime
- [ ] **WATCH-02**: Background disk→FTP sync detects local file create, update, delete, and rename-relevant changes and applies them to the FTP destination
- [ ] **WATCH-03**: Background sync returns a lifecycle handle that supports readiness, error observation, wait, and deterministic shutdown
- [ ] **WATCH-04**: Background sync stops all watchers, timers, retry sleeps, goroutines, and FTP connections when its context is cancelled or the handle is stopped
- [ ] **WATCH-05**: Background sync does not expose FTP→disk polling or bidirectional conflict resolution in v2.0

### Package Reduction

- [ ] **PRUNE-01**: Library build no longer depends on CLI entrypoints, flag parsing, daemon/process management, HTTP file server, gRPC APIs, task runtime, auth/session, SFTP, MinIO, or Docker release surfaces
- [ ] **PRUNE-02**: FTP driver, path encoding, retry, ignore/filtering, rate limiting, and core disk/FTP sync internals remain available behind the public library API
- [ ] **PRUNE-03**: Internal packages do not leak old `conf.Config`, `core.VFS`, CLI URL parsing, or server/task types into the public `ftpsync` API
- [ ] **PRUNE-04**: `go.mod` is tidied so removed runtime surfaces no longer keep unnecessary dependencies such as Gin, gRPC, SFTP, MinIO, Redis/cache, QUIC, OAuth, and protobuf packages

### Verification and Documentation

- [ ] **VERIFY-01**: Automated tests cover public `FTPSyncService` construction, validation, one-shot push, one-shot pull, and background disk→FTP lifecycle behavior
- [ ] **VERIFY-02**: Real FTP integration tests verify library-based local→FTP and FTP→local one-shot flows without the CLI
- [ ] **VERIFY-03**: Regression tests cover cwd safety, FTP path encoding, passive-mode defaults, cancellation, and background goroutine shutdown
- [ ] **DOC-01**: README and package examples show how to use `ftpsync.FTPSyncService` with typed options for one-shot push, one-shot pull, and background disk→FTP sync
- [ ] **DOC-02**: Documentation clearly states removed v2.0 surfaces and limitations: no CLI runtime, no FTPS, no FTP server, no FTP↔FTP sync, no FTP→disk background polling, and no bidirectional conflict resolution
- [ ] **DOC-03**: Release notes or migration documentation explain the shift from CLI usage to Go package invocation and identify the supported import/package path

## Future Requirements

Deferred to future releases. Tracked but not in current roadmap.

### Compatibility and Protocol Expansion

- **FUT-01**: Developer can optionally parse legacy YAML or CLI-style FTP URLs into typed library options
- **FUT-02**: Developer can run background FTP→local polling sync through the library
- **FUT-03**: Developer can use FTPS/TLS endpoints with certificate and security configuration
- **FUT-04**: Developer can synchronize FTP↔FTP or bidirectional flows with explicit conflict handling
- **FUT-05**: Developer can use resumable transfers or connection pooling when server capabilities allow

## Out of Scope

Explicitly excluded from this milestone.

| Feature | Reason |
|---------|--------|
| CLI runtime compatibility | v2.0 intentionally moves to package invocation and drops CLI as a supported runtime |
| Legacy YAML/CLI parser in public API | User selected typed Go options only for v2.0 |
| Background FTP→disk polling | User selected background support only for disk→FTP in v2.0 |
| HTTP/gRPC/file server/task/auth/session surfaces | These are application runtime surfaces, not required for focused FTP sync library usage |
| SFTP and MinIO support | v2.0 library scope is FTP only |
| FTPS and active FTP | Protocol/security expansion is deferred beyond library extraction |
| FTP server mode | The package remains a client-side FTP sync library |
| FTP↔FTP and bidirectional conflict resolution | Remote-to-remote and conflict semantics expand scope beyond the focused extraction |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| API-01 | Phase 5 | Pending |
| API-02 | Phase 5 | Pending |
| API-03 | Phase 5 | Pending |
| API-04 | Phase 5 | Pending |
| API-05 | Phase 5 | Pending |
| ONCE-01 | Phase 6 | Pending |
| ONCE-02 | Phase 6 | Pending |
| ONCE-03 | Phase 6 | Pending |
| ONCE-04 | Phase 6 | Pending |
| ONCE-05 | Phase 6 | Pending |
| WATCH-01 | Phase 7 | Pending |
| WATCH-02 | Phase 7 | Pending |
| WATCH-03 | Phase 7 | Pending |
| WATCH-04 | Phase 7 | Pending |
| WATCH-05 | Phase 7 | Pending |
| PRUNE-01 | Phase 8 | Pending |
| PRUNE-02 | Phase 8 | Pending |
| PRUNE-03 | Phase 8 | Pending |
| PRUNE-04 | Phase 8 | Pending |
| VERIFY-01 | Phase 9 | Pending |
| VERIFY-02 | Phase 9 | Pending |
| VERIFY-03 | Phase 9 | Pending |
| DOC-01 | Phase 9 | Pending |
| DOC-02 | Phase 9 | Pending |
| DOC-03 | Phase 9 | Pending |

**Coverage:**
- v2.0 requirements: 25 total
- Mapped to phases: 25
- Unmapped: 0

---
*Requirements defined: 2026-04-27*
*Last updated: 2026-04-27 after v2.0 roadmap creation*
