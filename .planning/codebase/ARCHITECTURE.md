# Architecture

**Analysis Date:** 2026-04-23

## Pattern Overview

**Overall:** Interface-oriented file synchronization engine with mode-selected transports and server adapters.

**Key Characteristics:**
- `cmd/gofs/main.go` and `cmd/gofs.go` route all execution through a single config-driven bootstrap path.
- `monitor/` detects change events while `sync/` owns file transfer and mutation logic.
- `server/httpfs/file_server.go` and `api/apiserver/grpc_server.go` expose HTTP and gRPC surfaces around the same sync core.

## Layers

**CLI and bootstrap:**
- Purpose: Parse config, initialize logging, choose one-shot vs long-running execution, and wire subsystems.
- Location: `cmd/gofs/main.go`, `cmd/gofs.go`, `cmd/logger.go`, `flag/flag.go`, `conf/config.go`, `conf/parser.go`
- Contains: Program entrypoint, config loading, daemon setup, logger setup, subsystem construction.
- Depends on: `conf/`, `auth/`, `daemon/`, `server/`, `monitor/`, `sync/`, `report/`, `retry/`, `ignore/`
- Used by: End-user CLI invocation and integration tests via `cmd.RunWithConfig*`.

**Configuration and type system:**
- Purpose: Model runtime options and virtual file system endpoints.
- Location: `conf/config.go`, `core/vfs.go`, `core/vfs_type.go`, `core/flagset.go`, `core/path.go`, `core/size.go`, `core/duration.go`
- Contains: `conf.Config`, `core.VFS`, typed flag encoders/decoders, path and size helpers.
- Depends on: Standard library plus `github.com/no-src/nsgo/yamlutil`.
- Used by: All higher layers; `sync.NewSync` and `monitor.NewMonitor` branch on `core.VFS` capabilities.

**Sync execution layer:**
- Purpose: Apply file operations between source and destination backends.
- Location: `sync/sync.go`, `sync/disk_sync.go`, `sync/remote_server_sync.go`, `sync/remote_client_sync.go`, `sync/push_client_sync.go`, `sync/sftp_*`, `sync/minio_*`, `sync/driver_*`
- Contains: `Sync` interface plus concrete strategies for disk, remote disk, push server, SFTP, MinIO, and generic drivers.
- Depends on: `core/`, `encrypt/`, `driver/`, `ignore/`, `retry/`, `report/`, `server/`, `api/`
- Used by: `monitor/` implementations and one-shot sync flows in `cmd/gofs.go`.

**Monitoring and orchestration layer:**
- Purpose: Observe source changes or subscribe to remote events, then dispatch sync actions.
- Location: `monitor/monitor.go`, `monitor/base_monitor.go`, `monitor/fsnotify_monitor.go`, `monitor/remote_client_monitor.go`, `monitor/task_client_monitor.go`, `monitor/sftp_pull_client_monitor.go`, `monitor/minio_pull_client_monitor.go`
- Contains: `Monitor` interface, fsnotify loop, remote gRPC consumer, delayed write batching, cron-triggered sync.
- Depends on: `sync/`, `retry/`, `ignore/`, `eventlog/`, `report/`, `api/apiclient`
- Used by: `cmd/gofs.go` long-running execution path.

**HTTP server layer:**
- Purpose: Serve source and destination files, push API, login, config, profiling, and report endpoints.
- Location: `server/httpfs/file_server.go`, `server/handler/*.go`, `server/middleware/*.go`, `server/session.go`, `server/resource.go`
- Contains: Gin engine setup, route registration, auth middleware, embedded templates, session storage.
- Depends on: `auth/`, `core/`, `driver/minio`, `driver/sftp`, `report/`, `server/`
- Used by: Browser access, remote pull clients, push clients, and management endpoints.

**gRPC API layer:**
- Purpose: Expose remote monitor streaming, login, task subscription, and server metadata.
- Location: `api/apiserver/*.go`, `api/apiclient/*.go`, `api/info/*.go`, `api/monitor/*.go`, `api/auth/*.go`, `api/task/*.go`
- Contains: gRPC server/client wrappers, auth interceptors, token login, monitor stream fan-out, task dispatch.
- Depends on: `auth/`, `report/`, generated protobuf files in `api/*/*.pb.go`
- Used by: `sync/remote_server_sync.go`, `monitor/remote_client_monitor.go`, `monitor/task_client_monitor.go`.

