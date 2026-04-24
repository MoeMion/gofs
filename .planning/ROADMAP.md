# Roadmap: gofs

## Overview

This roadmap adds FTP to gofs as a minimal-change client-side sync backend by extending the existing VFS, driver, sync, and monitor patterns already used for other protocols. The work moves from making FTP endpoints configurable and routable, to implementing the backend driver, to enabling one-way sync flows, and finally to proving correctness with automated tests and clear user documentation.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: FTP Endpoint Contract & Routing** - Make FTP endpoints configurable and reachable through existing factory paths.
- [x] **Phase 2: FTP Driver Backend** - Add the FTP backend operations the sync engine depends on. (completed 2026-04-24)
- [x] **Phase 3: One-Way FTP Sync Flows** - Enable disk→FTP and FTP→disk flows using existing sync semantics. (completed 2026-04-24)
- [x] **Phase 4: FTP Verification & Discoverability** - Prove the FTP path with automated tests and user-facing documentation. (completed 2026-04-24)

## Phase Details

### Phase 1: FTP Endpoint Contract & Routing
**Goal**: Users can define FTP endpoints in config and have gofs recognize them as valid source or destination sync targets.
**Depends on**: Nothing (first phase)
**Requirements**: FTP-01, FTP-02, FTP-03, FTP-04
**Success Criteria** (what must be TRUE):
  1. User can configure an FTP endpoint as either the source or destination of a sync using host, port, username, password, and remote path.
  2. User can set FTP timeout and passive-mode-compatible behavior per endpoint in configuration.
  3. A configured `ftp://` endpoint is accepted by gofs and routed into the existing sync and monitor selection flow instead of being treated as unsupported.
**Plans**: 2 plans

Plans:
- [x] 01-01-PLAN.md — Define the FTP VFS contract, defaults, and automated parsing coverage.
- [x] 01-02-PLAN.md — Route FTP endpoints through sync and monitor factories with thin FTP entry points.

### Phase 2: FTP Driver Backend
**Goal**: FTP endpoints behave like a usable remote storage backend inside the existing driver abstraction.
**Depends on**: Phase 1
**Requirements**: FTPD-01, FTPD-02, FTPD-03, FTPD-04, FTPD-05, FTPD-06, FTPD-07, FTPD-08, FTPD-09
**Success Criteria** (what must be TRUE):
  1. A configured FTP endpoint can connect and authenticate successfully with the supplied credentials.
  2. The sync engine can inspect nested files and directories on FTP endpoints well enough to compare remote state against local state.
  3. The system can upload, download, create directories, delete entries, and rename entries on FTP endpoints when sync behavior requires those actions.
  4. FTP-backed comparisons and operations remain usable across normal transient connection interruptions, with documented size and modification-time comparison caveats.
**Plans**: 2 plans

Plans:
- [x] 02-01-PLAN.md — Implement the FTP driver package with conservative metadata and bounded reconnect behavior.
- [x] 02-02-PLAN.md — Replace FTP sync placeholders with driver-backed push/pull constructors and regression tests.

### Phase 3: One-Way FTP Sync Flows
**Goal**: Users can run the intended one-way sync workflows between local disk and FTP without changing gofs sync semantics.
**Depends on**: Phase 2
**Requirements**: SYNC-01, SYNC-02, SYNC-03, SYNC-04
**Success Criteria** (what must be TRUE):
  1. User can run a sync from local disk to an FTP destination and see expected file additions and updates appear on the FTP endpoint.
  2. User can run a sync from an FTP source to local disk and see expected file additions and updates appear on the local filesystem.
  3. FTP-backed sync runs preserve existing one-way behavior, including delete or rename handling where supported, without introducing bidirectional conflict resolution.
  4. A second sync run with no eligible file changes does not perform unnecessary transfers under the supported FTP metadata conditions.
**Plans**: 2 plans

Plans:
- [x] 03-01-PLAN.md — Replace the FTP source monitor placeholder with a real polling monitor and truthful startup behavior.
- [x] 03-02-PLAN.md — Lock down one-way FTP flow semantics and conservative no-op behavior with targeted sync tests.

### Phase 4: FTP Verification & Discoverability
**Goal**: Users and maintainers can trust and adopt the FTP path because it is tested against realistic flows and documented clearly.
**Depends on**: Phase 3
**Requirements**: TEST-01, TEST-02, TEST-03, DOC-01, DOC-02
**Success Criteria** (what must be TRUE):
  1. Automated tests verify both `disk→FTP` and `FTP→disk` flows against an FTP test server.
  2. Automated coverage proves nested path handling plus delete or rename behavior on the FTP path.
  3. User-facing documentation includes at least one working FTP configuration example.
  4. User-facing documentation clearly states the v1 FTP limitations, including plain FTP only and no FTP↔FTP sync.
**Plans**: 3 plans

Plans:
- [x] 04-01-PLAN.md — Add repo-owned FTP integration bootstrap and fixture files for real push/pull verification.
- [x] 04-02-PLAN.md — Add the tagged FTP integration suite and CI workflow execution.
- [x] 04-03-PLAN.md — Add FTP push/pull examples and explicit v1 limitations to both READMEs.

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. FTP Endpoint Contract & Routing | 1/2 | In Progress | - |
| 2. FTP Driver Backend | 2/2 | Complete   | 2026-04-24 |
| 3. One-Way FTP Sync Flows | 2/2 | Complete | 2026-04-24 |
| 4. FTP Verification & Discoverability | 3/3 | Complete | 2026-04-24 |
