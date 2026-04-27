---
phase: 05-public-ftpsyncservice-api-contract
verified: 2026-04-27T04:02:57Z
status: passed
score: 5/5 must-haves verified
overrides_applied: 0
---

# Phase 05: Public FTPSyncService API Contract Verification Report

**Phase Goal:** Developers can construct and validate a focused `ftpsync.FTPSyncService` using typed Go options without invoking CLI, server, daemon, or global reporting runtime code.
**Verified:** 2026-04-27T04:02:57Z
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
| --- | --- | --- | --- |
| 1 | Developer can import package `ftpsync` and construct an `FTPSyncService` from typed Go options only. | ✓ VERIFIED | `ftpsync/options_test.go:93-102` constructs via `NewFTPSyncService`; `ftpsync/service.go:28-37` accepts `Options`; `go test ./ftpsync -count=1` passed; `go list -deps ./ftpsync` output contains only stdlib + `github.com/no-src/gofs/ftpsync`. |
| 2 | Developer can configure local path, FTP host/port/credentials/remote path, passive mode, timeout, path encoding, retry behavior, and ignore rules without YAML, CLI flags, or URL parser requirements. | ✓ VERIFIED | `ftpsync/options.go:15-68` defines typed `Endpoint`, `FTPOptions`, `RetryOptions`, `IgnoreRule`, `Options`; `ftpsync/options_test.go:12-64` verifies all required fields are expressible as Go values. |
| 3 | Developer receives explicit validation failures for unsupported or ambiguous endpoint combinations before transfer work starts. | ✓ VERIFIED | `ftpsync/service.go:39-140` validates endpoint shape, direction, host/remote path/port, nil context; `ftpsync/validation_test.go:32-76` covers local↔local, FTP↔FTP, non-FTP, missing path/host/remote path, invalid port, ambiguous combinations, and password-safe errors. |
| 4 | Developer can pass `context.Context` to public sync methods and distinguish validation, cancellation, connection/authentication, transfer, and unsupported-capability errors. | ✓ VERIFIED | `ftpsync/service.go:47-67,132-139` exposes `SyncOnce(ctx)` / `StartBackground(ctx)` and classifies cancellation; `ftpsync/errors.go:5-68` exports `ErrorKind`, `Error`, `IsKind`; `ftpsync/context_test.go:12-144` verifies `errors.Is`, `IsKind`, and public contract compilation. |
| 5 | Developer can attach optional logging, progress, and sync event hooks, while the default service remains no-op and library-local. | ✓ VERIFIED | `ftpsync/hooks.go:3-49` exports `Logger`, `LoggerFunc`, `Progress`, `ProgressHook`, `SyncEvent`, `EventHook`, `HookSet`; `ftpsync/service.go:168-204` normalizes nil hooks to no-op handlers; `ftpsync/hooks_test.go:57-120` verifies zero-value normalization and custom callback dispatch. |

**Score:** 5/5 truths verified

### Required Artifacts

| Artifact | Expected | Status | Details |
| --- | --- | --- | --- |
| `ftpsync/doc.go` | Package declaration and public package documentation | ✓ VERIFIED | Exists; `package ftpsync` present at `ftpsync/doc.go:1-4`; provides focused API boundary docs. |
| `ftpsync/options.go` | Typed public option structs and direction constants | ✓ VERIFIED | Exists; exports `Endpoint`, `FTPOptions`, `Options`, `Direction`, `RetryOptions`, `IgnoreRule`, `IgnoreKind`; no legacy runtime imports. |
| `ftpsync/service.go` | `FTPSyncService` type, constructor, validation, context-aware methods, hook normalization | ✓ VERIFIED | Exists; substantive implementation across constructor, validation, context checking, hook normalization, and helper dispatch methods. |
| `ftpsync/errors.go` | Public error kinds and classification helpers | ✓ VERIFIED | Exists; exports `ErrorKind`, `Error`, `IsKind`, error-kind constants, `Unwrap`, `Kind`. |
| `ftpsync/hooks.go` | Public hook function/types and no-op defaults | ✓ VERIFIED | Exists; exports logger/progress/event contracts with library-local `noopLogger`. |
| `ftpsync/options_test.go` | Import, construction, and typed option coverage | ✓ VERIFIED | Exists; external-package tests verify typed configuration and dependency boundary. |
| `ftpsync/validation_test.go` | Unsupported combination and missing/ambiguous field coverage | ✓ VERIFIED | Exists; substantive table-driven validation coverage. |
| `ftpsync/context_test.go` | Context cancellation and error classification coverage | ✓ VERIFIED | Exists; verifies cancellation ordering and typed error behavior. |
| `ftpsync/hooks_test.go` | Hook default/custom callback coverage | ✓ VERIFIED | Exists; verifies no-op defaults, custom dispatch, and dependency boundary for hooks. |

### Key Link Verification

