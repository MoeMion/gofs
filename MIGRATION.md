# Migration to ftpsync v2.0

## Supported import path

v2.0 is a breaking migration from the old `gofs` CLI/server application to a local Go library module named `ftpsync`.

Because package source remains in the `ftpsync/` subdirectory under module `ftpsync`, the supported package import path is:

```go
import "ftpsync/ftpsync"
```

See [README.md](README.md) for typed-option examples.

## What changed

- The supported integration point is now `ftpsync.FTPSyncService` rather than a process-level CLI or server runtime.
- Callers create `ftpsync.Options` in Go code and call `SyncOnce` or `StartBackground` directly.
- The repository is scoped to FTP sync library behavior for v2.0.

## Removed and unsupported surfaces

- no CLI runtime
- no HTTP/gRPC/file server/task/auth/session runtime
- no SFTP
- no MinIO
- no Docker release artifact
- no FTPS
- no FTP server mode
- no FTP<->FTP sync
- no FTP->disk background polling
- no bidirectional conflict resolution

## Configuration migration

Callers now configure everything with typed Go values instead of YAML or CLI config files:

- `Options` selects direction, endpoints, retry behavior, ignore rules, and hooks.
- `Endpoint` represents either a local filesystem path or an FTP endpoint.
- `FTPOptions` carries host, port, username, password, remote path, passive mode, timeout, and path encoding.
- `RetryOptions` configures retry count, wait duration, and async retry behavior.
- `IgnoreRule` configures literal, glob, or regexp filtering.
- `HookSet` configures library-local logging, progress, and event callbacks.

## Capability mapping

| Old capability | v2.0 library equivalent |
|----------------|--------------------------|
| One-shot FTP push | `DirectionLocalToFTP` + `SyncOnce` |
| One-shot FTP pull | `DirectionFTPToLocal` + `SyncOnce` |
| Persistent local FTP push | `DirectionLocalToFTP` + `StartBackground` |
| YAML or CLI flag configuration | Typed `Options`, `Endpoint`, `FTPOptions`, `RetryOptions`, `IgnoreRule`, and `HookSet` |

The old `github.com/no-src/gofs` application repository name may appear in historical context, but it is not the supported final consumer package path for v2.0 library adoption.
