# Phase 8: FTP-Only Package Reduction - Research

**Researched:** 2026-04-27
**Status:** Complete

## Research Question

What must be known to plan a full-fidelity reduction of the root Go module into a focused FTP sync library while preserving `ftpsync.FTPSyncService` behavior from Phases 5-7?

## Current Dependency Leak

`ftpsync/oneshot.go` is the critical import fan-out:

- Imports `core`, `ignore`, `logger`, `retry`, and legacy `sync`.
- Legacy `sync.NewSync` fans out into disk, remote disk, push client, SFTP, FTP, MinIO, server, API/gRPC, auth, report, and logger runtime paths.
- `core.VFS` brings legacy endpoint parsing and URL/config concepts that Phase 8 must keep out of the public API and default library dependency graph.
- `driver/ftp` contains the desired FTP behavior but imports `core`, `driver`, `logger`, `retry`, and `net/http` file-serving abstractions.

## Recommended Strategy

Use the locked minimal inline FTP core strategy:

1. Add automated dependency-boundary tests first so pruning cannot silently regress.
2. Replace the legacy adapter in `ftpsync/oneshot.go` with package-local FTP operations and local filesystem walking.
3. Internalize only the FTP behaviors needed by the public API:
   - FTP connect/login/binary transfer setup using `github.com/jlaffaye/ftp`.
   - passive-mode rejection/validation consistent with existing v1 behavior.
   - timeout and path encoding behavior equivalent to the retained FTP driver behavior.
   - create directory, write file, read file, walk remote tree, delete/remove, rename-relevant semantics used by one-shot and background resync.
   - typed ignore matching already present in `ftpsync` can remain package-local and replace `ignore.PathIgnore`.
   - retry can be a small package-local helper using `RetryOptions`, context, and `time.Timer`; no legacy logger/retry package required.
4. Remove old runtime directories and release surfaces after `ftpsync` no longer imports them.
5. Run `go mod tidy` and verify `go list -deps ./...` no longer includes blacklisted packages or third-party runtime dependencies.

## Dependency Blacklist

The proof must cover at least these import path fragments and module names:

- Internal packages: `/cmd`, `/flag`, `/daemon`, `/server`, `/api`, `/auth`, `/monitor`, `/driver/sftp`, `/driver/minio`, `/conf`, `/core`, `/sync`, `/logger`, `/report`, `/eventlog`, `/action`, `/task`.
- Third-party modules: `github.com/gin-gonic/gin`, `google.golang.org/grpc`, `github.com/pkg/sftp`, `github.com/minio/minio-go/v7`, `github.com/no-src/nscache`, `github.com/quic-go/quic-go`, `golang.org/x/oauth2`, `google.golang.org/protobuf`, `github.com/gin-contrib`, Redis/cache/session modules, and Docker/release-only surfaces.

## Validation Architecture

Dimension 1 — Public API preservation:
- `go test ./ftpsync -run 'Test(NewFTPSyncService|Validate|SyncOnce|StartBackground|Hook|Background)' -count=1`

Dimension 2 — Dependency graph reduction:
- `go list -deps ./...` piped through a blacklist check fails if old runtime packages or removed third-party modules appear.

Dimension 3 — Module cleanup:
- `go mod tidy` followed by blacklist checks on `go.mod` and `go.sum` proves runtime dependencies are removed.

Dimension 4 — Behavior preservation:
- Existing `ftpsync` tests for local→FTP one-shot, FTP→local one-shot, cwd safety, partial failure, hooks, and background lifecycle must pass after replacing the legacy adapter.

## Planning Implications

- Build tests before extraction; otherwise a shallow import rewrite could appear successful while leaving dependency leaks.
- Split implementation from deletion. First make `ftpsync` independent; then delete old runtime directories and tidy modules.
- Keep Phase 9 docs/release wording deferred, but Phase 8 must leave an unambiguous package shape for Phase 9 to document.

## Out of Scope

- FTPS, active FTP, FTP server mode, bidirectional sync, FTP→disk background polling, and legacy YAML/CLI parsing remain deferred.

## Research Complete

Phase 8 can be planned without additional external discovery. The only retained external runtime dependency expected for the library is the FTP client package plus `fsnotify` for approved background disk→FTP monitoring and standard/transitive support packages required by those dependencies.