| From | To | Via | Status | Details |
| --- | --- | --- | --- | --- |
| `ftpsync/service.go` | `ftpsync/options.go` | `NewFTPSyncService` accepts `Options` and stores a sanitized service copy | ✓ WIRED | `gsd-tools verify key-links` passed; `ftpsync/service.go:28-37,158-166` copies and stores `Options`. |
| `ftpsync/options_test.go` | `ftpsync` | external-style package test imports public API only | ✓ WIRED | `ftpsync/options_test.go:1-10` uses `package ftpsync_test` and imports `github.com/no-src/gofs/ftpsync`. |
| `ftpsync/service.go` | `ftpsync/errors.go` | validation and method contracts return typed `*Error` values | ✓ WIRED | `ftpsync/service.go:42,88,96,100,113,116,119,125,129,134,137` call `newError(...)`; `errors.go` defines typed wrapper and classifier. |
| `ftpsync/context_test.go` | `ftpsync/service.go` | context cancellation checked before transfer capability dispatch | ✓ WIRED | `ftpsync/context_test.go:64-116` asserts validation then cancellation then unsupported-capability behavior. |
| `ftpsync/service.go` | `ftpsync/hooks.go` | constructor normalizes nil hooks to no-op functions stored on service | ✓ WIRED | `ftpsync/service.go:31,168-179,202-204` uses `HookSet` and `normalizeHooks`. |
| `ftpsync/hooks_test.go` | `ftpsync/service.go` | custom hooks invoked only by explicit service helper paths | ✓ WIRED | `ftpsync/hooks_test.go:80-120` invokes `svc.log`, `svc.reportProgress`, `svc.reportEvent` and observes callbacks. |

### Data-Flow Trace (Level 4)

| Artifact | Data Variable | Source | Produces Real Data | Status |
| --- | --- | --- | --- | --- |
| N/A | N/A | N/A | N/A | 此阶段交付的是库 API、校验和 hook 合约，不是渲染动态数据的 UI/页面组件；Level 4 数据流追踪不适用。 |

### Behavioral Spot-Checks

| Behavior | Command | Result | Status |
| --- | --- | --- | --- |
| `ftpsync` package contract compiles and tests pass | `go test ./ftpsync -count=1` | `ok github.com/no-src/gofs/ftpsync` | ✓ PASS |
| Public API focused verification tests pass | `go test ./ftpsync -run 'Test(NewFTPSyncServiceConstructsFromTypedOptions|TestValidateAcceptsSupportedDirections|TestSyncOnceChecksValidationAndContext|TestHookDefaultsAreNoOp|TestPackageDependencyBoundary|TestPackageDependencyBoundaryForHooks)' -count=1` | passed | ✓ PASS |
| Package dependency graph excludes legacy runtime surfaces | `go list -deps ./ftpsync` | Output contained stdlib packages plus `github.com/no-src/gofs/ftpsync` only; no `cmd/conf/flag/server/daemon/api/monitor/report/logger/driver/sftp/driver/minio` entries | ✓ PASS |
| Module-wide compile sanity without runtime startup | `go test ./... -run TestNonExistent -count=1` | all packages compiled successfully | ✓ PASS |

### Requirements Coverage

| Requirement | Source Plan | Description | Status | Evidence |
| --- | --- | --- | --- | --- |
| API-01 | 05-01 | Developer can import a `ftpsync` package and construct an `FTPSyncService` without invoking CLI, process, server, or daemon code | ✓ SATISFIED | `ftpsync/options_test.go:93-146`; `go list -deps ./ftpsync` excludes legacy runtime packages; `ftpsync/service.go:28-37` constructs service only. |
| API-02 | 05-01 | Developer can configure `FTPSyncService` through typed Go options covering local path, FTP host, port, username, password, remote path, passive mode, timeout, path encoding, retry behavior, and ignore rules | ✓ SATISFIED | `ftpsync/options.go:15-68`; `ftpsync/options_test.go:12-64`. |
| API-03 | 05-02 | `FTPSyncService` validates supported directions explicitly and rejects unsupported local↔local, FTP↔FTP, non-FTP, missing-path, and ambiguous endpoint combinations | ✓ SATISFIED | `ftpsync/service.go:69-126`; `ftpsync/validation_test.go:32-76`. |
| API-04 | 05-02 | Public sync methods accept `context.Context` and return structured errors for validation, cancellation, connection/authentication, transfer, and unsupported-capability failures | ✓ SATISFIED | `ftpsync/service.go:47-67,132-139`; `ftpsync/errors.go:5-68`; `ftpsync/context_test.go:12-144`. |
| API-05 | 05-03 | Public API exposes optional no-op-by-default hooks for logging, progress, and sync event reporting without requiring global loggers or web reports | ✓ SATISFIED | `ftpsync/hooks.go:3-49`; `ftpsync/service.go:168-204`; `ftpsync/hooks_test.go:42-143`. |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
| --- | --- | --- | --- | --- |
| `ftpsync/service.go` | 55 | `SyncOnce` currently returns `ErrUnsupportedCapability` after validation/context checks | ℹ️ Info | 不是本阶段阻塞项；Phase 5 只要求 API 合约、校验与错误分类，真正一次性同步实现明确属于 Phase 6。 |
| `ftpsync/service.go` | 66 | `StartBackground` currently returns `ErrUnsupportedCapability` after validation/context checks | ℹ️ Info | 不是本阶段阻塞项；后台生命周期实现明确属于 Phase 7，本阶段只需保留公共方法签名与错误语义。 |

### Human Verification Required

无。

此阶段产物是静态 Go API 合约、验证逻辑、错误分类和 hook 契约；相关目标可通过源码检查、依赖边界检查和自动化测试充分验证，不需要额外人工 UI/交互验证。

### Gaps Summary

未发现阻塞本阶段目标达成的缺口。

验证结论：Phase 05 已经实际交付了一个可导入、可构造、可校验的 `ftpsync.FTPSyncService` 公共 API，且其依赖边界不触发旧 CLI / server / daemon / report 运行时代码。虽然 `SyncOnce` 与 `StartBackground` 仍未实现具体传输行为，但这与路线图分工一致，属于后续 Phase 6/7 的范围，不构成 Phase 05 目标失败。

---

_Verified: 2026-04-27T04:02:57Z_
_Verifier: the agent (gsd-verifier)_
