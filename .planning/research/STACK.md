# Technology Stack — v2.0 FTP Sync Library

**Project:** gofs  
**Scope:** Stack changes to convert shipped FTP sync support into a focused Go package centered on `FTPSyncService`  
**Researched:** 2026-04-27  
**Confidence:** HIGH for dependency removals/API direction; MEDIUM for final public package naming until implementation design is accepted.

## Bottom Line

Do **not** add a new framework. v2.0 should mostly be a dependency and surface-area reduction:

- Keep Go modules, standard library filesystem/network primitives, `github.com/jlaffaye/ftp`, and `github.com/fsnotify/fsnotify`.
- Keep only the internal pieces needed for disk↔FTP sync: FTP driver, disk sync, driver pull/push sync, ignore rules, retry/rate limiting, minimal logging hooks, and config normalization.
- Remove or isolate CLI/server/protocol infrastructure: Gin, gRPC, protobuf, SFTP, MinIO, QUIC, OAuth, Redis/cache/task loaders, cron scheduler if background monitoring can be expressed with a simple ticker/poll interval.
- Introduce a small public package API around `FTPSyncService` with explicit structs and `context.Context`-first methods.

## Recommended Core Stack

| Technology | Keep/Add/Remove | Purpose in v2.0 | Rationale |
|------------|-----------------|-----------------|-----------|
| Go `1.24.4` | Keep | Library implementation and tests | Already pinned in `go.mod`; no runtime framework needed. |
| Go standard library `context` | Add to public API contract | Cancellation/deadlines for one-shot sync and background monitor lifecycle | Official Go guidance: pass `context.Context` explicitly as first parameter; do not store it in structs. |
| Go standard library `io/fs`, `os`, `path/filepath`, `time`, `net` | Keep | Filesystem traversal, metadata, deadlines, address handling | Existing code already depends on these primitives; they are enough for the library core. |
| `github.com/jlaffaye/ftp` `v0.2.0` | Keep | Plain FTP client transport | Existing v1.0 uses it successfully; docs expose needed APIs: `Dial`, `DialWithContext`, `DialWithTimeout`, `Login`, `Retr`, `Stor`, `Walk`, `GetTime`, `SetTime`, `Rename`, `Delete`. Note: `ServerConn` is not concurrency-safe, so keep serialized access or per-worker connections. |
| `github.com/fsnotify/fsnotify` | Keep, update to latest compatible if desired | Local disk event monitoring for disk→FTP background sync | Official docs confirm cross-platform support, but recursive watches and network FS notifications are not automatic; keep existing batching/recursive-watch handling rather than inventing a watcher. |
| `github.com/no-src/log` | Prefer remove from public path; optional internal only | Internal default logging if retained | A library should not force a logging dependency or global logger. Prefer `Logger interface { Printf(string, ...any) }` or event callbacks. |
| `github.com/no-src/nsgo`, `github.com/no-src/fsctl`, `golang.org/x/time` | Review/keep only if directly used by retained sync/rate code | Helpers/rate limiting | Keep only if `go mod tidy` after extraction proves they are still needed. Do not expose them in public API. |

## Dependencies to Remove or Isolate

| Dependency / Area | Recommendation | Why |
|-------------------|----------------|-----|
| `github.com/gin-gonic/gin`, `gin-contrib/*` | Remove | HTTP file server/session/web UI are out of scope for package invocation. |
| `google.golang.org/grpc`, `google.golang.org/protobuf`, `golang.org/x/oauth2` | Remove | Remote-disk gRPC, auth, tasks, generated protobufs are not part of FTP sync library. |
| `github.com/pkg/sftp`, `golang.org/x/crypto`, `github.com/kevinburke/ssh_config` | Remove | SFTP and SSH config are explicitly abandoned for v2.0. |
| `github.com/minio/minio-go/v7` and MinIO transitive deps | Remove | Non-FTP object storage support is out of scope. |
| `github.com/quic-go/quic-go` | Remove | HTTP/3 server mode is out of scope. |
| `github.com/no-src/nscache`, Redis/BuntDB/etcd transitives | Remove | Task/session/config stores are CLI/server runtime concerns. |
| `github.com/robfig/cron/v3` | Prefer remove | For a library, background polling should use `time.Ticker` plus `context.Context`; cron syntax is legacy CLI semantics, not a necessary v2 dependency. Keep only behind a compatibility adapter if preserving exact `sync_cron` strings is mandatory. |
| `github.com/schollz/progressbar/v3` | Remove | CLI progress UI must not be in a package API. Expose progress callbacks/events instead. |
| Docker image/release binary tooling | Remove or move to examples/dev only | v2.0 ships as a Go package, not a supported CLI/container runtime. |

## Public API Shape

Recommended package: `ftpsync` or root package `gofs` with only FTP sync exports. Prefer `ftpsync` if module path can change later; prefer root `gofs` if import stability matters.

```go
type FTPSyncService struct {
	// no stored context; hold immutable config and internal collaborators only
}

type Config struct {
	Source Endpoint
	Dest   Endpoint

	DeleteExtraneous bool
	IgnorePatterns   []string
	ChunkSize        int64
	MaxTransferRate  int64
	Retry            RetryConfig
	Monitor          MonitorConfig
	Logger           Logger // optional; nil = no-op
	OnEvent          func(Event)
}

type Endpoint struct {
	Kind       EndpointKind // LocalDisk or FTP
	LocalPath  string
	RemotePath string
	FTP        FTPConfig
}

type FTPConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	PassiveMode bool
	Timeout     time.Duration
	Encoding    string // keep v1 semantics: auto/utf8/gbk
}
```

