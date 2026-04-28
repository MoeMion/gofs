---
phase: "05"
slug: public-ftpsyncservice-api-contract
status: approved
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-28
updated: 2026-04-28
source_state: reconstructed-from-plan-and-summary-artifacts
---

# Phase 05 — Validation Strategy

> Nyquist validation coverage reconstructed from Phase 05 PLAN/SUMMARY artifacts and re-audited against the current `ftpsync` implementation and test suite.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` via `go test` |
| **Config file** | `go.mod` |
| **Quick run command** | `go test ./ftpsync -count=1` |
| **Full suite command** | `go test ./... -run TestNonExistent -count=1` |
| **Estimated runtime** | ~8 seconds for focused package tests; <1 second for compile-only suite in this workspace |

---

## Sampling Rate

- **After every task commit:** Run `go test ./ftpsync -count=1`.
- **After every plan wave:** Run `go test ./... -run TestNonExistent -count=1` plus any plan-specific focused command below.
- **Before `/gsd-verify-work`:** Focused package tests and compile-only suite must be green.
- **Max feedback latency:** ~10 seconds for the Phase 05 focused package suite in this workspace.

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 05-01-01 | 01 | 1 | API-02 | T-05-01-02 | Public configuration is typed Go values only; no YAML/CLI/URL parser requirement. | unit | `go test ./ftpsync -run TestOptions -count=1` | ✅ `ftpsync/options_test.go` | ✅ green |
| 05-01-02 | 01 | 1 | API-01 | T-05-01-01, T-05-01-03 | `FTPSyncService` constructs from public options without leaking passwords or importing legacy runtime surfaces. | unit | `go test ./ftpsync -run 'Test(NewFTPSyncService|PackageDependencyBoundary)' -count=1` | ✅ `ftpsync/options_test.go`, `ftpsync/dependency_boundary_test.go` | ✅ green |
| 05-02-01 | 02 | 2 | API-04 | T-05-02-02 | Public errors are classifiable without string matching and preserve wrapped causes. | unit | `go test ./ftpsync -run 'Test(ErrorKind|ErrorWrapping|ContextCancellation)' -count=1` | ✅ `ftpsync/context_test.go` | ✅ green |
| 05-02-02 | 02 | 2 | API-03 | T-05-02-01, T-05-02-02 | Validation accepts only local→FTP and FTP→local, rejects unsupported/missing/ambiguous combinations, and omits FTP passwords from errors. | unit | `go test ./ftpsync -run TestValidate -count=1` | ✅ `ftpsync/validation_test.go` | ✅ green |
| 05-02-03 | 02 | 2 | API-04 | T-05-02-03, T-05-02-04 | `SyncOnce` / `StartBackground` validate before dispatch, classify cancellation, and enforce background direction policy. | unit | `go test ./ftpsync -run 'Test(SyncOnce|StartBackground|Context)' -count=1` | ✅ `ftpsync/context_test.go` | ✅ green |
| 05-03-01 | 03 | 3 | API-05 | T-05-03-01, T-05-03-03 | Hook contracts use library-local types and do not expose legacy logger/report/eventlog/web runtime types. | unit | `go test ./ftpsync -run TestHookContracts -count=1` | ✅ `ftpsync/hooks_test.go` | ✅ green |
| 05-03-02 | 03 | 3 | API-05 | T-05-03-01, T-05-03-03 | Omitted hooks normalize to no-ops; custom hooks receive typed callbacks without password leakage; dependency boundary remains enforced. | unit | `go test ./ftpsync -run 'Test(HookDefaults|CustomHooks|PackageDependencyBoundary)' -count=1` | ✅ `ftpsync/hooks_test.go`, `ftpsync/dependency_boundary_test.go` | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing Go test infrastructure covers all Phase 05 requirements. No Wave 0 scaffolding is required.

---

## Manual-Only Verifications

All Phase 05 behaviors have automated verification.

---

## Validation Audit 2026-04-28

| Metric | Count |
|--------|-------|
| Input state | B — no `*-VALIDATION.md`; SUMMARY artifacts present |
| Requirements audited | 5 |
| Task rows audited | 7 |
| Gaps found | 0 behavioral test gaps; 1 documentation artifact gap (`05-VALIDATION.md` missing) |
| Resolved | 1 documentation artifact reconstructed |
| Escalated | 0 |

### Commands Re-run During Audit

| Command | Result |
|---------|--------|
| `go test ./ftpsync -run 'Test(Options|NewFTPSyncService|PackageDependencyBoundary|Validate|ErrorKind|ErrorWrapping|ContextCancellation|SyncOnceChecksValidationAndContext|StartBackgroundChecksValidationAndContext|StartBackgroundRejectsFTPToLocal|StartBackgroundDirectionPolicyRemainsLocalToFTPOnly|ContextAwarePublicContractsCompile|HookContracts|HookDefaults|CustomHooks|PublicAPIStaysTypedOptionsOnly|VerificationCoverageChecklist)' -count=1` | ✅ passed (`ok ftpsync/ftpsync 3.754s`) |
| `go test ./ftpsync -count=1` | ✅ passed (`ok ftpsync/ftpsync 7.637s`) |
| `go list -deps ./ftpsync` | ✅ passed; dependency graph contains stdlib plus allowed `github.com/fsnotify/fsnotify`, `github.com/jlaffaye/ftp`, `golang.org/x/text`, and supporting transitive packages; no old `github.com/no-src/gofs/*` runtime surfaces detected |
| `go test ./... -run TestNonExistent -count=1` | ✅ passed (`ok ftpsync/ftpsync 0.006s [no tests to run]`) |

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies.
- [x] Sampling continuity: no 3 consecutive tasks without automated verify.
- [x] Wave 0 covers all MISSING references.
- [x] No watch-mode flags.
- [x] Feedback latency < 10s for focused Phase 05 package suite.
- [x] `nyquist_compliant: true` set in frontmatter.

**Approval:** approved 2026-04-28
