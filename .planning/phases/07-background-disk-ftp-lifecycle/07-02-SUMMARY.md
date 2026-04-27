---
phase: 07-background-disk-ftp-lifecycle
plan: 02
subsystem: ftpsync background lifecycle
tags: [ftpsync, background, fsnotify, debounce, lifecycle]
requires: [07-01]
provides: [recursive-fsnotify-watch, debounced-background-sync, non-terminal-sync-errors]
affects: [ftpsync/background.go, ftpsync/background_test.go]
tech-stack:
  added: [github.com/fsnotify/fsnotify]
  patterns: [recursive directory watch registration, resettable debounce timer, single buffered sync trigger, latest-error handle state]
key-files:
  created: []
  modified: [ftpsync/background.go, ftpsync/background_test.go]
decisions:
  - Background watcher events trigger full local→FTP sync passes instead of per-event FTP mutations to preserve coalescing and reuse Phase 6 transfer semantics.
  - Steady-state sync pass failures are recorded and hook-reported as latest handle errors while the watcher loop continues.
metrics:
  duration: 5min
  completed: 2026-04-27T08:54:29Z
  tasks: 2
  files: 2
---

# Phase 07 Plan 02: Background Watch and Debounce Summary

## One-liner

Recursive fsnotify background disk→FTP monitoring with debounced full-pass resyncs and observable non-terminal sync failures.

## What Changed

- Added recursive `fsnotify.Watcher` registration for the configured local source root.
- Added newly created directory registration so later nested writes are observed.
- Added dirty-event detection for create, write, remove, and rename fsnotify events.
- Added resettable debounce behavior with a single buffered trigger channel so bursts collapse into bounded sync passes.
- Reused `executeSyncOnce` for each coalesced background sync pass.
- Preserved long-running service semantics by recording sync pass failures on the handle, emitting hook-visible failure signals, and continuing to process later triggers.

## Tasks Completed

| Task | Name | Commit | Files |
|------|------|--------|-------|
| 1 RED | Add failing background debounce watcher coverage | d048ad8 | `ftpsync/background_test.go` |
| 1 GREEN | Debounce background filesystem sync triggers | 68332f9 | `ftpsync/background.go` |
| 2 RED | Cover non-terminal background sync failures | 82d020f | `ftpsync/background_test.go` |
| 2 GREEN | Keep background sync failures non-terminal | 28f22d1 | `ftpsync/background.go` |
| Rule 1 | Use real source roots in existing lifecycle tests | 536729e | `ftpsync/background_test.go` |

## Verification

- `go test ./ftpsync -run 'TestBackgroundDebouncesBurstEvents|TestBackgroundWatchesNewDirectories' -count=1` — passed after Task 1 implementation.
- `go test ./ftpsync -run 'TestBackgroundSyncFailureIsObservableAndNonTerminal' -count=1` — passed after Task 2 implementation.
- `go test ./ftpsync -run 'TestBackgroundDebouncesBurstEvents|TestBackgroundWatchesNewDirectories|TestBackgroundSyncFailureIsObservableAndNonTerminal' -count=1` — passed.
- `go test ./ftpsync -count=1` — passed.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Existing background lifecycle tests used a nonexistent source root after watcher setup became active**
- **Found during:** Overall verification after Task 2.
- **Issue:** `TestBackgroundHandleWait` and `TestStartBackgroundInitialSyncBeforeReady` constructed services with `/data/source`; once watcher registration became real, startup stored a watcher registration error and `Stop`/`Wait` returned it.
- **Fix:** Pointed those tests at `t.TempDir()` source roots while preserving their handle wait and initial readiness assertions.
- **Files modified:** `ftpsync/background_test.go`
- **Commit:** 536729e

## Threat Flags

| Flag | File | Description |
|------|------|-------------|
| threat_flag: filesystem-watch | `ftpsync/background.go` | Introduces a local filesystem watch surface that converts untrusted/noisy fsnotify events into network-affecting sync triggers. Covered by plan threat mitigations T-07-02-01 and T-07-02-02. |

## Known Stubs

None.

## TDD Gate Compliance

- RED commits present: d048ad8, 82d020f
- GREEN commits present after RED: 68332f9, 28f22d1

## Decisions Made

- Background change handling remains a coalesced full-resync model rather than per-event FTP mutation logic.
- Latest background sync errors remain observable through `Handle.Err()` but do not close `Done()` unless shutdown/cancellation occurs.

## Self-Check: PASSED

- Found `ftpsync/background.go`.
- Found `ftpsync/background_test.go`.
- Found commits d048ad8, 68332f9, 82d020f, 28f22d1, and 536729e in git history.
