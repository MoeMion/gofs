# Phase 2: FTP Driver Backend - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-23
**Phase:** 2-FTP Driver Backend
**Areas discussed:** 文件状态比较, 连接与重连策略, 目录与删除语义, 时间戳精度策略

---

## 文件状态比较

| Option | Description | Selected |
|--------|-------------|----------|
| 保守比较 | 以 size + mtime 为主，但当 FTP 元数据可疑或不完整时，宁可多传也不冒漏传风险；更符合文件同步工具的安全取向 | ✓ |
| 激进跳过传输 | 只要 size + mtime 看起来相同就尽量跳过；性能更好，但误判漏传风险更高 | |
| 强制更多校验 | 更频繁地走 checksum 或额外校验；更稳，但会明显扩大 v1 复杂度和成本 | |
| You decide | 你按最小正确原则定 | |

**User's choice:** 保守比较
**Notes:** 用户明确接受“宁可多传，不要漏传”的 Phase 2 比较策略。

---

## 连接与重连策略

| Option | Description | Selected |
|--------|-------------|----------|
| 保守自动重连 | 延续 SFTP/MinIO 的思路，有限重试与重连，但不要做激进并发恢复；更符合当前 concern 里的风险控制 | ✓ |
| 尽量 fail-fast | 一旦掉线就尽快报错，不做太多恢复；实现简单，但用户体验更脆弱 | |
| 更积极自愈 | 做更强的自动恢复和更多重试；成功率可能更高，但更容易引入隐藏状态问题 | |
| You decide | 你按最小正确原则定 | |

**User's choice:** 保守自动重连
**Notes:** 用户希望 FTP driver 的重连行为保持克制，避免过度复杂的自动恢复。

---

## 目录与删除语义

| Option | Description | Selected |
|--------|-------------|----------|
| 按现有 Driver 完整支持 | 为了接入现有通用 sync 层，FTP driver 需要提供递归遍历、目录创建、删除、rename；遇到服务端限制时返回明确错误 | ✓ |
| 先做最小子集 | 先只做上传/下载，删除和 rename 留后面；实现更小，但会和当前 Phase 2 requirement 不一致 | |
| 软支持危险操作 | 删除/rename 只在部分场景启用；复杂度更高，也会增加 planner 的条件分支 | |
| You decide | 你按现有架构兼容原则定 | |

**User's choice:** 按现有 Driver 完整支持
**Notes:** 用户要求 Phase 2 对齐既有 `driver.Driver` 契约，而不是引入 FTP 特殊子集模式。

---

## 时间戳精度策略

| Option | Description | Selected |
|--------|-------------|----------|
| 接受保守精度 | 明确接受 FTP mtime 可能较粗，比较时偏保守，宁可多传不漏传；并把这个限制写进后续测试与文档口径 | ✓ |
| 追求严格对齐 | 尽量把 FTP 端时间戳对齐到更精细比较；更理想，但实现和兼容性成本更高 | |
| 弱化 mtime 作用 | 尽量少依赖时间戳，把更多责任交给其他比较策略；会牵动更多实现设计 | |
| You decide | 你按最小正确原则定 | |

**User's choice:** 接受保守精度
**Notes:** 用户接受 FTP 时间戳天然不稳定，并把“保守精度”当作实现和后续测试口径的一部分。

---

## the agent's Discretion

- 具体选用哪一个 Go FTP client 库。
- 内部元数据包装和 helper 设计细节。
- 保守重连策略的具体阈值与触发点。

## Deferred Ideas

None.
