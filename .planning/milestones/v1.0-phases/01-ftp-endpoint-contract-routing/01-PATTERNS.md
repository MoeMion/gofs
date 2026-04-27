# Phase 01 Patterns — FTP Endpoint Contract & Routing

## Analog files to follow

### VFS parsing pattern

- `core/vfs.go`
- `core/vfs_test.go`

Use the existing pattern where a scheme-specific branch in `NewVFS` calls `parse(path, fsType)` and default-port logic lives inside `parse`.

### Thin sync wrapper pattern

- `sync/sftp_push_client_sync.go`
- `sync/sftp_pull_client_sync.go`

Use protocol-specific constructor files that only adapt the generic sync helpers rather than creating a new sync engine.

### Thin pull monitor pattern

- `monitor/sftp_pull_client_monitor.go`

Use a minimal wrapper around `driverPullClientMonitor` for remote-source monitor routing.

### Routing factory pattern

- `sync/sync.go`
- `monitor/monitor.go`

Insert FTP branches alongside SFTP and MinIO branches instead of adding a separate orchestration path.

## Concrete pattern notes

- Scheme constants are declared near the top of `core/vfs.go` with a matching default-port constant.
- Existing remote endpoint parsing relies on query keys such as `path`, `remote_path`, and backend-specific booleans.
- `core/vfs_test.go` centralizes canonical endpoint strings as package constants and reuses them across marshal/unmarshal/default-port tests.
- Protocol wrappers in `sync/` generally keep constructor-only responsibility and let generic helpers own file operations.
