# Codebase Structure

**Analysis Date:** 2026-04-23

## Directory Layout

```text
gofs/
├── cmd/             # CLI entrypoints and runtime bootstrap
├── conf/            # Config models and parsers
├── core/            # Shared value types and VFS modeling
├── monitor/         # Event sources and orchestration loops
├── sync/            # Concrete sync engines by transport pair
├── server/          # HTTP server, handlers, middleware, templates
├── api/             # gRPC APIs, generated protobuf code, client/server wrappers
├── driver/          # Remote filesystem drivers for SFTP and MinIO
├── auth/            # User, permission, and session models
├── report/          # Runtime stats collection and reporting models
├── integration/     # End-to-end and scenario-based tests with test configs
├── internal/        # Internal-only support packages and version assets
├── scripts/         # Release, docker, env, and coverage helper scripts
└── .planning/       # Generated planning and codebase map documents
```

## Directory Purposes

**`cmd/`:**
- Purpose: Hold executable entrypoints and application assembly.
- Contains: `cmd/gofs/main.go`, `cmd/gofs.go`, `cmd/logger.go`.
- Key files: `cmd/gofs.go`, `cmd/gofs/main.go`.

**`conf/`:**
- Purpose: Parse and serialize runtime config.
- Contains: Config structs, parser helpers, format conversion.
- Key files: `conf/config.go`, `conf/parser.go`, `conf/format.go`.

**`core/`:**
- Purpose: Define shared primitive types used across config and sync logic.
- Contains: `VFS`, duration, size, path, and flag helpers.
- Key files: `core/vfs.go`, `core/vfs_type.go`, `core/flagset.go`.

**`monitor/`:**
- Purpose: Convert change sources into sync actions.
- Contains: Local fsnotify monitor, remote client monitor, task monitor, transport-specific pull monitors.
- Key files: `monitor/monitor.go`, `monitor/base_monitor.go`, `monitor/fsnotify_monitor.go`, `monitor/remote_client_monitor.go`.

**`sync/`:**
- Purpose: Execute filesystem mutations and transfers.
- Contains: Factory, base sync, disk sync, remote server/client sync, push sync, SFTP and MinIO syncers, driver-backed syncers.
- Key files: `sync/sync.go`, `sync/disk_sync.go`, `sync/remote_server_sync.go`, `sync/remote_client_sync.go`, `sync/push_client_sync.go`.

**`server/`:**
- Purpose: Provide HTTP access and management features.
- Contains: Route constants, option structs, embedded templates, HTTPFS server, handlers, middleware, sessions.
- Key files: `server/httpfs/file_server.go`, `server/handler/push_handler.go`, `server/session.go`, `server/resource.go`.

**`api/`:**
- Purpose: Provide internal gRPC APIs and generated stubs.
- Contains: gRPC client/server wrappers plus `auth`, `info`, `monitor`, `task`, and protobuf sources/generated code.
- Key files: `api/apiserver/grpc_server.go`, `api/apiclient/grpc_client.go`, `api/task/dispatcher.go`, `api/README.md`.

**`driver/`:**
- Purpose: Abstract remote storage providers behind a shared interface.
- Contains: Generic `driver.Driver` plus `driver/sftp/` and `driver/minio/` implementations.
- Key files: `driver/driver.go`, `driver/sftp/sftp.go`, `driver/minio/minio.go`.

**`auth/`:**
- Purpose: Model users, permissions, and session users.
- Contains: Permission parsing, user parsing, session types.
- Key files: `auth/user.go`, `auth/perm.go`, `auth/session_user.go`.

**`report/`:**
- Purpose: Aggregate runtime metrics and expose report payloads.
- Contains: Reporter implementation and report/stat data types.
- Key files: `report/reporter.go`, `report/report.go`, `report/api_stat.go`, `report/conn_stat.go`.

**`integration/`:**
- Purpose: Exercise multi-process and multi-backend behavior.
- Contains: Integration tests and YAML config fixtures.
- Key files: `integration/integration_local_test.go`, `integration/integration_remote_test.go`, `integration/testdata/test/*.yaml`.

**`internal/`:**
- Purpose: Package-private support code not intended for external import.
- Contains: `about/`, `clist/`, `rate/`, `signal/`, `toplist/`, `version/`.
- Key files: `internal/signal/`, `internal/rate/`, `internal/version/version.go`.

## Key File Locations

**Entry Points:**
- `cmd/gofs/main.go`: Binary main function.
- `cmd/gofs.go`: Primary bootstrap and runtime orchestration API.

**Configuration:**
- `conf/config.go`: Central config schema.
- `conf/parser.go`: File/content parsing into `conf.Config`.
- `flag/flag.go`: CLI flag parsing front door.
- `go.mod`: Module and dependency manifest.

**Core Logic:**
- `monitor/monitor.go`: Monitor factory.
- `monitor/base_monitor.go`: Shared batching, delay, and cron logic.
- `sync/sync.go`: Sync factory.
- `sync/disk_sync.go`: Local disk copy engine.
- `sync/remote_server_sync.go`: Remote server broadcaster plus local sync.
- `sync/remote_client_sync.go`: Remote pull consumer.

**Network Surfaces:**
- `server/httpfs/file_server.go`: HTTP server construction and route wiring.
- `server/handler/push_handler.go`: Push upload endpoint.
- `api/apiserver/grpc_server.go`: gRPC server startup and service registration.
- `api/apiclient/grpc_client.go`: gRPC client connection and login.

