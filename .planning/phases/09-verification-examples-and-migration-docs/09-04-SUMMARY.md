---
phase: 09-verification-examples-and-migration-docs
plan: 04
subsystem: adoption-and-migration-docs
tags: [documentation, migration, readme]
requires: [09-01]
provides: [README-adoption-guide, migration-notes]
affects: [README.md, MIGRATION.md]
tech-stack:
  added: []
  patterns: [typed-options-docs, migration-guide]
key-files:
  created:
    - MIGRATION.md
  modified:
    - README.md
key-decisions:
  - README and migration docs identify `import "ftpsync/ftpsync"` as the supported package path while source remains in the `ftpsync/` subdirectory.
requirements-completed: [DOC-01, DOC-02, DOC-03]
metrics:
  duration: 10 min
  completed: 2026-04-28
---

# Phase 09 Plan 04: Adoption and Migration Docs Summary

Replaced stale application-oriented documentation with final local-library adoption guidance and explicit breaking migration notes.

## What Changed

- Rewrote `README.md` for the focused `ftpsync` local Go module.
- Added typed-option snippets for one-shot local→FTP, one-shot FTP→local, and background local→FTP usage.
- Added `MIGRATION.md` documenting the breaking migration from the old CLI/server application to package invocation.
- Listed removed and unsupported surfaces exactly, including no CLI runtime, no HTTP/gRPC/file server/task/auth/session runtime, no SFTP, no MinIO, no Docker release artifact, no FTPS, no FTP server mode, no FTP<->FTP sync, no FTP->disk background polling, and no bidirectional conflict resolution.

## Tasks Completed

| Task | Description | Commit |
|------|-------------|--------|
| 1-3 | Rewrote README, added migration notes, and completed docs consistency checks. | 6f2c782 |

## Verification

| Command | Result |
|---------|--------|
| `go test ./ftpsync -run 'ExampleFTPSyncService' -count=1` | PASS |
| `go test ./...` | PASS |
| `go test ./... && go test ./ftpsync -run 'ExampleFTPSyncService' -count=1` | PASS |
| `git grep -n -E 'go install github.com/no-src/gofs\|docker pull nosrc/gofs\|SFTP Push Client\|MinIO Push Client\|Task Server\|File Server\|github.com/no-src/gofs/ftpsync' -- README.md || true` | PASS — no README matches |

## Deviations from Plan

None beyond the previously approved Phase 9 plan adjustment from `import "ftpsync"` to `import "ftpsync/ftpsync"` for the current subdirectory package layout.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

- Created file exists: `MIGRATION.md`.
- Commit exists: `6f2c782`.
- Verification commands passed.
