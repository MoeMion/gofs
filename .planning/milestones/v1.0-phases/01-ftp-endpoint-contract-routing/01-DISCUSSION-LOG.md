# Phase 1: FTP Endpoint Contract & Routing - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-23
**Phase:** 1-FTP Endpoint Contract & Routing
**Areas discussed:** FTP endpoint contract, Credential mapping, Passive mode configuration, Defaulting behavior

---

## FTP endpoint contract

| Option | Description | Selected |
|--------|-------------|----------|
| 复用现有 query 风格 | 像 `sftp://host:port?path=...&remote_path=...` 一样继续走 query 参数；改动最小，和现有 VFS 解析模式一致 | ✓ |
| 更像标准 FTP URL | 尽量把用户信息和远端路径放进 URL 主体里，比如 `ftp://user:pass@host/path`；更直观，但会和现有 parse 模式分叉 | |
| 折中：主体+少量 query | 基础信息放 URL 主体，额外选项如 timeout/passive 放 query；可读性更好，但实现会更复杂 | |
| You decide | 你按最小改动原则替我定 | |

**User's choice:** 复用现有 query 风格
**Notes:** 用户明确选择与现有 `sftp://` / `minio://` 风格对齐，以最小改动延续当前 VFS 解析模型。

---

## Credential mapping

| Option | Description | Selected |
|--------|-------------|----------|
| 独立 FTP 参数 | 给 FTP 单独定义如 `ftp_user` / `ftp_pass`，语义清楚，不挪用 SSH 字段 | ✓ |
| 复用 SSH 参数名 | 直接借用现有 `ssh_user` / `ssh_pass`，减少字段数，但语义会混淆 | |
| 同时兼容两套 | 新旧参数都认；兼容更宽，但 Phase 1 复杂度会上升 | |
| You decide | 你按最小正确方案定 | |

**User's choice:** 独立 FTP 参数
**Notes:** 用户希望 FTP 认证字段保持协议语义清晰，不复用 SSH 命名。

---

## Passive mode configuration

| Option | Description | Selected |
|--------|-------------|----------|
| 一个布尔开关 | 只暴露最小选择，例如 `passive=true/false`；先满足 requirement，不提前设计更多兼容旋钮 | ✓ |
| 布尔开关 + 默认说明 | 仍然只有一个开关，但在文档里明确默认值和推荐值 | |
| 做细分兼容参数 | 现在就预留更细控制，如 EPSV 之类兼容选项；更灵活，但超出最小 Phase 1 目标 | |
| You decide | 你按最小改动原则定 | |

**User's choice:** 一个布尔开关
**Notes:** 用户希望 Phase 1 只暴露最小 passive-mode 控制面，不提前扩展兼容参数集合。

---

## Defaulting behavior

| Option | Description | Selected |
|--------|-------------|----------|
| 默认 21 + 端点可选超时 | 未写端口时默认 21，timeout 作为可选 query 参数；最符合现有远端协议模式 | ✓ |
| 端口必填 + 超时可选 | 更显式，但和现有 VFS 的默认端口习惯不一致 | |
| 默认 21 + 强制显式超时 | 每个 FTP 端点都要求写 timeout；更严格，但会增加配置负担 | |
| You decide | 你按最小改动原则定 | |

**User's choice:** 默认 21 + 端点可选超时
**Notes:** 用户要求与现有远端协议默认端口模式保持一致，同时把 timeout 作为可选端点级配置。

---

## the agent's Discretion

- FTP 参数的精确命名方案，只要保持 FTP 语义明确并与现有配置风格一致。
- timeout 最终是否命名为通用远端超时字段或 FTP 专属字段，只要满足“端点级可选”的已决约束。

## Deferred Ideas

None.
