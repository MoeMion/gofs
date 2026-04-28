---
phase: 08-ftp-only-package-reduction
verified: 2026-04-28T00:00:00Z
status: passed
score: 15/15 must-haves verified
overrides_applied: 0
---

# Phase 8: FTP-Only Package Reduction Verification Report

**Phase Goal:** Library consumers see a small FTP-only package surface whose build graph excludes old CLI/server/protocol runtimes while retaining the internal helpers needed for disk↔FTP sync.
**Verified:** 2026-04-28T00:00:00Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Importing and testing the library no longer requires CLI entrypoints, flag parsing, daemon/process management, HTTP file server, gRPC APIs, task runtime, auth/session, SFTP, MinIO, or Docker release surfaces. | ✓ VERIFIED | `go list ./...` outputs only `github.com/no-src/gofs/ftpsync`; `go list -deps ./...` contains no `github.com/no-src/gofs/{cmd,api,server,sync,driver/sftp,driver/minio,...}` packages and no Gin/gRPC/SFTP/MinIO/protobuf runtime modules. Deleted package-tree glob for old runtime Go files returned no files. |
| 2 | FTP driver, path encoding, retry, ignore/filtering, rate limiting, and core disk/FTP sync internals remain usable behind `FTPSyncService`. | ✓ VERIFIED | `ftpsync/internal_ftp.go` implements package-local `ftpClient` over `github.com/jlaffaye/ftp` with connect/login/binary transfer, path codec, mkdir/write/remove/walk/read/close operations; `ftpsync/internal_retry.go` implements context-aware retry; `ftpsync/oneshot.go` wires these through `executeSyncOnce`; `ftpsync/background.go` retains fsnotify-backed background disk→FTP lifecycle. `go test ./... -count=1` passed. |
| 3 | Public `ftpsync` APIs do not expose old `conf.Config`, `core.VFS`, CLI URL parsing, or server/task types. | ✓ VERIFIED | `ftpsync/options.go` exposes typed `Options`, `Endpoint`, `FTPOptions`, `RetryOptions`, `IgnoreRule`, and hooks only. `ftpsync/service.go` public methods are `NewFTPSyncService(Options)`, `SyncOnce(context.Context)`, and `StartBackground(context.Context)`. `TestPublicAPIStaysTypedOptionsOnly` reflection/doc scan exists and passed as part of full suite. |
| 4 | `go.mod` no longer retains unnecessary Gin, gRPC, SFTP, MinIO, Redis/cache, QUIC, OAuth, or protobuf dependencies for the library build. | ✓ VERIFIED | `go.mod` requires only `github.com/fsnotify/fsnotify`, `github.com/jlaffaye/ftp`, `golang.org/x/text`, and minimal indirect test/sys deps. Grep scan for `gin-gonic`, `google.golang.org/grpc`, `pkg/sftp`, `minio-go`, `nscache`, `quic-go`, `oauth2`, and `protobuf` in `go.mod`/`go.sum` found no files/matches. |
| 5 | A failing automated guard identifies old runtime packages in the default library dependency graph. | ✓ VERIFIED | `ftpsync/dependency_boundary_test.go` contains `TestPackageDependencyBoundaryRejectsOldRuntime`, runs exact command text `go list -deps ./...`, and fails if forbidden old runtime/internal or third-party fragments appear. Current run passes because no forbidden dependencies remain. |
| 6 | Public API tests prove `ftpsync` exposes typed options only and no old config/server/task types. | ✓ VERIFIED | `ftpsync/public_api_test.go` contains `TestPublicAPIStaysTypedOptionsOnly`, `reflect.TypeOf(NewFTPSyncService)`, signature assertions, approved typed-option compile references, and rejected public markers `Config`, `VFS`, `Server`, `Task`. |
| 7 | `FTPSyncService.SyncOnce` no longer imports legacy `core`, `sync`, `logger`, `retry`, or `ignore` packages. | ✓ VERIFIED | `ftpsync/oneshot.go` imports only standard library packages; grep for old imports in production `ftpsync` code found only blacklist strings inside `dependency_boundary_test.go`, not runtime imports. |
| 8 | Local→FTP and FTP→local one-shot tests still pass with the package-local FTP core. | ✓ VERIFIED | `ftpsync/oneshot_test.go` includes `TestSyncOnceLocalToFTP`, `TestSyncOnceFTPToLocalSuccess`, `TestSyncOnceFTPToLocalAutoCreateRoot`, partial-failure tests, and cwd-safety tests using package-local `ftpCore` fakes. Full `go test ./... -count=1` passed. |
| 9 | Path encoding, retry, ignore/filtering, cwd safety, result, and hook behavior remain behind the public API. | ✓ VERIFIED | Path codec and GBK/UTF-8/auto handling exist in `internal_ftp.go`; retry is in `internal_retry.go`; ignore matching is in `oneshot.go`; cwd safety is enforced by `ensureTargetUnderRoot`/`localTargetPath`; result/hook tests are present in `oneshot_test.go`. |
| 10 | Default module packages contain only the FTP sync library and retained internal helpers. | ✓ VERIFIED | `go list ./...` returned exactly one package: `github.com/no-src/gofs/ftpsync`. |
| 11 | Old CLI/server/protocol runtime directories are removed or no longer part of `go list ./...`. | ✓ VERIFIED | Glob for Go files under `cmd`, `action`, `api`, `auth`, `conf`, `core`, `daemon`, `driver`, `flag`, `monitor`, `report`, `server`, and `sync` returned no files; `go list ./...` contains no old runtime package paths. |
| 12 | The retained library still passes all `ftpsync` tests after old runtime deletion. | ✓ VERIFIED | `go test ./... -count=1` passed with only `github.com/no-src/gofs/ftpsync`; dependency boundary targeted test also passed. |
| 13 | `go.mod` and `go.sum` no longer retain unnecessary runtime dependencies. | ✓ VERIFIED | Module grep found no forbidden runtime dependencies in `go.mod`/`go.sum`; root dependency graph is reduced to FTP/watch/text plus standard library/transitive support. |
| 14 | `go test ./...` passes for the reduced root module. | ✓ VERIFIED | Ran `go test ./... -count=1` → `ok github.com/no-src/gofs/ftpsync 7.538s`. |
| 15 | `go list -deps ./...` passes the explicit blacklist from D-11. | ✓ VERIFIED | Ran `go test ./ftpsync -run TestPackageDependencyBoundaryRejectsOldRuntime -count=1` → passed; raw `go list -deps ./...` output contained no forbidden dependency fragments. |

