---
phase: 01
slug: ftp-endpoint-contract-routing
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-23
---

# Phase 01 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing Go test layout |
| **Quick run command** | `go test ./core ./sync ./monitor -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~60 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./core ./sync ./monitor -count=1`
- **After every plan wave:** Run `go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 60 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 01-01-01 | 01 | 1 | FTP-01, FTP-02, FTP-03, FTP-04 | T-01-01 | Malformed FTP endpoint strings fail predictably without silently reusing SSH semantics | unit | `go test ./core -count=1` | ✅ | ⬜ pending |
| 01-01-02 | 01 | 1 | FTP-01, FTP-02, FTP-03, FTP-04 | T-01-02 | FTP query fields round-trip through VFS parsing and flag/config surfaces | unit | `go test ./core -count=1` | ✅ | ⬜ pending |
| 01-02-01 | 02 | 2 | FTP-01, FTP-02 | T-01-03 | FTP endpoints enter FTP-specific sync routing instead of generic unsupported-path branching | unit | `go test ./sync -count=1` | ✅ | ⬜ pending |
| 01-02-02 | 02 | 2 | FTP-01 | T-01-04 | FTP source endpoints enter FTP-specific monitor routing instead of generic unsupported-path branching | unit | `go test ./monitor -count=1` | ✅ | ⬜ pending |

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
- [x] Feedback latency < 60s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** pending
