# Phase 1 Research — FTP Endpoint Contract & Routing

**Phase:** 01 — FTP Endpoint Contract & Routing
**Researched:** 2026-04-23
**Status:** Complete

## Objective

Answer what is needed to plan Phase 1 well: how gofs should express FTP endpoints in configuration and route them through existing factories without prematurely implementing full FTP backend behavior.

## Key Findings

### Recommended approach

- Extend `core.VFS` to recognize `ftp://` endpoints using the same query-parameter style already used by `sftp://` and `minio://` per D-01 and D-02.
- Add FTP-specific endpoint fields on `core.VFS` instead of reusing `SSHConfig`, so credentials and transport behavior stay explicit per D-03 and D-04.
- Default FTP port to `21` when omitted per D-07.
- Represent timeout as an optional endpoint-level field per D-08.
- Represent passive-mode compatibility as a single boolean field in Phase 1 per D-05, without adding extra compatibility knobs such as EPSV/MLSD/UTF-8 flags per D-06.
- Route FTP endpoints through `sync.NewSync` and `monitor.NewMonitor` so `ftp://` is classified as a real backend path instead of falling through to generic unsupported-path logic.

### Brownfield implementation pattern

- `core/vfs.go` is the highest-value insertion point for this phase.
- `sync/sync.go` and `monitor/monitor.go` should gain FTP-specific dispatch branches parallel to SFTP and MinIO.
- Because Phase 2 owns the real protocol driver, Phase 1 should only make FTP endpoints reachable by configuration and factory routing. Any transport execution beyond routing belongs to later work.

### Existing analogs to reuse

- `core/vfs.go` — existing query parsing, scheme detection, default-port logic.
- `core/vfs_test.go` — scheme-specific parsing and default-port tests for `rs://`, `sftp://`, and `minio://`.
- `sync/sftp_push_client_sync.go` and `sync/sftp_pull_client_sync.go` — thin protocol wrapper pattern used by remote backends.
- `monitor/sftp_pull_client_monitor.go` — thin remote pull monitor wrapper pattern.

## Constraints for planning

- Do not introduce standard FTP authority-style parsing like `ftp://user:pass@host/path`; preserve the existing query-driven VFS grammar.
- Do not reuse SSH names such as `ssh_user` / `ssh_pass` for FTP endpoint auth.
- Do not add FTPS, active mode, or broader server-quirk toggles in this phase.
- Do not expand into HTTP file-server or FTP server-mode work; this milestone is client-side sync backend support only.

## Risks to account for in plans

- The repo already has `core.FTP` in `core/vfs_type.go`, but `core.NewVFS`, `sync.NewSync`, and `monitor.NewMonitor` do not route FTP yet.
- If Phase 1 changes config parsing without tests, FTP endpoints may still degrade to `Disk` or `Unknown` behavior.
- If FTP credentials are modeled with SSH fields, later driver work becomes ambiguous and documentation debt increases.

## Recommended deliverables for Phase 1

1. `core.VFS` FTP contract: parsing, getters, defaults, and tests.
2. `sync.NewSync` and `monitor.NewMonitor` FTP routing updates.
3. Thin FTP sync/monitor entry points sufficient to prove routing reaches FTP-specific code paths, while leaving real backend behavior to Phase 2.

## Sources

- `.planning/phases/01-ftp-endpoint-contract-routing/01-CONTEXT.md`
- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/research/SUMMARY.md`
- `.planning/research/STACK.md`
- `.planning/research/FEATURES.md`
- `.planning/research/ARCHITECTURE.md`
- `.planning/research/PITFALLS.md`
- `core/vfs.go`
- `core/vfs_test.go`
- `sync/sync.go`
- `monitor/monitor.go`
