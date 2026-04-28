# External Integrations

**Analysis Date:** 2026-04-23

## APIs & External Services

**Remote file sync services:**
- Remote Disk gRPC service - distributes file metadata, auth, monitor streams, and task subscription for remote sync workflows.
  - SDK/Client: `google.golang.org/grpc`, `golang.org/x/oauth2`, `google.golang.org/grpc/credentials/oauth`
  - Server implementation: `api/apiserver/grpc_server.go`
  - Client implementation: `api/apiclient/grpc_client.go`
  - Related sync entrypoint: `sync/remote_server_sync.go`

- HTTP/HTTPS file server - exposes login, browse, query, push, manage, report, and pprof endpoints.
  - Framework: `github.com/gin-gonic/gin`
  - Implementation: `server/httpfs/file_server.go`
  - Route reference: `server/README.md`

- SFTP servers - external remote filesystem target/source for pull and push sync.
  - SDK/Client: `github.com/pkg/sftp`, `golang.org/x/crypto/ssh`
  - Implementation: `driver/sftp/sftp.go`
  - User-facing examples: `README.md`

- MinIO / S3-compatible object storage - external object storage target/source for pull and push sync.
  - SDK/Client: `github.com/minio/minio-go/v7`
  - Implementation: `driver/minio/minio.go`
  - User-facing examples: `README.md`

**Task/config backends:**
- Cache-backed task stores - task definitions can be persisted outside process memory.
  - Supported connection types: file, memory, Redis, BuntDB, etcd
  - Loader entrypoint: `api/task/loader/loader.go`
  - Cache-backed implementation: `api/task/loader/cache.go`
  - Backing cache client: `github.com/no-src/nscache/all`

## Internal Boundaries

**CLI to sync orchestration:**
- `cmd/gofs/main.go` starts the application and delegates to `github.com/no-src/gofs/cmd`.
- `flag/flag.go` and `conf/config.go` define the runtime contract for all transport and integration settings.

**Sync engine to transport drivers:**
- Sync packages under `sync/` select and invoke local disk, remote gRPC, SFTP, and MinIO drivers.
- SFTP boundary lives in `driver/sftp/sftp.go`.
- MinIO boundary lives in `driver/minio/minio.go`.

**Server/API split:**
- gRPC service boundary is implemented in `api/apiserver/grpc_server.go`.
- Browser/file API boundary is implemented in `server/httpfs/file_server.go` with documented routes in `server/README.md`.

## Data Storage

**Databases:**
- Redis - optional session backend and optional task/config backend.
  - Connection: CLI flag `-session_connection` in `flag/flag.go` for server sessions; task loaders accept `redis://...` via `-task_conf` documented by `api/task/loader/loader.go`.
  - Client: `github.com/gin-contrib/sessions/redis`, `github.com/no-src/nscache`
  - Implementation: `server/session.go`, `api/task/loader/cache.go`

- etcd - optional task/config backend only.
  - Connection: `etcd://...` via task loader contract in `api/task/loader/loader.go`
  - Client: indirect through `github.com/no-src/nscache/all` and related dependencies declared in `go.mod`

- BuntDB - optional embedded/local task/config backend only.
  - Connection: `buntdb://...` via `api/task/loader/loader.go`
  - Client: indirect through `github.com/no-src/nscache/all`

**File Storage:**
- Local filesystem is the default source/destination for sync operations, described in `README.md` and implemented across `sync/disk_sync.go` and related packages.
- SFTP servers act as remote file storage through `driver/sftp/sftp.go`.
- MinIO buckets act as object/file storage through `driver/minio/minio.go`.

**Caching:**
- No dedicated application cache layer is detected for business data.
- `github.com/no-src/nscache` is used as a persistence abstraction for task config and Redis-backed session secret storage in `api/task/loader/cache.go` and `server/session.go`.

## Authentication & Identity

**Auth Provider:**
- Custom auth - user/password credentials are supplied by CLI/config and transformed into session or token-based access.
  - gRPC auth implementation: token issuance/login service in `api/apiserver/grpc_server.go` and token use in `api/apiclient/grpc_client.go`
  - HTTP auth implementation: Gin session middleware and login flow in `server/session.go` and `server/httpfs/file_server.go`
  - User/permission model: `auth/user.go`, `auth/perm.go`, `server/middleware/auth.go`

