# Roadmap: gofs

## Milestones

- ✅ **v1.0 FTP client sync support** - Phases 1-4 (shipped 2026-04-27)
- 🚧 **v2.0 FTP Sync Library** - Phases 5-9 (in progress)

## Overview

v2.0 converts the existing FTP sync capability from a CLI/server-centered application feature into a focused Go library. The work starts by fixing the public `ftpsync`/`FTPSyncService` API boundary, proves one-shot disk<->FTP behavior through that API, adds only the approved persistent disk→FTP background mode, then prunes unrelated runtime surfaces before hardening tests and documentation around the final package contract.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

<details>
<summary>✅ v1.0 FTP client sync support (Phases 1-4) - SHIPPED 2026-04-27</summary>

See archived milestone artifacts:

- Roadmap: `.planning/milestones/v1.0-ROADMAP.md`
- Requirements: `.planning/milestones/v1.0-REQUIREMENTS.md`
- Audit: `.planning/milestones/v1.0-MILESTONE-AUDIT.md`
- Phase artifacts: `.planning/milestones/v1.0-phases/`

</details>

### 🚧 v2.0 FTP Sync Library (In Progress)

**Milestone Goal:** Provide a focused Go package for reliable FTP file synchronization without requiring the existing CLI/server runtime.

- [x] **Phase 5: Public FTPSyncService API Contract** - Developers can import `ftpsync`, configure `FTPSyncService` with typed Go options, and receive explicit validation/errors without invoking old runtime code. (completed 2026-04-27)
- [ ] **Phase 6: One-Shot Disk↔FTP Sync Through Library API** - Developers can run one-shot local→FTP and FTP→local sync through `FTPSyncService` while preserving v1 FTP behavior and cwd safety.
- [ ] **Phase 7: Background Disk→FTP Lifecycle** - Developers can start, observe, and stop persistent local disk→FTP sync with deterministic cleanup and no FTP→disk polling API.
- [ ] **Phase 8: FTP-Only Package Reduction** - The library build exposes only FTP sync capabilities and drops or isolates unrelated CLI/server/protocol dependencies.
- [ ] **Phase 9: Verification, Examples, and Migration Docs** - Tests and docs prove the library contract, real FTP flows, limitations, and migration path.

## Phase Details

### Phase 5: Public FTPSyncService API Contract
**Goal**: Developers can construct and validate a focused `ftpsync.FTPSyncService` using typed Go options without invoking CLI, server, daemon, or global reporting runtime code.
**Depends on**: Phase 4
**Requirements**: API-01, API-02, API-03, API-04, API-05
**Success Criteria** (what must be TRUE):
  1. Developer can import package `ftpsync` and construct an `FTPSyncService` from typed Go options only.
  2. Developer can configure local path, FTP host/port/credentials/remote path, passive mode, timeout, path encoding, retry behavior, and ignore rules without YAML, CLI flags, or URL parser requirements.
  3. Developer receives explicit validation failures for unsupported or ambiguous endpoint combinations before transfer work starts.
  4. Developer can pass `context.Context` to public sync methods and distinguish validation, cancellation, connection/authentication, transfer, and unsupported-capability errors.
  5. Developer can attach optional logging, progress, and sync event hooks, while the default service remains no-op and library-local.
**Plans:** 3 plans
Plans:
- [x] 05-01-PLAN.md — Create the public ftpsync typed option surface and FTPSyncService constructor without legacy runtime imports.
- [x] 05-02-PLAN.md — Add explicit validation plus context-aware public method and structured error contracts.
- [x] 05-03-PLAN.md — Add optional no-op logging, progress, and sync event hooks isolated from global runtime reporting.

### Phase 6: One-Shot Disk↔FTP Sync Through Library API
**Goal**: Developers can run one-shot local disk→FTP and FTP→local disk synchronization through `FTPSyncService` with the existing FTP v1 transfer semantics preserved.
**Depends on**: Phase 5
**Requirements**: ONCE-01, ONCE-02, ONCE-03, ONCE-04, ONCE-05
**Success Criteria** (what must be TRUE):
  1. Developer can call `FTPSyncService` to perform a one-shot local disk→FTP synchronization without invoking the CLI.
  2. Developer can call `FTPSyncService` to perform a one-shot FTP→local disk synchronization without invoking the CLI.
  3. One-shot sync preserves create, update, delete, rename-relevant, nested path, passive-mode, timeout, and path-encoding behavior from FTP v1.
  4. FTP→local sync writes only under the explicitly configured local endpoint and never silently falls back to the process working directory.
  5. Developer receives a useful structured result summary after a one-shot run instead of relying on terminal output.
**Plans**: 3 plans

