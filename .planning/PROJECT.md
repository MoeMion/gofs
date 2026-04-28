# gofs

## What This Is

gofs is now a focused Go FTP synchronization library. The broad legacy CLI/server/runtime application surfaces have been removed from the active package, and the supported local-library import path is `ftpsync/ftpsync`.

## Core Value

Provide a focused Go package for reliable FTP file synchronization without requiring the old CLI/server runtime.

## Current State

**Shipped:** v2.0 FTP Sync Library on 2026-04-28.

The current module exposes `FTPSyncService` with typed Go options for FTP connection details, local paths, remote paths, direction, retry, ignore rules, timeout, passive mode, path encoding, and optional hooks.

Supported v2.0 flows:

- One-shot local disk->FTP synchronization.
- One-shot FTP->local disk synchronization with cwd-safety guarantees.
- Background local disk->FTP synchronization with lifecycle handle, readiness, error observation, wait, and deterministic shutdown.

The reduced root module package set is `ftpsync/ftpsync`; old CLI, HTTP/gRPC server, task, auth/session, SFTP, MinIO, Docker, and release surfaces are no longer active runtime surfaces.

## Requirements

### Validated

- ✓ Local filesystem synchronization with change monitoring and copy orchestration — existing before v2.0
- ✓ Remote synchronization over HTTP and gRPC server/client flows — existing before v2.0, removed from active v2.0 package scope
- ✓ Config-driven runtime with CLI entrypoints, background monitoring, and optional web/file server modes — existing before v2.0, removed from active v2.0 package scope
- ✓ SFTP-backed sync support through the driver and sync abstractions — existing before v2.0, removed from active v2.0 package scope
- ✓ MinIO/S3-compatible storage sync support through the driver and sync abstractions — existing before v2.0, removed from active v2.0 package scope
- ✓ FTP can be configured as a sync source endpoint — v1.0
- ✓ FTP can be configured as a sync destination endpoint — v1.0
- ✓ FTP connection setup supports host, port, username, password, passive mode, timeout, and path encoding controls — v1.0
- ✓ FTP disk<->endpoint flows are covered by automated unit and real-server integration tests — v1.0
- ✓ FTP usage is discoverable in CLI/configuration documentation — v1.0
- ✓ FTP sync can be invoked through a Go `FTPSyncService` instead of the CLI — v2.0
- ✓ `FTPSyncService` accepts typed Go options for endpoint, retry, ignore, and hook configuration — v2.0
- ✓ Library consumers can run one-shot FTP sync flows with compact `Result + error` semantics — v2.0
- ✓ Library consumers can run background persistent disk->FTP monitoring/sync flows — v2.0
- ✓ Non-FTP protocols and non-library runtime surfaces are removed from the active package API and build graph — v2.0
- ✓ README, compiler-checked examples, real FTP integration tests, and migration notes match the final local-library contract — v2.0

### Active

- None. Fresh requirements should be defined by the next milestone.

### Out of Scope

- FTPS/TLS support — deferred beyond v2.0 to avoid expanding protocol and certificate handling scope.
- Running gofs as an FTP server — the package remains client-side FTP sync.
- Preserving the CLI as a supported runtime — v2.0 intentionally shifts to package-based invocation.
- Preserving SFTP, MinIO, HTTP, gRPC, task, auth/session, and file-server runtime surfaces — v2.0 is scoped to FTP sync library extraction.
- Legacy YAML/CLI parser support in the public API — typed Go options are the supported v2.0 configuration path.
- Background FTP->local polling — background support is local disk->FTP only in v2.0.
- FTP<->FTP and bidirectional conflict resolution — remote-to-remote and conflict semantics remain future scope.

## Context

- v1.0 added FTP as a first-class client sync endpoint within the original gofs architecture.
- v2.0 intentionally broke from the previous broad product shape and extracted a much smaller FTP-focused library.
- The final package uses `module ftpsync` with source in `ftpsync/`, so the supported import path is `ftpsync/ftpsync`.
- Verification includes full `go test ./...`, real loopback FTP integration tests for both one-shot directions, compiler-checked examples, dependency-boundary guards, and migration documentation.
- Known accepted tech debt at close is planning-source drift that was reconciled during archiving: ROADMAP, REQUIREMENTS archive, PROJECT, and STATE were updated to match the completed implementation.

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Implement FTP as client-side sync support, not server-side FTP exposure | Matches the current need and keeps the change aligned with existing backend protocol integrations | Shipped in v1.0 |
| Support FTP as both source and destination in v1 | Avoids shipping an incomplete protocol mode that breaks symmetry with existing storage backends | Shipped in v1.0 |
| Start with plain FTP, not FTPS | Minimizes surface area and keeps the first increment focused on protocol integration | Shipped in v1.0; FTPS remains out of scope |
| Shift v2.0 to a Go FTP sync library | The desired consumer path is package invocation through `FTPSyncService`, not CLI operation | Shipped in v2.0 |
| Use typed Go options only in the public API | Keeps library consumers off legacy CLI/YAML parser semantics and makes validation explicit | Shipped in v2.0 |
| Keep one-shot sync results compact and summary-oriented | Avoids inflating the public API with per-file report structures while still giving callers actionable execution feedback | Shipped in v2.0 |
| One-shot sync should be best-effort and return `Result + error` on partial failure | Preserves visibility into completed work and later-path continuation for callers | Shipped in v2.0 |
| Restrict background sync in v2.0 to disk->FTP only | Matches approved scope and keeps FTP->disk polling out of the public lifecycle API | Shipped in v2.0 |
| Use coalesced full-pass background sync after fsnotify events | Reuses one-shot transfer semantics and avoids duplicating per-event FTP mutation logic | Shipped in v2.0 |
| Remove non-FTP runtime surfaces from the active module | Reduces package size and API complexity for the library use case | Shipped in v2.0 |
| Keep source in `ftpsync/` under `module ftpsync` | Resolves the actual Go import path for the existing layout without a package move | Shipped as `ftpsync/ftpsync` in v2.0 |

## Next Milestone Goals

No next milestone is currently defined. Candidate future directions include FTPS/TLS, FTP->local background polling, compatibility parsers for legacy config/URLs, resumable transfers, connection pooling, or conflict-aware remote flows.

---
*Last updated: 2026-04-28 after v2.0 milestone completion*
