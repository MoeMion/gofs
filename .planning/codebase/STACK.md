# Technology Stack

**Analysis Date:** 2026-04-23

## Languages

**Primary:**
- Go 1.24.4 module target in `go.mod`; application entrypoint is `cmd/gofs/main.go` and the repository implementation is primarily `*.go` packages under `action/`, `api/`, `driver/`, `server/`, `sync/`, and related directories.

**Secondary:**
- YAML for CI and config serialization in `.github/workflows/*.yml` and `conf/config.go`.
- Bash for local/dev automation in `scripts/build-docker.sh`, `scripts/init-env.sh`, `scripts/govulncheck.sh`, and related scripts.
- Protocol Buffers generated Go stubs in `api/auth/*.pb.go`, `api/info/*.pb.go`, `api/monitor/*.pb.go`, and `api/task/*.pb.go`.

## Runtime

**Environment:**
- Go CLI/server runtime; repository README requires Go 1.23+ in `README.md`, while `go.mod` pins the module to Go 1.24.4.
- Cross-platform targets are exercised in CI on Ubuntu, Windows, and macOS in `.github/workflows/go.yml`.

**Package Manager:**
- Go modules via `go.mod` and `go.sum`.
- Lockfile: present as `go.sum`.

## Frameworks

**Core:**
- Standard library networking and filesystem APIs power the main runtime in packages like `sync/remote_client_sync.go`, `driver/sftp/sftp.go`, and `driver/minio/minio.go`.
- Gin `github.com/gin-gonic/gin` provides the HTTP file server and manage/login endpoints in `server/httpfs/file_server.go`.
- gRPC `google.golang.org/grpc` provides internal RPC services and clients in `api/apiserver/grpc_server.go` and `api/apiclient/grpc_client.go`.

**Testing:**
- Go `testing` package drives unit and integration tests across the repository, including tagged integration suites in `.github/workflows/go.yml`.

**Build/Dev:**
- Docker multi-stage build in `Dockerfile` compiles the binary in `golang:latest` and packages it into `alpine:latest`.
- Protocol buffer toolchain is documented in `api/README.md` with `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc`.
- GitHub Actions orchestrates build, test, release, Docker, CodeQL, and vulnerability scans in `.github/workflows/go.yml`, `.github/workflows/docker.yml`, `.github/workflows/release.yml`, `.github/workflows/codeql.yml`, and `.github/workflows/govulncheck.yml`.

## Key Dependencies

**Critical:**
- `github.com/gin-gonic/gin` v1.10.0 - HTTP server/router for file browser, login flow, manage APIs, and file APIs in `server/httpfs/file_server.go`.
- `google.golang.org/grpc` v1.75.0 - RPC transport for auth, info, monitor, and task services in `api/apiserver/grpc_server.go` and `api/apiclient/grpc_client.go`.
- `github.com/pkg/sftp` v1.13.9 - SFTP client transport for remote filesystem sync in `driver/sftp/sftp.go`.
- `github.com/minio/minio-go/v7` v7.0.94 - S3-compatible object storage client for MinIO sync in `driver/minio/minio.go`.
- `golang.org/x/crypto` v0.41.0 - SSH and cryptographic primitives used by the SFTP driver in `driver/sftp/sftp.go`.
- `github.com/fsnotify/fsnotify` v1.8.0 - filesystem change monitoring declared in `go.mod` for real-time sync workflows described in `README.md`.

**Infrastructure:**
- `github.com/gin-contrib/sessions` v1.0.2 with memory and Redis stores supports file-server session state in `server/session.go`.
- `github.com/gin-contrib/gzip` v1.2.3 adds HTTP response compression in `server/httpfs/file_server.go`.
- `github.com/gin-contrib/pprof` v1.5.2 exposes profiling endpoints under manage routes in `server/httpfs/file_server.go`.
- `github.com/quic-go/quic-go` v0.57.0 enables HTTP/3 server mode in `server/httpfs/file_server.go`; UDP buffer tuning is scripted in `scripts/init-env.sh`.
- `golang.org/x/oauth2` v0.30.0 and `google.golang.org/grpc/credentials/oauth` support bearer-token style gRPC credentials in `api/apiclient/grpc_client.go`.
- `github.com/no-src/nscache` v0.1.3 provides pluggable cache-backed storage for task config loaders and Redis-backed session secret storage in `api/task/loader/cache.go` and `server/session.go`.
- `github.com/robfig/cron/v3` v3.0.1 supports scheduled sync mode referenced in `README.md` and configured through `flag/flag.go`.