**Storage driver abstraction:**
- Purpose: Normalize remote filesystem access for non-disk targets.
- Location: `driver/driver.go`, `driver/sftp/*.go`, `driver/minio/*.go`
- Contains: `driver.Driver` interface and concrete remote filesystem adapters.
- Depends on: Backend SDKs and `net/http` / `io/fs` contracts.
- Used by: `sync/driver_*`, `server/httpfs/file_server.go` for mounted HTTP file serving.

**Operational support:**
- Purpose: Logging, retry, reporting, result signaling, event logging, and background execution.
- Location: `logger/logger.go`, `retry/`, `report/`, `result/result.go`, `eventlog/`, `daemon/daemon.go`, `wait/wait.go`
- Contains: Shared infrastructure components used across runtime modes.
- Depends on: Internal support packages in `internal/` and `github.com/no-src/log`.
- Used by: All runtime subsystems.

## Data Flow

**CLI startup flow:**

1. `cmd/gofs/main.go` calls `cmd.Run().Wait()`.
2. `cmd/gofs.go` parses flags or config content into `conf.Config` and resolves defaults in `initDefaultValue`.
3. `cmd/gofs.go` initializes loggers, retry policy, reporter, ignore rules, syncer, and monitor.
4. Optional HTTP server startup runs before monitor start via `startWebServer`.
5. `monitor.Monitor.Start()` returns a `wait.Wait` used as the process lifecycle handle.

**Local disk sync flow:**

1. `monitor/fsnotify_monitor.go` watches the source directory tree with `fsnotify`.
2. Events are filtered through `ignore.PathIgnore` and normalized into create/write/remove/rename/chmod actions.
3. `monitor/base_monitor.go` batches write-heavy paths and optionally spreads large writes across worker slots.
4. `sync/disk_sync.go` compares metadata and chunk hashes, then copies only required data into destination paths.
5. `eventlog/` and `report/` record processed events.

**Remote disk server → client flow:**

1. `sync/remote_server_sync.go` wraps `diskSync` and starts `api/apiserver/grpc_server.go`.
2. Local source changes still pass through monitor and sync interfaces, but remote server sync additionally emits `monitor.MonitorMessage` payloads.
3. `api/monitor/monitor.go` pushes each message to connected monitor streams.
4. `monitor/remote_client_monitor.go` receives streamed events through `api/apiclient/grpc_client.go`.
5. `sync/remote_client_sync.go` resolves file metadata through gRPC info plus HTTP file/query endpoints, then writes to local destination.

**Push-to-server flow:**

1. `monitor/fsnotify_monitor.go` detects local changes on the client.
2. `sync/push_client_sync.go` packages action metadata and file chunks for HTTP push.
3. `server/handler/push_handler.go` validates action payloads and mutates server-side storage paths.
4. Server-side writes update filesystem state under the configured source path.

**Task mode flow:**

1. `api/task/task.go` serves `SubscribeTask` over gRPC.
2. `api/task/dispatcher.go` loads task definitions from task config files.
3. `monitor/task_client_monitor.go` acquires tasks, then launches workers using returned config content.
4. Each worker re-enters the standard `cmd.RunWithConfigContent` pipeline.

**State Management:**
- Runtime configuration is centralized in `conf.Config` and copied into option structs in `sync/option.go`, `monitor/option.go`, and `server/option.go`.
- File ownership belongs to concrete sync implementations: `sync/disk_sync.go` owns local path mutation, `sync/remote_client_sync.go` owns remote-pull materialization, and `server/handler/push_handler.go` owns push-write application.
- Connection and operational state live in memory: `report/reporter.go` stores stats, `api/monitor/monitor.go` stores connected monitor channels in `sync.Map`, and `monitor/base_monitor.go` stores queued write state.

## Key Abstractions

**Virtual filesystem endpoint (`core.VFS`):**
- Purpose: Represent disk, remote disk, SFTP, and MinIO endpoints behind one config shape.
- Examples: `core/vfs.go`, `conf/config.go`, `sync/sync.go`, `monitor/monitor.go`
- Pattern: Capability checks such as `IsDisk()`, `Is(core.RemoteDisk)`, `Server()` select concrete runtime behavior.

