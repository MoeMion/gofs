---
phase: 01-ftp-endpoint-contract-routing
reviewed: 2026-04-23T00:00:00Z
depth: standard
files_reviewed: 9
files_reviewed_list:
  - core/vfs.go
  - core/vfs_test.go
  - sync/sync.go
  - sync/sync_test.go
  - monitor/monitor.go
  - monitor/monitor_test.go
  - sync/ftp_push_client_sync.go
  - sync/ftp_pull_client_sync.go
  - monitor/ftp_pull_client_monitor.go
findings:
  critical: 0
  warning: 1
  info: 1
  total: 2
status: issues_found
---

# Phase 1: Code Review Report

**Reviewed:** 2026-04-23T00:00:00Z
**Depth:** standard
**Files Reviewed:** 9
**Status:** issues_found

## Summary

Reviewed the Phase 1 FTP endpoint contract and routing changes across VFS parsing, sync routing, monitor routing, and their focused tests. The routing work is intentionally thin and generally matches the phase plans, but the FTP parser currently treats malformed explicit ports as if the port were omitted, which weakens the endpoint contract and can hide configuration mistakes.

## Warnings

### WR-01: Invalid explicit FTP port silently falls back to default port 21

**File:** `core/vfs.go:223-241`
**Issue:** `parse()` calls `strconv.Atoi(parseUrl.Port())` and treats any conversion error as “no port specified,” then applies the default FTP port. That means a malformed explicit endpoint such as `ftp://host:abc?...` can be accepted as `ftp://host:21?...` instead of being rejected. This contradicts the intended contract validation for FTP endpoints and can route users to the wrong service while masking bad configuration.
**Fix:** Distinguish between an omitted port and an invalid explicit port. Only use the default when `parseUrl.Port()` is empty; otherwise return an error.

```go
rawPort := parseUrl.Port()
if rawPort == "" {
	if scheme == ftpServerScheme {
		port = ftpServerDefaultPort
	}
} else {
	port, err = strconv.Atoi(rawPort)
	if err != nil {
		return scheme, host, 0, localPath, remotePath, isServer, fsServer, localSyncDisabled, secure, ftpConf, sshConf,
			fmt.Errorf("invalid %s port %q: %w", scheme, rawPort, err)
	}
}
```

## Info

### IN-01: Tests cover omitted-port behavior but not malformed explicit FTP ports

**File:** `core/vfs_test.go:104-122,328-342`
**Issue:** The new FTP coverage verifies default-port behavior when the port is omitted, but it does not add a regression test for an invalid explicit port. Because of that gap, the parser bug above can ship unnoticed.
**Fix:** Add a focused test such as `ftp://127.0.0.1:abc?...` and assert that `NewVFS` returns `NewEmptyVFS()` (or the chosen explicit failure behavior) rather than defaulting to port `21`.

---

_Reviewed: 2026-04-23T00:00:00Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