Plans:
- [x] 06-01-PLAN.md — Add one-shot result/partial-failure scaffolding and the typed-options adapter into the legacy FTP sync engine.
- [x] 06-02-PLAN.md — Implement local→FTP one-shot execution through FTPSyncService while preserving FTP v1 push semantics.
- [x] 06-03-PLAN.md — Implement FTP→local one-shot execution with explicit destination-root and cwd-safety regression coverage.

### Phase 7: Background Disk→FTP Lifecycle
**Goal**: Developers can run persistent local disk→FTP synchronization from the library with observable lifecycle controls and deterministic shutdown, while FTP→disk background polling remains unavailable in v2.0.
**Depends on**: Phase 6
**Requirements**: WATCH-01, WATCH-02, WATCH-03, WATCH-04, WATCH-05
**Success Criteria** (what must be TRUE):
  1. Developer can start background local disk→FTP monitoring through `FTPSyncService` without invoking the CLI monitor runtime.
  2. Local create, update, delete, and rename-relevant filesystem changes are applied to the FTP destination while the background run is active.
  3. Developer receives a lifecycle handle that exposes readiness, error observation, wait, and deterministic shutdown behavior.
  4. Cancelling the context or stopping the handle closes watchers, timers, retry sleeps, goroutines, and FTP connections.
  5. Developer cannot start FTP→disk polling or bidirectional background conflict resolution through the v2.0 public API.
**Plans**: 3 plans

Plans:
- [x] 07-01-PLAN.md — Add the public StartBackground local→FTP lifecycle entrypoint, stronger handle contract, and initial catch-up startup behavior.
- [ ] 07-02-PLAN.md — Add recursive fsnotify watching, event debounce/coalescing, and non-terminal steady-state sync error handling.
- [ ] 07-03-PLAN.md — Finalize deterministic stop/cancel cleanup, final wait/error semantics, and shutdown regression coverage.

### Phase 8: FTP-Only Package Reduction
**Goal**: Library consumers see a small FTP-only package surface whose build graph excludes old CLI/server/protocol runtimes while retaining the internal helpers needed for disk↔FTP sync.
**Depends on**: Phase 7
**Requirements**: PRUNE-01, PRUNE-02, PRUNE-03, PRUNE-04
**Success Criteria** (what must be TRUE):
  1. Importing and testing the library no longer requires CLI entrypoints, flag parsing, daemon/process management, HTTP file server, gRPC APIs, task runtime, auth/session, SFTP, MinIO, or Docker release surfaces.
  2. FTP driver, path encoding, retry, ignore/filtering, rate limiting, and core disk/FTP sync internals remain usable behind `FTPSyncService`.
  3. Public `ftpsync` APIs do not expose old `conf.Config`, `core.VFS`, CLI URL parsing, or server/task types.
  4. `go.mod` no longer retains unnecessary Gin, gRPC, SFTP, MinIO, Redis/cache, QUIC, OAuth, or protobuf dependencies for the library build.
**Plans**: TBD

### Phase 9: Verification, Examples, and Migration Docs
**Goal**: Developers can trust and adopt the new FTP sync library because automated tests, real FTP integration coverage, examples, and migration notes match the final v2.0 package contract.
**Depends on**: Phase 8
**Requirements**: VERIFY-01, VERIFY-02, VERIFY-03, DOC-01, DOC-02, DOC-03
**Success Criteria** (what must be TRUE):
  1. Automated tests cover `FTPSyncService` construction, validation, one-shot push, one-shot pull, and background disk→FTP lifecycle behavior.
  2. Real FTP integration tests verify library-based local→FTP and FTP→local one-shot flows without the CLI.
  3. Regression tests prove cwd safety, FTP path encoding, passive-mode defaults, cancellation, and background goroutine shutdown.
  4. README and package examples show typed-option usage for one-shot push, one-shot pull, and background disk→FTP sync.
  5. Migration or release notes clearly state the supported import/package path and removed/unsupported v2.0 surfaces.
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 5 → 6 → 7 → 8 → 9

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. FTP Endpoint Contract & Routing | v1.0 | - | Complete | 2026-04-27 |
| 2. FTP Driver Backend | v1.0 | - | Complete | 2026-04-27 |
| 3. One-Way FTP Sync Flows | v1.0 | - | Complete | 2026-04-27 |
| 4. FTP Verification & Discoverability | v1.0 | - | Complete | 2026-04-27 |
| 5. Public FTPSyncService API Contract | v2.0 | 3/3 | Complete | 2026-04-27 |
| 6. One-Shot Disk↔FTP Sync Through Library API | v2.0 | 3/3 | Complete | 2026-04-27 |
| 7. Background Disk→FTP Lifecycle | v2.0 | 0/3 | Planned | - |
| 8. FTP-Only Package Reduction | v2.0 | 0/TBD | Not started | - |
| 9. Verification, Examples, and Migration Docs | v2.0 | 0/TBD | Not started | - |