## Storage

**Local filesystem:**
- Local disk sync is a first-class mode described in `README.md` and handled through sync/server packages such as `sync/disk_sync.go` and `server/httpfs/file_server.go`.

**Remote/object storage:**
- SFTP storage is accessed through `driver/sftp/sftp.go`.
- MinIO/S3-compatible storage is accessed through `driver/minio/minio.go`.
- Task configuration persistence can use file, memory, Redis, BuntDB, or etcd loaders via `api/task/loader/loader.go` and `api/task/loader/cache.go`.
- Session storage supports in-memory and Redis in `server/session.go`.

## Configuration

**Environment:**
- Minimal environment-variable usage is detected; Gin mode reads `GIN_MODE` via `os.Getenv(gin.EnvGinMode)` in `server/httpfs/file_server.go`.
- Runtime configuration is primarily CLI-flag and YAML/JSON struct based through `flag/flag.go` and `conf/config.go`.
- TLS certificate and key paths are configured by flags `-tls_cert_file` and `-tls_key_file` in `flag/flag.go`; actual secret files are not stored in the repository.

**Build:**
- Module metadata is in `go.mod` and `go.sum`.
- Container build is defined by `Dockerfile` and `scripts/build-docker.sh`.
- CI/build/test/release/security workflows live in `.github/workflows/`.
- Protobuf regeneration instructions are documented in `api/README.md`.

## Infra / Build / Dev Tooling

**Containerization:**
- Docker image build and packaging use `Dockerfile` and `scripts/build-docker.sh`.
- Published image target is `nosrc/gofs` per `scripts/build-docker.sh` and `README.md`.

**CI/CD:**
- Main build/test matrix: `.github/workflows/go.yml`.
- Docker build validation: `.github/workflows/docker.yml`.
- Release artifact build: `.github/workflows/release.yml`.
- Vulnerability scan: `.github/workflows/govulncheck.yml`.
- Static security analysis: `.github/workflows/codeql.yml`.
- Coverage upload to Codecov: `.github/workflows/go.yml`.

**Developer scripts:**
- `scripts/init-env.sh` tunes kernel UDP socket buffers for QUIC/HTTP3 testing.
- `scripts/build-docker.sh` builds and smoke-tests the Docker image.
- `scripts/test-coverage.sh` and `scripts/govulncheck.sh` support local verification flows.

## Notable Technical Choices

- The repository is a single Go module with package-oriented boundaries rather than a multi-module workspace, centered on `go.mod`.
- Transport support spans plain local disk, gRPC remote disk sync, HTTP/HTTPS file serving, optional HTTP/3 over QUIC, SFTP, and MinIO object storage, with representative code in `sync/remote_server_sync.go`, `server/httpfs/file_server.go`, `driver/sftp/sftp.go`, and `driver/minio/minio.go`.
- Authentication is built around user/password login and token issuance for gRPC in `api/apiserver/grpc_server.go` and `api/apiclient/grpc_client.go`, while the Gin file server uses session middleware in `server/session.go`.
- Configuration favors explicit flags over environment-driven configuration, with nearly all operational settings declared in `flag/flag.go` and stored in `conf/config.go`.
- Generated protobuf code is committed to the repository under `api/*/*.pb.go`, with regeneration documented in `api/README.md`.

## Platform Requirements

**Development:**
- Go toolchain 1.23+ per `README.md`; module file targets Go 1.24.4 in `go.mod`.
- Docker is required for image builds and Docker workflow validation in `Dockerfile` and `scripts/build-docker.sh`.
- `protoc` plus Go gRPC plugins are required only when regenerating API stubs per `api/README.md`.

**Production:**
- Standalone compiled binary built from `cmd/gofs/main.go`.
- Optional container deployment via the multi-stage `Dockerfile`.
- Network-facing deployments can require TLS cert/key files, optional Redis for sessions, and reachable SFTP/MinIO/etcd/Redis endpoints depending on enabled sync/task modes.

---

*Stack analysis: 2026-04-23*
