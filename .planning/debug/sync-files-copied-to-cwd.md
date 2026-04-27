---
status: resolved
trigger: "同步的文件会被复制一份到程序的执行目录下"
created: 2026-04-27
updated: 2026-04-27
---

# Debug Session: sync-files-copied-to-cwd

## Symptoms

- Expected behavior: synced files should be written only to the configured destination path.
- Actual behavior: synced files are also copied into the program execution directory.
- Error messages: none reported.
- Timeline: observed after recent FTP sync work.
- Reproduction: run a sync flow and inspect the process working directory for extra copied files.

## Current Focus

- hypothesis: a sync path or FTP driver path is using a relative path without joining it to the configured destination/root path.
- test: inspect disk/driver sync write paths and targeted tests around destination path handling.
- expecting: a code path that calls create/open/write with a relative target path derived from the source path.
- next_action: fixed and verified

## Evidence

- timestamp: 2026-04-27T10:54:00Z
  observation: `core.newPath("")` normalizes an omitted FTP `path` query to `.`; `driverPushClientSync` then treats local sync as enabled when `local_sync_disabled` is false/omitted.
  implication: FTP push destinations without an explicit local `path` mirror files into the process current working directory before uploading to FTP.
- timestamp: 2026-04-27T10:55:00Z
  observation: `driverPushClientSync.Create/Write/Remove/Symlink` all invoked `diskSync` based only on `!dest.LocalSyncDisabled()`.
  implication: remote-only FTP pushes could still perform local disk mutations at the implicit cwd path.
- timestamp: 2026-04-27T10:56:00Z
  observation: Added regression `TestFTPPushClientSync_SkipsImplicitCWDLocalMirrorWhenPathOmitted`, which changes cwd to a temp dir and verifies FTP create goes remote without creating `leaked.txt` in cwd.
  implication: The fixed path prevents the reported extra cwd copy while preserving remote FTP writes.

## Eliminated

## Resolution

- root_cause: FTP destination URLs that omit the local `path` query were normalized to `.` and the driver push sync treated that implicit cwd path as an enabled local mirror.
- fix: Added `core.VFS.HasLocalPath()` and changed driver-backed push sync to perform optional local mirroring only when a local path was explicitly configured and `local_sync_disabled` is false.
- verification: `go test ./sync` passed; `go test ./core -run 'TestNewVFS_FTP|TestVFS_Compare'` passed; `go test ./core ./sync` still fails existing SSH-config-dependent core tests unrelated to this change.
- files_changed: core/vfs.go; sync/driver_push_client_sync.go; sync/sync_test.go; .planning/debug/sync-files-copied-to-cwd.md
