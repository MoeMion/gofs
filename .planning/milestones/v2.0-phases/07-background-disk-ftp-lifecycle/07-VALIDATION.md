---
phase: 07
slug: background-disk-ftp-lifecycle
status: passed
nyquist_compliant: true
wave_0_complete: true
created: 2026-04-27
validated: 2026-04-28
---

# Phase 07 — Validation Strategy

> Per-phase validation contract for feedback sampling during execution.

---

## Test Infrastructure

| Property | Value |
|----------|-------|
| **Framework** | go test |
| **Config file** | none — standard Go test discovery |
| **Quick run command** | `go test ./ftpsync -run 'Test(StartBackground|BackgroundHandle|BackgroundLifecycle)' -count=1` |
| **Full suite command** | `go test ./ftpsync -count=1` |
| **Estimated runtime** | ~15 seconds |

---

## Sampling Rate

- **After every task commit:** Run `go test ./ftpsync -run 'Test(StartBackground|BackgroundHandle|BackgroundLifecycle)' -count=1`
- **After every plan wave:** Run `go test ./ftpsync -count=1`
- **Before `/gsd-verify-work`:** Full suite must be green
- **Max feedback latency:** 15 seconds

---

## Per-Task Verification Map

| Task ID | Plan | Wave | Requirement | Threat Ref | Secure Behavior | Test Type | Automated Command | File Exists | Status |
|---------|------|------|-------------|------------|-----------------|-----------|-------------------|-------------|--------|
| 07-01-01 | 01 | 1 | WATCH-01, WATCH-03, WATCH-05 | T-07-01-01 / T-07-01-02 | local→FTP only; startup validates and exposes handle lifecycle | unit | `go test ./ftpsync -run 'Test(StartBackgroundChecksValidationAndContext|StartBackgroundRejectsFTPToLocal|BackgroundHandleWait)' -count=1` | ✅ | ✅ green |
| 07-01-02 | 01 | 1 | WATCH-01, WATCH-03 | T-07-01-03 | initial sync runs before ready; handle reports readiness and exit | unit | `go test ./ftpsync -run 'TestStartBackgroundInitialSyncBeforeReady' -count=1` | ✅ | ✅ green |
| 07-02-01 | 02 | 2 | WATCH-02 | T-07-02-01 / T-07-02-02 | fsnotify burst coalesces into bounded sync passes | unit | `go test ./ftpsync -run 'TestBackgroundDebouncesBurstEvents|TestBackgroundWatchesNewDirectories' -count=1` | ✅ | ✅ green |
| 07-02-02 | 02 | 2 | WATCH-02, WATCH-03 | T-07-02-03 | sync failures surface but do not terminate the run | unit | `go test ./ftpsync -run 'TestBackgroundSyncFailureIsObservableAndNonTerminal' -count=1` | ✅ | ✅ green |
| 07-03-01 | 03 | 3 | WATCH-04 | T-07-03-01 / T-07-03-02 | cancellation and stop close watcher, timers, retries, goroutines, and done channel deterministically | unit | `go test ./ftpsync -run 'TestBackgroundStopShutsDownDeterministically|TestBackgroundContextCancelStopsRunner' -count=1` | ✅ | ✅ green |
| 07-03-02 | 03 | 3 | WATCH-04, WATCH-05 | T-07-03-03 | unsupported background directions stay blocked while local→FTP shutdown remains deterministic | unit | `go test ./ftpsync -run 'TestStartBackgroundRejectsFTPToLocal|TestBackgroundWaitReturnsFinalError' -count=1` | ✅ | ✅ green |

*Status: ⬜ pending · ✅ green · ❌ red · ⚠️ flaky*

---

## Wave 0 Requirements

Existing infrastructure covers all phase requirements.

---

## Manual-Only Verifications

All phase behaviors have automated verification.

---

## Validation Audit 2026-04-28

| Metric | Count |
|--------|-------|
| Input state | A — existing VALIDATION.md audited |
| Gaps found | 0 |
| Resolved | 0 |
| Escalated | 0 |
| Task rows updated from pending to green | 6 |

### Evidence Cross-Check

| Evidence Source | Result |
|-----------------|--------|
| `07-01-SUMMARY.md` | Confirms local→FTP `StartBackground`, waitable handle, FTP→local rejection, initial catch-up, and passing plan tests. |
| `07-02-SUMMARY.md` | Confirms recursive fsnotify watching, debounce/coalescing, non-terminal failure handling, and passing plan tests. |
| `07-03-SUMMARY.md` | Confirms deterministic stop/cancel shutdown, final `Wait()` status semantics, unsupported direction coverage, and passing plan tests. |
| `07-VERIFICATION.md` | Status `passed`, score `6/6 must-haves verified`, no remaining gaps; includes extra evidence for active-sync follow-up and negative FTP timeout validation. |

No true coverage gaps were found during this audit. The previous `status: draft` frontmatter and `⬜ pending` task rows were stale relative to the completed SUMMARY and VERIFICATION artifacts plus green automated test runs.

---

## Validation Sign-Off

- [x] All tasks have `<automated>` verify or Wave 0 dependencies
- [x] Sampling continuity: no 3 consecutive tasks without automated verify
- [x] Wave 0 covers all MISSING references
- [x] No watch-mode flags
- [x] Feedback latency < 15s
- [x] `nyquist_compliant: true` set in frontmatter

**Approval:** approved 2026-04-27