**Testing:**
- `integration/`: End-to-end scenarios.
- `*_test.go` across packages: Package-level unit tests.

## Module Organization

**Factory-led instantiation:**
- Use `sync.NewSync` in `sync/sync.go` to choose transport-specific syncers.
- Use `monitor.NewMonitor` in `monitor/monitor.go` to choose monitoring strategy.
- Use `server.NewServerOption`, `sync.NewSyncOption`, and `monitor.NewMonitorOption` to pass assembled dependencies.

**Transport partitioning:**
- Local disk logic stays in `sync/disk_sync.go`.
- Remote disk server/client split across `sync/remote_server_sync.go` and `sync/remote_client_sync.go`.
- SFTP and MinIO variants are isolated in dedicated `sync/sftp_*`, `sync/minio_*`, and `driver/*` files.

**API partitioning:**
- HTTP endpoints live under `server/handler/` and `server/middleware/`.
- gRPC contracts and service implementations live under `api/` by concern: `auth`, `info`, `monitor`, `task`.

## Naming Conventions

**Files:**
- Package-level snake_case filenames dominate implementation files, often pairing transport and role, such as `remote_client_sync.go`, `task_client_monitor.go`, and `push_handler.go`.

**Directories:**
- Top-level directories map to domain areas: `monitor/`, `sync/`, `server/`, `api/`, `driver/`, `report/`.
- Nested API/service directories use concern names directly, such as `api/monitor/` and `server/middleware/`.

## Notable Hotspots

**High-complexity orchestration:**
- `cmd/gofs.go` (364 lines): Central startup path, many mode switches, and lifecycle handling.
- `monitor/base_monitor.go` (320 lines): Shared queueing, batching, cron, and worker coordination.

**Heavy transfer logic:**
- `sync/push_client_sync.go` (504 lines): Push-upload implementation and comparison logic.
- `sync/disk_sync.go` (478 lines): Local file comparison, copy, encryption, and metadata handling.
- `sync/remote_client_sync.go` (458 lines): Remote metadata fetch, auth, and ranged file download logic.

**Protocol surface complexity:**
- `server/handler/push_handler.go` (336 lines): Multi-action upload endpoint and chunk handling.
- `driver/minio/minio.go` (355 lines) and `driver/sftp/sftp.go` (351 lines): Backend-specific filesystem behavior.

**Generated code area:**
- `api/*/*.pb.go` and `api/*/*_grpc.pb.go`: Generated protobuf and gRPC bindings; large but generated.

## How Code Is Partitioned

**By runtime responsibility:**
- Bootstrapping stays in `cmd/`.
- Domain-neutral config and types stay in `conf/` and `core/`.
- Change detection stays in `monitor/`.
- Mutation and transfer stay in `sync/` and `driver/`.
- External interfaces stay in `server/` and `api/`.

**By transport/backend:**
- Disk, remote disk, SFTP, and MinIO implementations are separated into dedicated files rather than hidden behind one monolithic package file.

**By test scope:**
- Unit tests live beside packages using `*_test.go`.
- Scenario tests live under `integration/` with reusable YAML fixtures under `integration/testdata/`.

## Where to Add New Code

**New sync backend:**
- Primary code: `sync/` for orchestration and `driver/` if the backend needs a reusable filesystem abstraction.
- Factory wiring: `sync/sync.go` and possibly `monitor/monitor.go` if the source side needs custom monitoring.
- Tests: Add package tests beside the new files and integration fixtures under `integration/testdata/` when end-to-end behavior matters.

**New monitor mode:**
- Implementation: `monitor/` beside existing monitor strategies.
- Factory wiring: `monitor/monitor.go`.

**New HTTP endpoint:**
- Handler: `server/handler/`.
- Middleware: `server/middleware/` if cross-cutting.
- Route registration: `server/httpfs/file_server.go`.

**New gRPC service:**
- Contract: `api/proto/` then generated code in a matching `api/<service>/` package.
- Server/client wiring: `api/apiserver/grpc_server.go` and `api/apiclient/grpc_client.go`.

**Shared utilities:**
- Reusable external-facing helpers: package-specific directories such as `core/`, `fs/`, `retry/`, or `report/`.
- Internal-only helpers: `internal/`.

## Special Directories

**`.planning/codebase/`:**
- Purpose: Generated codebase maps for GSD planning and execution.
- Generated: Yes.
- Committed: Project-dependent; current repository contains `.planning/`.

**`api/proto/`:**
- Purpose: Protobuf source definitions for internal gRPC services.
- Generated: No.
- Committed: Yes.

**`api/*/*.pb.go`:**
- Purpose: Generated protobuf and gRPC bindings.
- Generated: Yes.
- Committed: Yes.

**`server/template/`:**
- Purpose: HTML templates embedded by `server/resource.go`.
- Generated: No.
- Committed: Yes.

**`integration/testdata/`:**
- Purpose: Scenario configs for integration tests.
- Generated: No.
- Committed: Yes.

**`scripts/`:**
- Purpose: Build, release, docker, env, and validation helper scripts.
- Generated: No.
- Committed: Yes.

**`relay/`:**
- Purpose: Not applicable in current codebase state; directory exists but no Go implementation files were detected during this architecture pass.
- Generated: No.
- Committed: Yes.

---

*Structure analysis: 2026-04-23*
