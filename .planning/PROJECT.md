# gofs

## What This Is

gofs is an existing Go-based file synchronization system for local and remote filesystem replication. It supports local disk sync, remote event-driven sync, HTTP and gRPC server surfaces, and storage backends such as SFTP, MinIO, and FTP.

## Core Value

Keep protocol support broad while preserving the existing minimal-change sync architecture.

## Requirements

### Validated

- ✓ Local filesystem synchronization with change monitoring and copy orchestration — existing
- ✓ Remote synchronization over HTTP and gRPC server/client flows — existing
- ✓ Config-driven runtime with CLI entrypoints, background monitoring, and optional web/file server modes — existing
- ✓ SFTP-backed sync support through the driver and sync abstractions — existing
- ✓ MinIO/S3-compatible storage sync support through the driver and sync abstractions — existing
- ✓ FTP can be configured as a sync source endpoint — v1.0
- ✓ FTP can be configured as a sync destination endpoint — v1.0
- ✓ FTP connection setup supports host, port, username, password, passive mode, timeout, and path encoding controls — v1.0
- ✓ FTP disk<->endpoint flows are covered by automated unit and real-server integration tests — v1.0
- ✓ FTP usage is discoverable in CLI/configuration documentation — v1.0

### Active

No active milestone requirements are defined.

### Out of Scope

- FTPS support — explicitly deferred to keep v1 focused on minimal FTP protocol enablement
- Running gofs as an FTP server — not required for this increment
- Non-essential UI expansion — the work should stay inside config, protocol implementation, tests, and minimal documentation

## Context

- The repository is a brownfield Go module with established abstractions around `core.VFS`, `driver.Driver`, `sync.Sync`, and `monitor.Monitor`.
- Existing protocol support already includes disk, remote-disk server/client flows, SFTP, and MinIO, which suggests FTP should be added by aligning with the current driver and sync patterns rather than introducing a new execution model.
- FTP v1.0 shipped as a client-side storage/sync option that can act as either source or destination.
- The FTP path fits beside SFTP and MinIO through the existing VFS, driver, sync, and monitor abstractions without a broad sync-engine refactor.
- The next milestone has not been defined yet.

## Constraints

- **Architecture**: Prefer minimal modifications to the existing VFS, driver, sync, and config layers — the change should reuse current abstractions where possible.
- **Scope**: FTP v1 must support both source and destination usage — the protocol cannot be added as a one-sided experiment.
- **Security**: Plain FTP only in v1 — FTPS is deferred to avoid expanding protocol and certificate handling scope.
- **Compatibility**: Connection configuration must cover host, port, username, password, passive mode, and timeout controls — these are required for practical interoperability.
- **Quality**: The new protocol path needs automated coverage and enough documentation for users to discover and configure it correctly.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Implement FTP as client-side sync support, not server-side FTP exposure | Matches the current need and keeps the change aligned with existing backend protocol integrations | Shipped in v1.0 |
| Support FTP as both source and destination in v1 | Avoids shipping an incomplete protocol mode that breaks symmetry with existing storage backends | Shipped in v1.0 |
| Start with plain FTP, not FTPS | Minimizes surface area and keeps the first phase focused on protocol integration | Shipped in v1.0; FTPS remains out of scope |
| Reuse existing driver/sync architecture with minimal code changes | The repository already has clear extension points for remote backends such as SFTP and MinIO | Shipped in v1.0 |
| Default FTP to passive mode and auto path encoding | Matches common FTP deployment behavior while keeping compatibility controls available | Added after v1.0 implementation as compatibility hardening |

## Evolution

This document evolves at phase transitions and milestone boundaries.

**After each phase transition** (via `/gsd-transition`):
1. Requirements invalidated? → Move to Out of Scope with reason
2. Requirements validated? → Move to Validated with phase reference
3. New requirements emerged? → Add to Active
4. Decisions to log? → Add to Key Decisions
5. "What This Is" still accurate? → Update if drifted

**After each milestone** (via `/gsd-complete-milestone`):
1. Full review of all sections
2. Core Value check — still the right priority?
3. Audit Out of Scope — reasons still valid?
4. Update Context with current state

---
*Last updated: 2026-04-27 after v1.0 milestone completion*
