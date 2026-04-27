---
phase: 08
slug: ftp-only-package-reduction
status: draft
nyquist_compliant: true
wave_0_complete: false
created: 2026-04-27
---

# Phase 08 — Validation Strategy

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | Go `testing` package |
| **Config file** | `go.mod` |
| **Quick run command** | `go test ./ftpsync -count=1` |
| **Full suite command** | `go test ./... -count=1` |
| **Estimated runtime** | ~60 seconds after dependency downloads |

## Sampling Rate

- **After every task commit:** Run `go test ./ftpsync -count=1`.
- **After every plan wave:** Run plan-specific verification plus `go test ./... -count=1` once package deletion/tidy permits it.
- **Before `/gsd-verify-work`:** Full suite, dependency blacklist, and module blacklist must be green.
- **Max feedback latency:** 60 seconds for package tests; module tidy may take longer on first dependency download.

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 08-01-01 | 01 | 1 | PRUNE-01, PRUNE-04 | T-08-01-01 | Dependency blacklist cannot be bypassed silently | unit/tooling | `go test ./ftpsync -run TestPackageDependencyBoundaryRejectsOldRuntime -count=1` | ❌ W0 | ⬜ pending |
| 08-01-02 | 01 | 1 | PRUNE-03 | T-08-01-02 | Public API does not expose old runtime types | unit/reflection | `go test ./ftpsync -run TestPublicAPIStaysTypedOptionsOnly -count=1` | ❌ W0 | ⬜ pending |
| 08-02-01 | 02 | 2 | PRUNE-02 | T-08-02-01 / T-08-02-03 | FTP helper preserves cwd safety and bounded retry | unit | `go test ./ftpsync -run 'TestSyncOnce(LocalToFTP|FTPToLocal|Partial|NeverWritesToCWD|Builds)' -count=1` | ⬜ existing tests | ⬜ pending |
| 08-02-02 | 02 | 2 | PRUNE-01, PRUNE-03 | T-08-02-02 | One-shot execution no longer imports old runtime packages | unit/tooling | `go test ./ftpsync -run 'Test(SyncOnce|StartBackground|PackageDependencyBoundaryRejectsOldRuntime|PublicAPIStaysTypedOptionsOnly)' -count=1` | ✅ | ⬜ pending |
| 08-03-01 | 03 | 3 | PRUNE-01, PRUNE-02 | T-08-03-01 | Deleted old packages do not break surviving library | tooling/unit | `go list ./... && go test ./ftpsync -count=1` | ✅ | ⬜ pending |
| 08-03-02 | 03 | 3 | PRUNE-01 | T-08-03-02 | Stale Docker/release surfaces cannot publish old CLI/server artifact | unit/tooling | `go test ./ftpsync -count=1` | ✅ | ⬜ pending |
| 08-04-01 | 04 | 4 | PRUNE-04 | T-08-04-01 | Module dependency set reflects library-only imports | tooling | `go mod tidy && go test ./... -count=1` | ✅ | ⬜ pending |
| 08-04-02 | 04 | 4 | PRUNE-01, PRUNE-04 | T-08-04-01 / T-08-04-02 | Final build graph excludes old runtime dependencies | tooling | `go test ./... -count=1 && go test ./ftpsync -run TestPackageDependencyBoundaryRejectsOldRuntime -count=1 && go list -deps ./...` | ✅ | ⬜ pending |

## Wave 0 Requirements

- [ ] `ftpsync/dependency_boundary_test.go` — created by Plan 01 Task 1.
- [ ] `ftpsync/public_api_test.go` — created by Plan 01 Task 2.

## Manual-Only Verifications

All Phase 8 behaviors have automated verification. Live FTP integration and final docs are Phase 9 scope.

## Validation Sign-Off

- [x] All tasks have `<automated>` verify.
- [x] Sampling continuity: no 3 consecutive tasks without automated verify.
- [x] Wave 0 covers all missing dependency/public API guard files.
- [x] No watch-mode flags.
- [x] Feedback latency target recorded.

**Approval:** draft 2026-04-27