**Score:** 15/15 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
|---|---|---|---|
| `ftpsync/dependency_boundary_test.go` | Dependency blacklist test for Phase 8 pruning | ✓ VERIFIED | Contains `TestPackageDependencyBoundaryRejectsOldRuntime`, exact `go list -deps ./...` command text, and forbidden list for old runtime packages/modules. Wired into `go test ./...`. |
| `ftpsync/public_api_test.go` | Public API leak regression test | ✓ VERIFIED | Contains reflection assertions for `NewFTPSyncService`, `SyncOnce`, `StartBackground`, and a `go doc -all` public type-name scan. Wired into package tests. |
| `ftpsync/internal_ftp.go` | Package-local FTP client operations | ✓ VERIFIED | 433-line substantive implementation with `type ftpClient`, `github.com/jlaffaye/ftp`, path codec, walk/read/write/remove/mkdir/close support. Used via `newFTPClient`/`openFTPClient` in `oneshot.go`. |
| `ftpsync/internal_retry.go` | Package-local context-aware retry helper | ✓ VERIFIED | Contains `retryWithContext(ctx, RetryOptions, operation, fn)` with bounded attempts, timers, and context cancellation. Called from local→FTP mkdir/write paths. |
| `ftpsync/oneshot.go` | One-shot orchestration without legacy adapter imports | ✓ VERIFIED | Contains `executeSyncOnce`, local→FTP and FTP→local runners, ignore matching, cwd safety, result/error classification, and no old runtime imports. |
| `ftpsync/` | Surviving public FTP sync library package | ✓ VERIFIED | Only package discovered by `go list ./...`; production and test files compile/pass. |
| `go.mod` | Root module definition and tidied dependency set | ✓ VERIFIED | Module remains `github.com/no-src/gofs`; dependencies are reduced to FTP/background-library needs, with no forbidden runtime dependencies. |
| `.github/workflows/go.yml` | Reduced CI workflow after runtime pruning | ✓ VERIFIED | Go matrix is `1.24` only; workflow runs `go build ./...` and `go test ... ./...`; grep found no removed `cmd/gofs`, integration, SFTP, MinIO, Docker, or release references. |

