---
phase: 06
slug: one-shot-disk-ftp-sync-through-library-api
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-27
---

# Phase 06 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing Go package tests |
| **Quick run command** | `go test ./ftpsync -count=1` |
| **Full suite command** | `go test ./ftpsync ./sync ./driver/ftp -count=1` |
| **Estimated runtime** | ~20 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./ftpsync -count=1`
- **After every plan wave:** Run `go test ./ftpsync ./sync ./driver/ftp -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 20 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 06-01-01 | 01 | 1 | ONCE-05 | T-06-01-01 | Result/error contract stays password-safe and returns partial-failure summary behavior | unit | `go test ./ftpsync -run 'TestSyncOnce(Result|Partial|Hooks)' -count=1` | ✅ | ⬜ pending |
| 06-01-02 | 01 | 1 | ONCE-01, ONCE-02 | T-06-01-02 | Typed options build internal VFS/sync dependencies without public CLI/runtime coupling | unit | `go test ./ftpsync -run 'Test(NewFTPSyncService|SyncOnceBuildsLegacyAdapter)' -count=1` | ✅ | ⬜ pending |
| 06-02-01 | 02 | 2 | ONCE-01, ONCE-03 | T-06-02-03 | Local→FTP one-shot reuses legacy push semantics and emits summary/hook updates | unit | `go test ./ftpsync -run 'TestSyncOnceLocalToFTP' -count=1` | ✅ | ⬜ pending |
| 06-02-02 | 02 | 2 | ONCE-03, ONCE-05 | T-06-02-04 | Best-effort execution returns `Result` plus typed transfer error on partial failure | unit | `go test ./ftpsync -run 'TestSyncOnceLocalToFTPPartialFailure' -count=1` | ✅ | ⬜ pending |
| 06-03-01 | 03 | 3 | ONCE-02, ONCE-04 | T-06-03-01 | FTP→local writes stay under explicit destination root and never leak to cwd | unit | `go test ./ftpsync -run 'TestSyncOnceFTPToLocal(CWDSafety|AutoCreateRoot)' -count=1` | ✅ | ⬜ pending |
| 06-03-02 | 03 | 3 | ONCE-03, ONCE-05 | T-06-03-02 | FTP→local preserves remote metadata comparison semantics and returns summary counts | unit | `go test ./ftpsync ./sync ./driver/ftp -run 'TestSyncOnceFTPToLocal' -count=1` | ✅ | ⬜ pending |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 20s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
