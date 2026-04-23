# gofs

## What This Is

gofs is an existing Go-based file synchronization system for local and remote filesystem replication. It already supports local disk sync, remote event-driven sync, HTTP and gRPC server surfaces, and storage backends such as SFTP and MinIO; the current project focus is to extend that protocol matrix with minimal-change FTP client sync support.

## Core Value

Add FTP as a first-class sync endpoint with the smallest correct change set, so gofs can cover one more common file transfer protocol without disrupting the existing architecture.

## Requirements

### Validated

- ✓ Local filesystem synchronization with change monitoring and copy orchestration — existing
- ✓ Remote synchronization over HTTP and gRPC server/client flows — existing
- ✓ Config-driven runtime with CLI entrypoints, background monitoring, and optional web/file server modes — existing
- ✓ SFTP-backed sync support through the driver and sync abstractions — existing
- ✓ MinIO/S3-compatible storage sync support through the driver and sync abstractions — existing

### Active

- [ ] FTP can be configured as a sync source endpoint
- [ ] FTP can be configured as a sync destination endpoint
- [ ] FTP connection setup supports host, port, username, and password
- [ ] FTP connection behavior supports passive mode selection and timeout configuration
- [ ] FTP flows are covered by automated tests on the new protocol path
- [ ] FTP usage is discoverable in CLI/configuration documentation

### Out of Scope

- FTPS support — explicitly deferred to keep v1 focused on minimal FTP protocol enablement
- Running gofs as an FTP server — not required for this increment
- Non-essential UI expansion — the work should stay inside config, protocol implementation, tests, and minimal documentation

## Context

- The repository is a brownfield Go module with established abstractions around `core.VFS`, `driver.Driver`, `sync.Sync`, and `monitor.Monitor`.
- Existing protocol support already includes disk, remote-disk server/client flows, SFTP, and MinIO, which suggests FTP should be added by aligning with the current driver and sync patterns rather than introducing a new execution model.
- The desired change is intentionally narrow: treat FTP as a new client-side storage/sync option that can act as either source or destination.
- The motivation is protocol completeness rather than a product pivot: FTP support should fit naturally beside SFTP and MinIO and avoid broad refactors.

## Constraints

- **Architecture**: Prefer minimal modifications to the existing VFS, driver, sync, and config layers — the change should reuse current abstractions where possible.
- **Scope**: FTP v1 must support both source and destination usage — the protocol cannot be added as a one-sided experiment.
- **Security**: Plain FTP only in v1 — FTPS is deferred to avoid expanding protocol and certificate handling scope.
- **Compatibility**: Connection configuration must cover host, port, username, password, passive mode, and timeout controls — these are required for practical interoperability.
- **Quality**: The new protocol path needs automated coverage and enough documentation for users to discover and configure it correctly.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Implement FTP as client-side sync support, not server-side FTP exposure | Matches the current need and keeps the change aligned with existing backend protocol integrations | — Pending |
| Support FTP as both source and destination in v1 | Avoids shipping an incomplete protocol mode that breaks symmetry with existing storage backends | — Pending |
| Start with plain FTP, not FTPS | Minimizes surface area and keeps the first phase focused on protocol integration | — Pending |
| Reuse existing driver/sync architecture with minimal code changes | The repository already has clear extension points for remote backends such as SFTP and MinIO | — Pending |

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
*Last updated: 2026-04-23 after initialization*
