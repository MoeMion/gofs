---
phase: 03
slug: one-way-ftp-sync-flows
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-04-24
---

# Phase 03 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none |
| **Quick run command** | `go test ./monitor ./sync -count=1` |
| **Full suite command** | `go test ./driver/ftp ./monitor ./sync -count=1` |
| **Estimated runtime** | ~30 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./monitor ./sync -count=1`
- **After every plan wave:** Run `go test ./driver/ftp ./monitor ./sync -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 30 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 03-01-01 | 01 | 1 | SYNC-02, SYNC-03 | T-03-01 / T-03-02 | FTP source mode fails explicitly when polling/once mode is absent instead of idling silently | unit | `go test ./monitor -count=1` | ✅ | ⬜ pending |
| 03-01-02 | 01 | 1 | SYNC-02, SYNC-03 | T-03-03 | FTP monitor routing and startup semantics are covered by regression tests | unit | `go test ./monitor -count=1` | ✅ | ⬜ pending |
| 03-02-01 | 02 | 2 | SYNC-01, SYNC-02, SYNC-03 | T-03-04 | FTP one-way flow tests prove create/update/delete/rename semantics stay one-way | unit | `go test ./sync -count=1` | ✅ | ⬜ pending |
| 03-02-02 | 02 | 2 | SYNC-04 | T-03-05 | Supported-metadata no-op behavior is verified without hiding ambiguous-metadata retransfers | unit | `go test ./driver/ftp ./sync -count=1` | ✅ | ⬜ pending |

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
- [x] Feedback latency < 30s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
