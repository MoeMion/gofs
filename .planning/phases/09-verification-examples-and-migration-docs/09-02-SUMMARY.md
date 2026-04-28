---
phase: 09-verification-examples-and-migration-docs
plan: 02
subsystem: real-ftp-integration
tags: [verification, integration-test, ftp]
requires: [09-01]
provides: [real-ftp-local-to-ftp-test, real-ftp-ftp-to-local-test]
affects: [go.mod, go.sum, ftpsync/real_ftp_integration_test.go]
tech-stack:
  added: [github.com/fclairamb/ftpserverlib, github.com/spf13/afero]
  patterns: [loopback-test-fixture, public-api-integration]
key-files:
  created:
    - ftpsync/real_ftp_integration_test.go
  modified:
    - go.mod
    - go.sum
key-decisions:
  - Kept FTP server implementation test-only and loopback-bound using ftpserverlib with afero-backed temp storage.
requirements-completed: [VERIFY-02, VERIFY-03]
metrics:
  duration: 18 min
  completed: 2026-04-28
---

# Phase 09 Plan 02: Real FTP Integration Summary

Added real FTP integration coverage through the public `FTPSyncService.SyncOnce` API using a Go-native loopback-only test fixture.

## What Changed

- Added test dependency `github.com/fclairamb/ftpserverlib` and its filesystem support dependency.
- Created `ftpsync/real_ftp_integration_test.go` with a local `127.0.0.1` FTP fixture serving a `t.TempDir()` root.
- Verified local→FTP one-shot sync writes expected nested files to the FTP server root.
- Verified FTP→local one-shot sync downloads expected nested files under the configured destination root.

## Tasks Completed

| Task | Description | Commit |
|------|-------------|--------|
| 1-3 | Added test dependency, loopback FTP fixture, and public API integration tests for both one-shot directions. | 3538d70 |

## Verification

| Command | Result |
|---------|--------|
| `go list -m github.com/fclairamb/ftpserverlib` | PASS — `github.com/fclairamb/ftpserverlib v0.30.0` |
| `go test ./ftpsync -run TestIntegrationRealFTPLocalToFTP -count=1` | PASS |
| `go test ./ftpsync -run 'TestIntegrationRealFTP(LocalToFTP\|ToLocal)' -count=1` | PASS |
| `go test ./ftpsync -run 'TestIntegrationRealFTP' -count=1` | PASS |
| `go test ./...` | PASS |

## Deviations from Plan

None - plan executed exactly as written.

## Known Stubs

None.

## Threat Flags

None.

## Self-Check: PASSED

- Created file exists: `ftpsync/real_ftp_integration_test.go`.
- Commit exists: `3538d70`.
- Verification commands passed.