Recommended methods:

```go
func NewFTPSyncService(cfg Config) (*FTPSyncService, error)
func (s *FTPSyncService) SyncOnce(ctx context.Context) error
func (s *FTPSyncService) Start(ctx context.Context) (*Run, error)

type Run struct { /* internal channels/wait group */ }
func (r *Run) Done() <-chan struct{}
func (r *Run) Err() error
func (r *Run) Stop() error // optional convenience; should cancel internal context if Start created one
```

API rules:

- `context.Context` is the first parameter for operations, not a field on `FTPSyncService`.
- `SyncOnce(ctx)` blocks until complete or canceled.
- `Start(ctx)` starts background monitoring and stops when `ctx` is canceled; do not require process signals, daemon mode, or global wait handles.
- Accept explicit structs, not CLI flag strings. Provide a compatibility parser separately only if needed: `ConfigFromLegacy(conf.Config)` or `ParseEndpointURL`.
- Support only `LocalDisk ↔ FTP`. Reject FTP↔FTP and non-FTP protocols with typed errors.

## Context and Cancellation Patterns

| Area | Pattern |
|------|---------|
| Public methods | `func (s *FTPSyncService) SyncOnce(ctx context.Context) error`; `func (s *FTPSyncService) Start(ctx context.Context) (*Run, error)`. |
| FTP dial | Pass `ftp.DialWithContext(ctx)` and `ftp.DialWithTimeout(cfg.FTP.Timeout)` to `github.com/jlaffaye/ftp`. |
| Long reads/writes | Check `ctx.Err()` before/after each file operation; where possible set deadlines or close the FTP connection on cancellation to unblock data transfers. |
| Background local monitor | Select on `ctx.Done()`, `watcher.Events`, and `watcher.Errors`; always `Close()` watcher on exit. |
| FTP source monitoring | FTP has no event stream; implement polling with `time.Ticker` and reject persistent monitoring unless `Monitor.PollInterval > 0`. |
| Retry loops | Make retry sleeps context-aware (`select { case <-time.After(d): ...; case <-ctx.Done(): return ctx.Err() }`). |
| Errors | Return `context.Canceled` / `context.DeadlineExceeded` directly when cancellation wins; wrap operational errors with `%w`. |

## Internal Package Simplification

Keep or extract:

- `driver/ftp` → core FTP transport adapter.
- `sync/disk_sync.go`, `sync/driver_pull_client_sync.go`, `sync/driver_push_client_sync.go`, `sync/ftp_*` → collapse into internal package names if possible, but preserve behavior.
- `monitor/fsnotify_monitor.go` and `monitor/ftp_pull_client_monitor.go` concepts → expose only through `FTPSyncService.Start`.
- `ignore`, `retry`, `internal/rate`, path encoding, and file metadata helpers → keep if directly used.

Remove from the library build:

- `cmd/`, `flag/`, `daemon/`, `server/`, `api/`, `auth/`, task loaders, report web endpoints, SFTP/MinIO drivers, generated protobufs, Docker entrypoint assumptions.

## What NOT to Add

| Do not add | Reason |
|------------|--------|
| FTPS/TLS config | Explicitly deferred; `jlaffaye/ftp` supports TLS options, but enabling them expands certificate/security UX. |
| FTP server support | v2.0 is a client sync library only. |
| New workflow/orchestration framework | `context`, goroutines, channels, and existing sync code are sufficient. |
| New config library | Public structs are clearer than YAML/CLI config in a Go package. YAML can live in examples, not core. |
| Global logger, global config, environment-driven behavior | Library consumers need isolated service instances. |
| Progress bars or terminal UI | Use callbacks/events; never write to stdout by default. |
| Cron dependency as default scheduler | Use polling interval/ticker; cron syntax is legacy CLI behavior. |
| Connection pooling initially | `jlaffaye/ftp` connections are not concurrency-safe; pooling adds lifecycle complexity. Start serialized, add pool only if benchmarks prove need. |

## Migration Sequence for Roadmap

1. **Define public API and config structs** — lock `FTPSyncService`, endpoint structs, typed errors, and context behavior.
2. **Extract minimal internal FTP sync core** — make one-shot disk↔FTP work without `cmd.RunWithConfig`.
3. **Add background service lifecycle** — local fsnotify for disk→FTP; ticker polling for FTP→disk; context cancellation tests.
4. **Prune dependencies and packages** — remove non-FTP/protocol/server deps and run `go mod tidy`.
5. **Compatibility and docs** — document mapping from old CLI URL/config fields to new structs; keep plain FTP only.

## Sources

- Project context: `.planning/PROJECT.md`, `.planning/MILESTONES.md`
- Existing architecture/stack/conventions: `.planning/codebase/ARCHITECTURE.md`, `.planning/codebase/STACK.md`, `.planning/codebase/CONVENTIONS.md`
- Current module dependencies: `go.mod`
- Current FTP implementation: `driver/ftp/ftp.go`, `sync/ftp_push_client_sync.go`, `sync/ftp_pull_client_sync.go`, `monitor/ftp_pull_client_monitor.go`, `core/vfs.go`
- Official docs: `https://pkg.go.dev/github.com/jlaffaye/ftp`, `https://pkg.go.dev/context`, `https://pkg.go.dev/github.com/fsnotify/fsnotify`