**Sync strategy (`sync.Sync`):**
- Purpose: Abstract file mutation operations across transports.
- Examples: `sync/sync.go`, `sync/disk_sync.go`, `sync/remote_server_sync.go`, `sync/remote_client_sync.go`
- Pattern: Factory dispatch in `sync.NewSync` chooses one implementation per source/destination pair.

**Monitor strategy (`monitor.Monitor`):**
- Purpose: Abstract event observation and lifecycle management.
- Examples: `monitor/monitor.go`, `monitor/fsnotify_monitor.go`, `monitor/remote_client_monitor.go`
- Pattern: Factory dispatch in `monitor.NewMonitor` selects by source VFS mode and task-client flags.

**Option bag construction:**
- Purpose: Preserve a thin dependency injection boundary around concrete components.
- Examples: `sync/option.go`, `monitor/option.go`, `server/option.go`
- Pattern: Copy `conf.Config` plus runtime services into immutable-ish option structs passed to constructors.

**Lifecycle signaling:**
- Purpose: Coordinate async startup, shutdown, and error propagation.
- Examples: `result/result.go`, `wait/wait.go`, `cmd/gofs.go`
- Pattern: Public entrypoints return wait/result handles rather than blocking directly.

## Entry Points

**CLI binary:**
- Location: `cmd/gofs/main.go`
- Triggers: `go run`, installed `gofs` binary, Docker image entry.
- Responsibilities: Invoke `cmd.Run()` and map failure to process exit code.

**Program bootstrap API:**
- Location: `cmd/gofs.go`
- Triggers: CLI path, tests, task workers via `RunWithConfigContent`.
- Responsibilities: Parse config, initialize runtime services, optionally execute one-shot commands, start servers and monitors.

**HTTP file server:**
- Location: `server/httpfs/file_server.go`
- Triggers: `conf.Config.EnableFileServer` in `cmd/gofs.go`.
- Responsibilities: Serve static files, query API, push API, login, report, and manage endpoints.

**gRPC server:**
- Location: `api/apiserver/grpc_server.go`
- Triggers: `sync/remote_server_sync.go` when source is remote-disk server mode.
- Responsibilities: Login, info, monitor stream fan-out, and task subscription.

**Task worker re-entry:**
- Location: `monitor/task_client_monitor.go`, callback passed from `cmd/gofs.go` to `monitor.NewMonitor`
- Triggers: Task subscription in task-client mode.
- Responsibilities: Start nested sync jobs from remote task payloads.

## Error Handling

**Strategy:** Constructor-first validation with asynchronous runtime error propagation through `wait.Wait` and `result.Result`.

**Patterns:**
- `cmd/gofs.go` accumulates fatal setup errors and passes them to `result.InitDoneWithError` / `DoneWithError`.
- `monitor/base_monitor.go` and monitor implementations prefer logging plus retry wrappers for transient file and connection errors.
- `api/apiserver/interceptor.go` converts auth failures into gRPC `codes.Unauthenticated`.
- `server/handler/push_handler.go` converts handler failures into structured API results instead of raw HTTP error bodies.
- Factories such as `sync.NewSync` and `monitor.NewMonitor` reject unsupported filesystem combinations immediately.

## Cross-Cutting Concerns

**Logging:**
- `cmd/logger.go` creates console, file, web, and event loggers.
- Concrete components receive `*logger.Logger` through option structs and log at subsystem boundaries.

**Validation:**
- Input normalization happens primarily during config parsing and constructor checks in `cmd/gofs.go`, `sync/*.go`, and `monitor/*.go`.
- gRPC auth validation is centralized in `api/apiserver/interceptor.go`.

**Authentication:**
- HTTP auth uses session middleware in `server/httpfs/file_server.go` plus route guards in `server/middleware/auth.go`.
- gRPC auth uses token login in `api/auth/` and interceptors in `api/apiserver/interceptor.go`.
- Anonymous fallback is explicit in `api/apiserver/grpc_server.go` and `api/apiclient/grpc_client.go` when no users are configured.

**Observability:**
- `report/reporter.go` aggregates API stats, active monitor connections, and file events.
- `server/httpfs/file_server.go` exposes pprof and report endpoints under `/manage` when enabled.

**Embedded resources:**
- `server/resource.go` embeds HTML templates from `server/template/` into the binary.

---

*Architecture analysis: 2026-04-23*
