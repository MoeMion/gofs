---
phase: 02-ftp-driver-backend
reviewed: 2026-04-24T02:33:48Z
depth: standard
files_reviewed: 8
files_reviewed_list:
  - driver/ftp/ftp.go
  - driver/ftp/file.go
  - driver/ftp/file_info.go
  - driver/ftp/ftp_test.go
  - sync/ftp_push_client_sync.go
  - sync/ftp_pull_client_sync.go
  - sync/sync_test.go
  - go.mod
findings:
  critical: 0
  warning: 1
  info: 1
  total: 2
status: issues_found
---

# Phase 2: Code Review Report

**Reviewed:** 2026-04-24T02:33:48Z
**Depth:** standard
**Files Reviewed:** 8
**Status:** issues_found

## Summary

Reviewed the Phase 2 FTP driver backend and FTP sync constructor wiring, including the new `driver/ftp` package, constructor routing in `sync/ftp_*`, associated tests, and dependency wiring in `go.mod`.

The overall architecture is aligned with the phase plan: FTP is integrated through the existing `driver.Driver` seam, pull sync metadata hooks are wired correctly, and tests cover the main connect/CRUD/reconnect paths. The primary issue is an overly broad interpretation of FTP `550` responses that can silently convert real server failures into success paths. I also found one leftover Phase 1 placeholder artifact that should be removed to avoid misleading future maintenance.

## Warnings

### WR-01: Generic `550` handling can silently hide real FTP failures

**File:** `driver/ftp/ftp.go:274-275,299-300,307-308,461-475`
**Issue:** `isFTPAlreadyExists` and `isFTPNotExist` both treat any error message containing `550` as a benign condition. On many FTP servers, `550` is a generic failure code used for permission denied, unavailable path, or other server-side errors, not just already-exists / not-found cases. That means `MkdirAll` can incorrectly ignore directory-creation failures, and `Remove` can incorrectly report success when deletion actually failed.
**Fix:** Narrow the classifier to explicit message text for the intended condition, or parse structured FTP status where available. Do not treat bare `550` as success-equivalent.

```go
func isFTPAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "file exists") ||
		strings.Contains(message, "directory already exists")
}

func isFTPNotExist(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "not exist") ||
		strings.Contains(message, "not found") ||
		strings.Contains(message, "no such file")
}
```

Also add tests that distinguish `550 not found` from `550 permission denied` so this behavior cannot regress silently.

## Info

### IN-01: Phase 1 deferred-error sentinel remains as dead code

**File:** `sync/ftp_push_client_sync.go:13`
**Issue:** `errFTPBackendDeferred` is still declared even though Phase 2 replaced the deferred placeholder flow with the real FTP constructor path. Keeping it around is misleading because the FTP backend is no longer deferred.
**Fix:** Remove the unused sentinel and the now-unneeded `errors` import from this file.

---

_Reviewed: 2026-04-24T02:33:48Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
