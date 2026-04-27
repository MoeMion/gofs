---
phase: 02
slug: ftp-driver-backend
status: passed
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-24
---

# Phase 02 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing Go package tests |
| **Quick run command** | `go test ./driver/ftp ./sync -count=1` |
| **Full suite command** | `go test ./driver/... ./sync ./monitor -count=1` |
| **Estimated runtime** | ~45 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./driver/ftp ./sync -count=1`
- **After every plan wave:** Run `go test ./driver/... ./sync ./monitor -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 45 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 02-01-01 | 01 | 1 | FTPD-01, FTPD-02, FTPD-05, FTPD-06, FTPD-07 | T-02-01 | Driver serializes connection use and returns explicit auth/operation errors | unit | `go test ./driver/ftp -count=1` | ✅ | ✅ green |
| 02-01-02 | 01 | 1 | FTPD-03, FTPD-04, FTPD-08, FTPD-09 | T-02-02 | Upload/download/time logic prefers retransfers over false no-ops; reconnect is bounded | unit | `go test ./driver/ftp -count=1` | ✅ | ✅ green |
| 02-02-01 | 02 | 2 | FTPD-01, FTPD-02, FTPD-03, FTPD-04 | T-02-03 | Sync constructors only expose FTP through driver-backed paths, not unsupported fallbacks | unit | `go test ./sync -count=1` | ✅ | ✅ green |
| 02-02-02 | 02 | 2 | FTPD-05, FTPD-06, FTPD-07, FTPD-08, FTPD-09 | T-02-04 | Constructor wiring preserves explicit errors and conservative metadata behavior | unit | `go test ./driver/ftp ./sync -count=1` | ✅ | ✅ green |

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
- [x] Feedback latency < 45s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved after re-running `go test ./driver/ftp -count=1`, `go test ./sync -count=1`, `go test ./driver/ftp ./sync -count=1`, and `go test ./driver/... ./sync ./monitor -count=1` on 2026-04-24.