**Transport security:**
- TLS is supported for gRPC and HTTP file server using cert/key paths from `flag/flag.go`, wired in `api/apiserver/grpc_server.go` and `server/httpfs/file_server.go`.
- HTTP/3 is available only when TLS is enabled in `server/httpfs/file_server.go`.
- SFTP host verification can be enforced with a host key file, or bypassed when absent, in `driver/sftp/sftp.go`.

## Payments

- Not detected.

## Analytics

- Product analytics integrations are not detected.
- Internal operational reporting is available through repository packages under `report/` and the manage report endpoint exposed by `server/httpfs/file_server.go` and documented in `server/README.md`.

## Monitoring & Observability

**Error Tracking:**
- External error tracking service is not detected.

**Logs:**
- Application logging is internal and configurable through flags in `flag/flag.go` and packages under `logger/`.
- Gin request logs are routed through the project logger in `server/httpfs/file_server.go`.

**Profiling / diagnostics:**
- pprof endpoints are exposed under manage routes via `github.com/gin-contrib/pprof` in `server/httpfs/file_server.go`.

## CI/CD & Deployment

**Hosting:**
- Standalone binary and Docker container deployment are supported.
- Docker packaging is defined in `Dockerfile`; image naming/publish target is `nosrc/gofs` in `scripts/build-docker.sh` and `README.md`.

**CI Pipeline:**
- GitHub Actions is the only CI/CD service detected.
  - Build/test matrix: `.github/workflows/go.yml`
  - Docker validation: `.github/workflows/docker.yml`
  - Release build: `.github/workflows/release.yml`
  - CodeQL scan: `.github/workflows/codeql.yml`
  - Govulncheck scan: `.github/workflows/govulncheck.yml`

**Coverage / quality services:**
- Codecov upload is configured in `.github/workflows/go.yml`.
- GitHub CodeQL is configured in `.github/workflows/codeql.yml`.

## Environment Configuration

**Required env vars:**
- `GIN_MODE` is the only directly read environment variable detected, via `server/httpfs/file_server.go`.
- No repository-defined `.env` file was detected in the project root.

**Secrets location:**
- Secrets are passed primarily as CLI/config values and external files rather than checked-in env files.
- TLS cert/key file paths are configured by `-tls_cert_file` and `-tls_key_file` in `flag/flag.go`.
- Token secret is configured by `-token_secret` in `flag/flag.go`.
- SFTP private key and known-hosts file paths are part of SSH config consumed by `driver/sftp/sftp.go`.
- Redis session secrets may be provided in the Redis connection query string or persisted in Redis by `server/session.go`.

## Data Flow Touchpoints

**Local disk → gRPC monitor → HTTP file fetch → destination write:**
1. Local source changes are observed and packaged by sync logic in `sync/remote_server_sync.go`.
2. The gRPC server in `api/apiserver/grpc_server.go` broadcasts monitor messages.
3. Clients connect through `api/apiclient/grpc_client.go` and use file API/base URLs advertised by the server.
4. File content is served by `server/httpfs/file_server.go` and written into the destination by sync/driver code.

**Local disk ↔ SFTP:**
1. Sync orchestration in `sync/sftp_pull_client_sync.go` or `sync/sftp_push_client_sync.go` selects the SFTP driver.
2. `driver/sftp/sftp.go` reads/writes remote files over SSH-authenticated SFTP.

**Local disk ↔ MinIO:**
1. Sync orchestration in `sync/minio_pull_client_sync.go` or `sync/minio_push_client_sync.go` selects the MinIO driver.
2. `driver/minio/minio.go` reads/writes objects in a configured bucket using static credentials.

**Task distribution:**
1. Task service is registered in `api/apiserver/grpc_server.go`.
2. Task configuration is loaded from file/memory/cache backends via `api/task/loader/loader.go`.
3. Task clients subscribe through `api/apiclient/grpc_client.go`.

## Webhooks & Callbacks

**Incoming:**
- Not detected. No webhook-specific inbound HTTP callbacks are defined.

**Outgoing:**
- Not detected. External webhook POST callbacks are not implemented.

## Reverse Proxy / Relay Integrations

- Optional relay/reverse-proxy usage is documented, not embedded as a code dependency.
- `relay/README.md` describes deployment patterns using frp and ngrok in front of the file server / remote sync endpoints.

## Absent Integrations

- Payment gateways: not detected.
- Third-party OAuth identity providers: not detected.
- Hosted analytics SDKs: not detected.
- Managed cloud deployment descriptors such as Kubernetes manifests or Terraform: not detected in the repository root.
- Message queues such as Kafka, RabbitMQ, or NATS: not detected.

---

*Integration audit: 2026-04-23*
