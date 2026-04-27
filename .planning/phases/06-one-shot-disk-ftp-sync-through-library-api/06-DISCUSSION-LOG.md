# Phase 6: One-Shot Disk↔FTP Sync Through Library API - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-27
**Phase:** 06-one-shot-disk-ftp-sync-through-library-api
**Areas discussed:** 结果返回粒度, 失败处理方式, 同步策略暴露, 本地目录语义, 部分成功返回形式

---

## 结果返回粒度

| Option | Description | Selected |
|--------|-------------|----------|
| Summary only | 返回方向、是否变更、总文件数、增改删数量、耗时、首个错误；先把 API 保持小而稳 | ✓ |
| Summary + failures | 在汇总基础上，再返回失败文件列表和失败原因，便于调用方做日志/告警 | |
| Full per-file report | 每个文件都返回详细动作和结果，能力最强，但 API 和内存开销都会更大 | |

**User's choice:** Summary only
**Notes:** Phase 6 优先保持 `Result` 小而稳，不把 per-file report 直接做进 public API。

---

## 失败处理方式

| Option | Description | Selected |
|--------|-------------|----------|
| Fail fast | 第一个关键错误就停止，返回错误和当前 summary；行为最清晰，也最容易和现有 sync 语义对齐 | |
| Best effort | 尽量继续处理其他文件，最后返回一个聚合错误和 summary | ✓ |
| By option later | Phase 6 先固定一种，后续再考虑通过 option 暴露策略切换 | |

**User's choice:** Best effort
**Notes:** 允许部分文件成功、部分文件失败；返回值需要表达 partial success。

---

## 同步策略暴露

| Option | Description | Selected |
|--------|-------------|----------|
| Fixed v1 semantics | 先直接沿用当前 FTP v1 语义，不新增 delete/overwrite/compare policy options；API 最小 | ✓ |
| Expose delete only | 只把最关键的删除策略做成显式 option，其他仍沿用默认语义 | |
| Expose small policy set | 公开少量策略项，例如 delete、overwrite、compare mode，但先保持范围克制 | |

**User's choice:** Fixed v1 semantics
**Notes:** Phase 6 不扩 public policy surface，先把既有一键同步行为库化。

---

## 本地目录语义

| Option | Description | Selected |
|--------|-------------|----------|
| Auto-create target | 目标目录不存在时自动创建；但仍必须要求显式 LocalPath，绝不允许回退到 cwd | ✓ |
| Must exist first | 要求调用方先创建目录，库只在已有目录下落盘 | |
| Mixed rules | 目录根必须存在，但子目录可自动创建 | |

**User's choice:** Auto-create target
**Notes:** `FTP→disk` 目标根目录可自动创建，但必须是显式配置的目标路径，cwd 安全边界不能退让。

---

## 部分成功返回形式

| Option | Description | Selected |
|--------|-------------|----------|
| Result + error | 返回 summary，同时 error 非空；最符合 Go 习惯，也能表达部分成功 | ✓ |
| Result only | 只在 Result 里放失败计数/状态，error 只用于完全不可执行的情况 | |
| Error only | 只返回 error，调用方自己解析错误，不推荐 | |

**User's choice:** Result + error
**Notes:** Best effort 模式下 summary 仍然必须返回；只要存在文件级失败，`error` 就非空。

---

## the agent's Discretion

- `Result` 的精确字段名和最小字段集合。
- 部分失败错误类型的具体结构与命名。
- 内部如何桥接 `ftpsync` 到现有 `sync/ftp_push_client_sync.go` / `sync/ftp_pull_client_sync.go`。

## Deferred Ideas

- 在 public API 中公开 delete / overwrite / compare policy options。
- 提供 full per-file report。
- 讨论后台 `FTP→disk` 轮询行为。