### Key Link Verification

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `ftpsync/dependency_boundary_test.go` | `go list -deps ./...` | `exec.Command` | ✓ WIRED | Lines 10-18 run root-scoped dependency graph command from repo root (`cmd.Dir = ".."`). |
| `ftpsync/oneshot.go` | `ftpsync/internal_ftp.go` | package-local FTP operations | ✓ WIRED | `openFTPClient` calls `newFTPClient`; one-shot runners use `ftpCore` methods. |
| `go list ./...` | `ftpsync/` | default package discovery | ✓ WIRED | Actual command output is only `github.com/no-src/gofs/ftpsync`. |
| `go.mod` | `go list -deps ./...` | module dependency resolution | ✓ WIRED | `go list -deps ./...` resolves only retained FTP/watch/text dependencies plus standard library/transitive support. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
|---|---|---|---|---|
| `ftpsync/oneshot.go` | `svc.opts` / `Result` | Public `Options` copied by `NewFTPSyncService`, then consumed by `executeSyncOnce` | Yes | ✓ FLOWING — typed options drive FTP client construction, local walking, remote path mapping, hooks, and result counters. |
| `ftpsync/internal_ftp.go` | `FTPOptions` | `openFTPClient(ctx, opts)` from one-shot execution | Yes | ✓ FLOWING — host/port/credentials/passive/timeout/path encoding flow into FTP dial/login/path operations. |
| `ftpsync/background.go` | source filesystem events | `fsnotify.NewWatcher`, `watchTree`, `runWatchLoop` | Yes | ✓ FLOWING — events queue debounced calls to `executeSyncOnce`; background tests exercise real temp-dir events. |
| `ftpsync/dependency_boundary_test.go` | dependency list | `go list -deps ./...` stdout | Yes | ✓ FLOWING — test parses actual Go tool output and fails on forbidden matches. |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
|---|---|---|---|
| Reduced module exposes only the FTP sync package | `go list ./...` | `github.com/no-src/gofs/ftpsync` | ✓ PASS |
| Reduced dependency graph is resolvable and excludes old runtime graph | `go list -deps ./...` | Completed successfully; output contains standard library, `fsnotify`, `jlaffaye/ftp`, `x/text`, and `github.com/no-src/gofs/ftpsync`; no forbidden runtime fragments observed. | ✓ PASS |
| Full reduced root module test suite passes | `go test ./... -count=1` | `ok github.com/no-src/gofs/ftpsync 7.538s` | ✓ PASS |
| Dependency boundary guard passes | `go test ./ftpsync -run TestPackageDependencyBoundaryRejectsOldRuntime -count=1` | `ok github.com/no-src/gofs/ftpsync 0.302s` | ✓ PASS |
| Runtime dependency names absent from module files | grep scan for `gin-gonic|google.golang.org/grpc|pkg/sftp|minio-go|nscache|quic-go|oauth2|protobuf` in `go.*` | No files/matches found. | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
|---|---|---|---|---|
| PRUNE-01 | 08-01, 08-02, 08-03, 08-04 | Library build no longer depends on CLI entrypoints, flag parsing, daemon/process management, HTTP file server, gRPC APIs, task runtime, auth/session, SFTP, MinIO, or Docker release surfaces. | ✓ SATISFIED | `go list ./...` only lists `ftpsync`; old runtime package glob returns no Go package files; dependency blacklist test passes; Docker/release workflows are removed and Go workflow has no stale removed-path references. |
| PRUNE-02 | 08-02, 08-03, 08-04 | FTP driver, path encoding, retry, ignore/filtering, rate limiting, and core disk/FTP sync internals remain available behind public library API. | ✓ SATISFIED | Package-local FTP implementation, path codec, retry helper, ignore matcher, one-shot sync, and background fsnotify flow are wired through `FTPSyncService`; full tests pass. Note: no separate rate package remains, but FTP behavior is retained behind the package-local core without leaking old runtime dependencies. |
| PRUNE-03 | 08-01, 08-02, 08-03, 08-04 | Internal packages do not leak old `conf.Config`, `core.VFS`, CLI URL parsing, or server/task types into public `ftpsync` API. | ✓ SATISFIED | `Options`/`Endpoint`/`FTPOptions` public API contains typed fields only; reflection/doc public API guard passes; grep found no production imports of old config/core/server/task packages. |
| PRUNE-04 | 08-01, 08-04 | `go.mod` is tidied so removed runtime surfaces no longer keep unnecessary dependencies such as Gin, gRPC, SFTP, MinIO, Redis/cache, QUIC, OAuth, and protobuf packages. | ✓ SATISFIED | `go.mod` has only FTP/watch/text dependencies plus minimal indirect test/sys deps; grep scan over `go.mod`/`go.sum` found none of the forbidden runtime modules. |

No orphaned Phase 8 requirements found: `REQUIREMENTS.md` maps exactly PRUNE-01 through PRUNE-04 to Phase 8, and all four are declared across Phase 8 plan frontmatter.

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|---|---:|---|---|---|
| `ftpsync/*.go` | various | `return nil` matches from generic scan | ℹ️ Info | Benign Go control-flow returns in implementations/tests; not placeholders or hollow stubs. |
| `ftpsync/dependency_boundary_test.go` | 21-48 | Forbidden dependency strings | ℹ️ Info | Intentional blacklist literals, not imports. |

No blocker or warning anti-patterns were found. No TODO/FIXME/placeholder/not-implemented user-visible stubs were identified in production `ftpsync` code.

### Human Verification Required

None. This phase is dependency/package-shape and automated test focused; all success criteria are programmatically verifiable with Go tooling and source inspection.

### Gaps Summary

No gaps found. Phase 8 achieved the package-reduction goal: the default module has only the `ftpsync` package, the public API remains typed-options-only, old runtime/protocol dependencies are absent from the build graph and module files, and retained FTP sync internals remain wired through `FTPSyncService` with passing tests.

---

_Verified: 2026-04-28T00:00:00Z_
_Verifier: the agent (gsd-verifier)_
