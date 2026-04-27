---
phase: 04
slug: ftp-verification-discoverability
status: draft
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-24
---

# Phase 04 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — existing Go test + integration harness |
| **Quick run command** | `go test ./integration -tags=integration_test_ftp -run TestIntegration_FTP -count=1` |
| **Full suite command** | `go test ./integration -count=1 -tags=integration_test_ftp && go test ./... -count=1` |
| **Estimated runtime** | ~45 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./integration -tags=integration_test_ftp -run TestIntegration_FTP -count=1`
- **After every plan wave:** Run `go test ./integration -count=1 -tags=integration_test_ftp && go test ./... -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 45 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 04-01-01 | 01 | 1 | TEST-01 / TEST-02 / TEST-03 | T-04-01 | FTP test server starts from repo-owned script with passive-only plain FTP and deterministic workspace | script/integration | `bash scripts/ftp/init-ftp.sh && go test ./integration -tags=integration_test_ftp -run TestIntegration_FTP -count=1` | ✅ | ⬜ pending |
| 04-01-02 | 01 | 1 | TEST-01 / TEST-02 / TEST-03 | T-04-02 | FTP YAML fixtures prove nested paths and delete/rename-relevant flows through real protocol operations | integration | `go test ./integration -tags=integration_test_ftp -run TestIntegration_FTP -count=1` | ✅ | ⬜ pending |
| 04-02-01 | 02 | 2 | TEST-01 / TEST-02 / TEST-03 | T-04-03 | Tagged FTP integration suite executes real disk↔FTP flows instead of fake-only checks | integration | `go test -v -race -tags=integration_test_ftp ./integration -run TestIntegration_FTP -count=1` | ✅ | ⬜ pending |
| 04-02-02 | 02 | 2 | TEST-01 / TEST-02 / TEST-03 | T-04-04 | CI provisions FTP and runs the same tagged suite on Ubuntu | ci | `go test -v -race -tags=integration_test_ftp ./integration -run TestIntegration_FTP -count=1` | ✅ | ⬜ pending |
| 04-03-01 | 03 | 1 | DOC-01 | T-04-05 | README examples use real FTP endpoint grammar with push and pull coverage | docs | `go test ./core -run TestVFSParseFTP -count=1` | ✅ | ⬜ pending |
| 04-03-02 | 03 | 1 | DOC-02 | T-04-06 | Docs explicitly state plain FTP only, passive mode only, and no FTP↔FTP sync | docs | `go test ./core -run TestVFSParseFTP -count=1` | ✅ | ⬜ pending |

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
