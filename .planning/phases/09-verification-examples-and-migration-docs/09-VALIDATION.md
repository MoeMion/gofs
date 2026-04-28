---
phase: 09
slug: verification-examples-and-migration-docs
status: approved
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-28
updated: 2026-04-28
state: B-reconstructed-from-artifacts
---

# Phase 09 — Validation Strategy

> Nyquist validation contract reconstructed from Phase 09 PLAN, SUMMARY, VERIFICATION, UAT, README/MIGRATION, real FTP integration, and compiler-checked example evidence.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` package |
| **Config file** | `go.mod` |
| **Quick run command** | `go test ./ftpsync -run 'TestVerificationCoverageChecklist|TestIntegrationRealFTP|ExampleFTPSyncService' -count=1` |
| **Full suite command** | `go test ./...` |
| **Estimated runtime** | ~10 seconds |

---

## Sampling Rate

- **After every task commit:** Run the task-specific `<automated>` command from the plan.
- **After every plan wave:** Run `go test ./...`.
- **Before `/gsd-verify-work`:** Full suite must be green.
- **Max feedback latency:** ~10 seconds for focused Phase 09 checks; full suite is currently cached/fast for the reduced module.

---

## Phase Artifact Evidence

| Artifact | Evidence Used |
|----------|---------------|
| `09-01-SUMMARY.md` | Module contract changed to `module ftpsync`; external tests import `ftpsync/ftpsync`; coverage checklist and dependency boundary guards added. |
| `09-02-SUMMARY.md` | Real FTP local→FTP and FTP→local integration tests added through public `FTPSyncService.SyncOnce` and loopback fixture. |
| `09-03-SUMMARY.md` | Compiler-checked examples and package docs added for one-shot push, one-shot pull, and background local→FTP usage. |
| `09-04-SUMMARY.md` | README and MIGRATION rewritten for typed-option adoption and removed/unsupported v2.0 surfaces. |
| `09-VERIFICATION.md` | All six Phase 09 requirements marked PASS with nine automated checks green. |
| `09-UAT.md` | Four UAT scenarios complete: module contract, real FTP integration, examples, docs/migration. |
| `README.md` / `MIGRATION.md` | Supported import path `import "ftpsync/ftpsync"`, typed options, limitations, and migration from old runtime documented. |
| `ftpsync/real_ftp_integration_test.go` | Go-native loopback-only FTP fixture covers real local→FTP and FTP→local one-shot flows through public API. |
| `ftpsync/examples_test.go` | Go Example tests compile consumer-facing usage for all supported public modes. |

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 09-01-01 | 01 | 1 | VERIFY-01, VERIFY-03 | T-09-01-01 | Module path and external tests use supported local package import `ftpsync/ftpsync`. | unit | `go test ./ftpsync -run 'TestValidateAcceptsSupportedDirections|TestContextAwarePublicContractsCompile' -count=1` | ✅ | ✅ green |
| 09-01-02 | 01 | 1 | VERIFY-01, VERIFY-03 | T-09-01-04 | Executable checklist keeps public construction, validation, one-shot, background, cwd safety, encoding, passive mode, cancellation, and shutdown coverage discoverable. | unit | `go test ./ftpsync -run TestVerificationCoverageChecklist -count=1` | ✅ | ✅ green |
| 09-01-03 | 01 | 1 | VERIFY-03 | T-09-01-02 | Reduced module exposes only `ftpsync/ftpsync` and rejects stale runtime/non-FTP dependencies. | unit | `go test ./ftpsync -run 'TestPackageDependencyBoundaryRejectsOldRuntime' -count=1` | ✅ | ✅ green |
| 09-02-01 | 02 | 2 | VERIFY-02, VERIFY-03 | T-09-02-05 | FTP server dependency remains test-only and discoverable as a module dependency. | integration | `go list -m github.com/fclairamb/ftpserverlib` | ✅ | ✅ green |
| 09-02-02 | 02 | 2 | VERIFY-02 | T-09-02-01 / T-09-02-02 | Public `FTPSyncService.SyncOnce` uploads local nested files to a loopback real FTP server. | integration | `go test ./ftpsync -run TestIntegrationRealFTPLocalToFTP -count=1` | ✅ | ✅ green |
| 09-02-03 | 02 | 2 | VERIFY-02 | T-09-02-01 / T-09-02-02 | Public `FTPSyncService.SyncOnce` downloads nested files from a loopback real FTP server to the configured local destination. | integration | `go test ./ftpsync -run 'TestIntegrationRealFTP(LocalToFTP|ToLocal)' -count=1` | ✅ | ✅ green |
| 09-03-01 | 03 | 2 | DOC-01 | T-09-03-01 / T-09-03-03 | One-shot local→FTP and FTP→local examples compile as external consumer examples with typed options. | docs/example | `go test ./ftpsync -run 'ExampleFTPSyncService_SyncOnce' -count=1` | ✅ | ✅ green |
| 09-03-02 | 03 | 2 | DOC-01 | T-09-03-03 | Background local→FTP example compiles and uses supported lifecycle cleanup. | docs/example | `go test ./ftpsync -run 'ExampleFTPSyncService' -count=1` | ✅ | ✅ green |
| 09-03-03 | 03 | 2 | DOC-01 | T-09-03-04 | Package docs match supported public API and local import path. | docs/example | `go test ./ftpsync -run 'ExampleFTPSyncService|TestContextAwarePublicContractsCompile' -count=1` | ✅ | ✅ green |
| 09-04-01 | 04 | 2 | DOC-01, DOC-02, DOC-03 | T-09-04-01 / T-09-04-02 | README presents only supported local library contract and typed-option usage modes. | docs | `go test ./ftpsync -run 'ExampleFTPSyncService' -count=1` | ✅ | ✅ green |
| 09-04-02 | 04 | 2 | DOC-02, DOC-03 | T-09-04-03 | MIGRATION documents breaking package invocation, typed configuration, and removed/unsupported surfaces. | docs | `go test ./...` | ✅ | ✅ green |
| 09-04-03 | 04 | 2 | DOC-01, DOC-02, DOC-03 | T-09-04-01 / T-09-04-04 | README and MIGRATION are internally consistent, link each other, and avoid stale final consumer import paths. | docs | `go test ./... && go test ./ftpsync -run 'ExampleFTPSyncService' -count=1` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Requirement Coverage Matrix

| Requirement | Coverage Status | Behavioral Evidence | Automated Command |
|-------------|-----------------|---------------------|-------------------|
| VERIFY-01 | COVERED | `TestVerificationCoverageChecklist` checks construction, validation, one-shot push/pull, and background lifecycle coverage targets. | `go test ./ftpsync -run TestVerificationCoverageChecklist -count=1` |
| VERIFY-02 | COVERED | `TestIntegrationRealFTPLocalToFTP` and `TestIntegrationRealFTPToLocal` exercise real loopback FTP via public `FTPSyncService.SyncOnce` without CLI/runtime code. | `go test ./ftpsync -run 'TestIntegrationRealFTP' -count=1` |
| VERIFY-03 | COVERED | Checklist preserves cwd safety, path encoding, passive mode, cancellation, dependency boundary, and background shutdown tests. | `go test ./ftpsync -run 'TestVerificationCoverageChecklist|TestPackageDependencyBoundaryRejectsOldRuntime' -count=1` |
| DOC-01 | COVERED | README, package docs, and Go examples show typed one-shot push, one-shot pull, and background local→FTP usage. | `go test ./ftpsync -run 'ExampleFTPSyncService' -count=1` |
| DOC-02 | COVERED | README and MIGRATION state removed/unsupported v2.0 surfaces and limitations. | `go test ./...` plus documentation audit |
| DOC-03 | COVERED | MIGRATION explains CLI→package shift and supported import path; README links to MIGRATION. | `go test ./...` plus documentation audit |

---

## Wave 0 Requirements

Existing infrastructure covers all Phase 09 requirements.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Audit 2026-04-28

| Metric | Count |
|--------|-------|
| Input state | B — `VALIDATION.md` absent, `SUMMARY.md` files present |
| Requirements audited | 6 |
| Task verification rows | 12 |
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
| Manual-only | 0 |

### Commands Re-run During Audit

| Command | Result |
|---------|--------|
| `go test ./ftpsync -run 'TestValidateAcceptsSupportedDirections|TestContextAwarePublicContractsCompile' -count=1` | PASS |
| `go test ./ftpsync -run TestVerificationCoverageChecklist -count=1` | PASS |
| `go test ./ftpsync -run 'TestPackageDependencyBoundaryRejectsOldRuntime' -count=1` | PASS |
| `go list ./...` | PASS — `ftpsync/ftpsync` |
| `go test ./ftpsync -run 'TestIntegrationRealFTP' -count=1` | PASS |
| `go test ./ftpsync -run 'ExampleFTPSyncService' -count=1` | PASS |
| `go test ./...` | PASS |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 10s for focused checks
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-04-28
