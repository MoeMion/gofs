# Requirements: gofs

**Defined:** 2026-04-23
**Core Value:** Add FTP as a first-class sync endpoint with the smallest correct change set, so gofs can cover one more common file transfer protocol without disrupting the existing architecture.

## v1 Requirements

Requirements for the FTP support milestone. Each will map to exactly one roadmap phase.

### FTP Endpoint Configuration

- [x] **FTP-01**: User can configure an FTP endpoint as a sync source using host, port, username, password, and remote path
- [x] **FTP-02**: User can configure an FTP endpoint as a sync destination using host, port, username, password, and remote path
- [x] **FTP-03**: User can configure FTP connection timeout behavior for an endpoint
- [x] **FTP-04**: User can configure passive-mode-compatible FTP behavior for an endpoint

### FTP Driver Capabilities

- [x] **FTPD-01**: System can connect to an FTP server and authenticate with configured credentials
- [x] **FTPD-02**: System can recursively list and traverse files and directories on an FTP endpoint
- [x] **FTPD-03**: System can upload a file from local storage to an FTP endpoint
- [x] **FTPD-04**: System can download a file from an FTP endpoint to local storage
- [x] **FTPD-05**: System can create required directories on an FTP endpoint during sync
- [x] **FTPD-06**: System can delete files and directories on an FTP endpoint when sync policy requires removal
- [x] **FTPD-07**: System can rename files or directories on an FTP endpoint when sync flow requires rename handling
- [x] **FTPD-08**: System can compare FTP-side file state using size and modification time with documented precision caveats
- [x] **FTPD-09**: System can recover from transient FTP connection failures using conservative reconnect or retry behavior

### Sync Flows

- [ ] **SYNC-01**: User can run sync from local disk to an FTP destination
- [ ] **SYNC-02**: User can run sync from an FTP source to local disk
- [ ] **SYNC-03**: System preserves existing one-way sync semantics for FTP-backed flows without introducing bidirectional conflict resolution
- [ ] **SYNC-04**: A second sync run with no file changes does not produce unnecessary file transfers for supported FTP metadata conditions

### Verification and Documentation

- [ ] **TEST-01**: Automated tests verify `disk→FTP` sync behavior against an FTP test server
- [ ] **TEST-02**: Automated tests verify `FTP→disk` sync behavior against an FTP test server
- [ ] **TEST-03**: Automated tests cover nested paths plus delete or rename behavior on the FTP path
- [ ] **DOC-01**: User-facing documentation includes at least one working FTP configuration example
- [ ] **DOC-02**: User-facing documentation states the v1 FTP limitations, including plain FTP only and no FTP↔FTP sync

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### FTP Compatibility and Security

- **FTPV2-01**: User can connect to FTPS endpoints with TLS configuration
- **FTPV2-02**: User can enable active FTP mode when passive-mode-compatible behavior is not sufficient
- **FTPV2-03**: User can tune compatibility flags for server quirks such as EPSV, MLSD, UTF-8, or MDTM handling
- **FTPV2-04**: System can resume interrupted FTP transfers when server capabilities allow restart support
- **FTPV2-05**: User can connect anonymously to read-only FTP endpoints when required

## Out of Scope

Explicitly excluded from the current roadmap.

| Feature | Reason |
|---------|--------|
| FTP server mode | This initiative adds FTP as a client-side sync backend, not a server surface |
| FTP↔FTP sync | Remote-to-remote support expands combinations and complexity beyond the minimal brownfield increment |
| Bidirectional FTP conflict resolution | The current goal is one-way sync parity, not a larger sync-engine semantic expansion |
| Major FTP-specific UI expansion | The current scope is configuration, backend implementation, tests, and minimal discoverability docs |

## Traceability

Which phases cover which requirements. Updated during roadmap creation.

| Requirement | Phase | Status |
|-------------|-------|--------|
| FTP-01 | Phase 1 | Complete |
| FTP-02 | Phase 1 | Complete |
| FTP-03 | Phase 1 | Complete |
| FTP-04 | Phase 1 | Complete |
| FTPD-01 | Phase 2 | Complete |
| FTPD-02 | Phase 2 | Complete |
| FTPD-03 | Phase 2 | Complete |
| FTPD-04 | Phase 2 | Complete |
| FTPD-05 | Phase 2 | Complete |
| FTPD-06 | Phase 2 | Complete |
| FTPD-07 | Phase 2 | Complete |
| FTPD-08 | Phase 2 | Complete |
| FTPD-09 | Phase 2 | Complete |
| SYNC-01 | Phase 3 | Pending |
| SYNC-02 | Phase 3 | Pending |
| SYNC-03 | Phase 3 | Pending |
| SYNC-04 | Phase 3 | Pending |
| TEST-01 | Phase 4 | Pending |
| TEST-02 | Phase 4 | Pending |
| TEST-03 | Phase 4 | Pending |
| DOC-01 | Phase 4 | Pending |
| DOC-02 | Phase 4 | Pending |

**Coverage:**
- v1 requirements: 22 total
- Mapped to phases: 22
- Unmapped: 0 ✓

---
*Requirements defined: 2026-04-23*
*Last updated: 2026-04-23 after roadmap creation*
