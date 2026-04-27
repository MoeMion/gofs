# Phase 3: One-Way FTP Sync Flows - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-24
**Phase:** 3-One-Way FTP Sync Flows
**Areas discussed:** FTP source 触发方式, 删除与 rename 行为, 无变化判定口径, 能力不足时的行为

---

## FTP source 触发方式

| Option | Description | Selected |
|--------|-------------|----------|
| 先保证 sync once | 先把一次性同步跑稳；长期运行的 FTP source 监听/轮询保持最小实现或延后。这样最符合“最小修改”原则 | |
| 长期运行也要支持 | 除了 sync once，还要让 FTP source 在常驻模式下可持续拉取；这通常意味着补轮询式 monitor | ✓ |
| 两者都要，但一次性优先 | roadmap 里同时覆盖两者，但实现上先保证 sync once，再补长期运行 | |
| You decide | 你按当前代码和最小范围原则定 | |

**User's choice:** 长期运行也要支持
**Notes:** 用户明确要求 `FTP→disk` 不只是一次性同步，还应在长期运行模式下持续工作。

---

## 删除与 rename 行为

| Option | Description | Selected |
|--------|-------------|----------|
| 尽量与现有 one-way 语义对齐 | 源端删除/rename 时，优先按现有 sync 语义传播到目标；能力不足时再显式报错。这样最符合 roadmap 的“preserve existing semantics” | ✓ |
| 更保守，少做危险传播 | 对 delete/rename 更谨慎，宁可少同步这些变化；更安全，但会偏离现有 sync 语义 | |
| 只保证新增和更新 | 本阶段只保证 add/update，delete/rename 弱化到后续阶段；会和当前 Phase 3 success criteria 不一致 | |
| You decide | 你按既有架构兼容原则定 | |

**User's choice:** 尽量与现有 one-way 语义对齐
**Notes:** 用户要求 FTP one-way sync 尽量保持当前系统的 delete/rename 行为契约，而不是弱化成 add/update-only。

---

## 无变化判定口径

| Option | Description | Selected |
|--------|-------------|----------|
| 偏稳健 | 接受偶尔多传，但不要因为 FTP 元数据不稳定而漏传；与前面保守比较决策一致 | ✓ |
| 偏省流量 | 尽量减少重复传输，即使会提高误判无变化的风险 | |
| 严格优化 no-op | 把 no-op 效果做得更强，哪怕引入更多复杂度或额外比较成本 | |
| You decide | 你按最小正确原则定 | |

**User's choice:** 偏稳健
**Notes:** 用户继续沿用 Phase 2 的保守比较取向，把 correctness 放在节省流量之前。

---

## 能力不足时的行为

| Option | Description | Selected |
|--------|-------------|----------|
| 显式失败并报清楚 | 遇到关键能力缺失就明确返回错误，不静默跳过；与前面 Phase 2 的“truthful backend semantics”一致 | ✓ |
| 部分成功继续 | 能同步多少先同步多少，再带 warning；用户体验更柔和，但语义更复杂 | |
| 按操作区分 | 新增/更新继续，delete/rename 失败时仅告警；更灵活，但会显著增加行为分叉 | |
| You decide | 你按当前一致性原则定 | |

**User's choice:** 显式失败并报清楚
**Notes:** 用户要求关键能力不足时，sync flow 以明确失败为主，而不是静默降级或部分成功伪装完成。

---

## the agent's Discretion

- 长期运行 `FTP→disk` 的具体实现机制。
- 轮询节奏、批处理与变化检测的内部策略。
- `sync/` 与 `monitor/` 的内部拆分细节。

## Deferred Ideas

None.
