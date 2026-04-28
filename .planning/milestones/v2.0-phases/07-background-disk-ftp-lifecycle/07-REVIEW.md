---
phase: 07-background-disk-ftp-lifecycle
reviewed: 2026-04-27T09:04:05Z
depth: standard
files_reviewed: 4
files_reviewed_list:
  - ftpsync/background.go
  - ftpsync/background_test.go
  - ftpsync/service.go
  - ftpsync/context_test.go
findings:
  critical: 0
  warning: 2
  info: 0
  total: 2
status: issues_found
---

# Phase 07: Code Review Report

**Reviewed:** 2026-04-27T09:04:05Z
**Depth:** standard
**Files Reviewed:** 4
**Status:** issues_found

## Summary

Reviewed the FTP background local-to-FTP lifecycle implementation and its related public contract tests. The background runner has a reasonable cancel/stop shape and protects handle state with a mutex, but two correctness gaps can cause missed syncs or invalid endpoint configuration to be accepted.

## Warnings

### WR-01: Events that arrive during an in-flight sync can be dropped

**File:** `ftpsync/background.go:228-232`

**Issue:** `queueSync` uses a non-blocking send into a one-element trigger channel and silently drops the event when a sync is already pending or being processed. If a filesystem change occurs while `executeSyncOnce` is still running, that change can be coalesced away without any follow-up sync pass. This is especially risky for FTP uploads because the in-flight sync may already have walked the source tree before the later file change occurred, leaving the remote endpoint stale until another unrelated change arrives.

**Fix:** Preserve at least one pending follow-up trigger while a sync pass is in progress. One minimal option is to let the sync worker track a `pending` flag and drain/loop until no trigger arrived during the last pass, or to use an explicit mutex/condition-backed scheduler instead of dropping sends. For example:

```go
func (h *backgroundHandle) runSyncTriggers(ctx context.Context, svc *FTPSyncService, trigger <-chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-trigger:
			if !ok {
				return
			}
		}

		for {
			if _, err := executeSyncOnce(ctx, svc); err != nil {
				svc.log("StartBackground sync pass failed: " + err.Error())
				svc.reportEvent(SyncEvent{Operation: "background_sync", Path: svc.opts.Source.LocalPath, Status: "failed", ErrorKind: backgroundErrorKind(err)})
				h.setCurrent(err)
			}

			select {
			case <-ctx.Done():
				return
			case _, ok := <-trigger:
				if !ok {
					return
				}
				continue
			default:
				return
			}
		}
	}
}
```

Also add a regression test where the second sync blocks, a file is changed during that sync, the sync is released, and the test asserts that a third pass runs.

### WR-02: Negative FTP timeout values pass validation

**File:** `ftpsync/service.go:125-135`

**Issue:** `validateFTPFields` validates host, remote path, and positive port, but it does not reject negative `FTPOptions.Timeout` values. A negative timeout is treated as an FTP field by `hasFTPFields`, but later option conversion only forwards positive timeouts, so invalid user input is silently accepted and normalized away. That can hide configuration mistakes and produce behavior different from what the caller requested.

**Fix:** Reject negative timeouts during endpoint validation while continuing to allow zero as “use default/no override”:

```go
func validateFTPFields(name string, ftp FTPOptions) error {
	if ftp.Host == "" {
		return newError(ErrValidation, name+" host is required", errInvalidOptions)
	}
	if ftp.RemotePath == "" {
		return newError(ErrValidation, name+" remote path is required", errInvalidOptions)
	}
	if ftp.Port <= 0 {
		return newError(ErrValidation, name+" port must be positive", errInvalidOptions)
	}
	if ftp.Timeout < 0 {
		return newError(ErrValidation, name+" timeout must be non-negative", errInvalidOptions)
	}
	return nil
}
```

Add validation tests for both local-to-FTP destination timeouts and FTP-to-local source timeouts.

---

_Reviewed: 2026-04-27T09:04:05Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
