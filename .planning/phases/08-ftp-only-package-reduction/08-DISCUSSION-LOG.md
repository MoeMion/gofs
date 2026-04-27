# Phase 8: FTP-Only Package Reduction - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-27
**Phase:** 08-FTP-Only Package Reduction
**Areas discussed:** 削减策略, 保留边界, 验证标准, 发布边界, 模块路径, 旧代码处理

---

## 削减策略

| Option | Description | Selected |
|--------|-------------|----------|
| 最小内联核心 | 把 `ftpsync` 需要的 disk<->FTP 操作、VFS/ignore/retry 适配最小化迁入或包内化，避免继续导入旧 `sync/core/logger` 链路。 | ✓ |
| 重构旧包 | 拆分 `sync/core/driver` 等旧包，让 FTP-only 路径不再导入 server/SFTP/MinIO。 | |
| Build tags 隔离 | 用 build tags 屏蔽 CLI/server/非 FTP 后端。 | |

**User's choice:** 最小内联核心
**Notes:** 目标是构建图收缩确定，不走大范围旧架构重构，也不要求用户使用 build tags。

---

## 保留边界

| Option | Description | Selected |
|--------|-------------|----------|
| 全部保留 | 保留 FTP driver 行为、path encoding、retry、ignore/filtering、rate limiting、cwd safety、best-effort 结果语义和 background debounce/shutdown 语义。 | ✓ |
| 只保留公共行为 | 只保证 `FTPSyncService` API 和测试可见语义，内部实现可换。 | |
| 允许简化 internals | 允许更小实现替换部分 helper，但用测试证明合同不破坏。 | |

**User's choice:** 全部保留
**Notes:** Phase 8 只改依赖边界，不改用户可见行为或已验证 FTP 行为。

---

## 验证标准

| Option | Description | Selected |
|--------|-------------|----------|
| 依赖黑名单+测试 | 用 `go list -deps` 明确禁止旧 runtime/protocol deps，同时测试通过。 | ✓ |
| go.mod 精简为准 | 以 `go mod tidy` 删除依赖为主要成功标准。 | |
| 只测 ftpsync 编译 | 只要求 `go test ./ftpsync` 不拉旧依赖。 | |

**User's choice:** 依赖黑名单+测试
**Notes:** `go.mod` 清理仍要做，但不是唯一证明；需要显式黑名单和测试。

---

## 发布边界

| Option | Description | Selected |
|--------|-------------|----------|
| 保持当前路径 | 保持当前 module/package，只收缩 `ftpsync` 构建图。 | |
| 准备子模块 | 为未来独立 module 做准备。 | |
| 立即独立模块 | 把 `ftpsync` 抽成独立 module。 | ✓ |

**User's choice:** 立即独立模块
**Notes:** 用户希望 Phase 8 直接进入独立库形态，而不是仅准备未来发布。

---

## 模块路径

| Option | Description | Selected |
|--------|-------------|----------|
| 子目录模块 | 在 `ftpsync/` 内新增独立 `go.mod`，路径为 `github.com/no-src/gofs/ftpsync`。 | |
| 仓库根改库 | 把根 `go.mod` 改成只服务 FTP sync library。 | ✓ |
| 新仓库路径预留 | 代码准备为独立库，但最终路径后续决定。 | |

**User's choice:** 仓库根改库
**Notes:** Phase 8 要把 repository root module 收敛成 FTP sync library，而不是添加 nested module。

---

## 旧代码处理

| Option | Description | Selected |
|--------|-------------|----------|
| 删除旧运行时代码 | 删除或迁出 CLI/server/gRPC/HTTP/SFTP/MinIO/task/auth/daemon 等旧 runtime 代码。 | ✓ |
| 目录隔离但保留 | 保留旧代码，但通过 tags 或目录隔离。 | |
| 只断依赖不删除 | 不主动删除旧代码，只重写 `ftpsync` 依赖链。 | |

**User's choice:** 删除旧运行时代码
**Notes:** 这是 Phase 8 的关键范围扩大点：planner 应允许删除旧 runtime 文件和依赖，只保留必要 FTP library internals。

---

## the agent's Discretion

- 具体包名、文件移动顺序、最小 helper 形状由 researcher/planner 决定。
- 需显式处理最终 import/package shape，供 Phase 9 文档承接。

## Deferred Ideas

- Phase 9 负责最终 README、examples、release/migration 文档。
- FTPS、FTP->disk background polling、FTP<->FTP、bidirectional sync、legacy YAML/CLI parsing 继续延期。
