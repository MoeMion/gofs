# Phase 7: Background Disk→FTP Lifecycle - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-27
**Phase:** 07-background-disk-ftp-lifecycle
**Areas discussed:** 初始同步行为, 错误暴露方式, 事件合并策略, 启动方式, Handle contract

---

## 初始同步行为

| Option | Description | Selected |
|--------|-------------|----------|
| Sync first | 启动后先跑一轮全量/增量 one-shot，再进入监听；最稳，能保证远端先追平 | ✓ |
| Watch only | 只监听启动后的新变化，不补历史差异；启动更快，但可能留下启动前未同步内容 | |
| Caller option later | Phase 7 先固定一种，后续再考虑开放成 option | |

**User's choice:** Sync first
**Notes:** 后台模式优先保证远端追平，再开始长期监听。

---

## 错误暴露方式

| Option | Description | Selected |
|--------|-------------|----------|
| Report and continue | 通过 handle / hook 暴露错误，但后台监听继续运行；更适合长期驻留服务 | ✓ |
| Fail terminally | 遇到一次关键同步失败就停止整个后台任务 | |
| Mixed policy later | Phase 7 先固定一种，后续再开放策略切换 | |

**User's choice:** Report and continue
**Notes:** 单次同步失败不应让后台任务默认退出。

---

## 事件合并策略

| Option | Description | Selected |
|--------|-------------|----------|
| Debounce/coalesce | 短时间窗口内合并重复变化，减少抖动和重复上传 | ✓ |
| Immediate per event | 每个事件都尽快处理，语义直接，但更容易重复上传 | |
| Fixed small delay | 总是加一个很小的稳定延迟，不做更复杂合并 | |

**User's choice:** Debounce/coalesce
**Notes:** 减少 bursty 本地变化导致的重复上传比逐事件即时响应更重要。

---

## 启动方式

| Option | Description | Selected |
|--------|-------------|----------|
| StartBackground only | 调用后立即启动，返回 handle；API 最小，和当前 Phase 5 合约一致 | ✓ |
| Two-step start | 先构造 run/worker，再显式 Start；更灵活，但 API 更复杂 | |
| Keep current, enrich handle | 保持 StartBackground 入口不变，把可控性都放到返回的 handle 上 | |

**User's choice:** StartBackground only
**Notes:** 不拆出新的 public lifecycle builder；控制能力集中在 handle。

---

## Handle contract

| Option | Description | Selected |
|--------|-------------|----------|
| Stop + Wait + Err | 调用方可以主动停止、等待退出、读取最终错误；最适合长期运行的库 API | ✓ |
| Observe only | 只观测状态/错误，停止完全交给 context 取消 | |
| Minimal stop only | 只提供 Stop，等待和最终状态走 hook/外部机制 | |

**User's choice:** Stop + Wait + Err
**Notes:** handle 需要是一个强控制生命周期对象，而不只是被动观测接口。

---

## the agent's Discretion

- debounce/coalescing 的具体窗口和实现机制
- `Handle` 的具体结构与方法细节
- 背景任务内部 worker/channel/watch loop 的实现拆分

## Deferred Ideas

- 背景 `FTP→disk` polling 生命周期
- 更丰富的后台 policy 开关
- Phase 8 的 package pruning / monitor surface 收缩
