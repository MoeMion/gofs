# Phase 4: FTP Verification & Discoverability - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-24
**Phase:** 4-FTP Verification & Discoverability
**Areas discussed:** FTP 集成测试形态, 测试环境来源, README 暴露程度, 已知限制的呈现方式

---

## FTP 集成测试形态

| Option | Description | Selected |
|--------|-------------|----------|
| 真实 integration 流 | 沿用现有 `integration/` + build tag 模式，跑真实 FTP 服务端和真实 gofs client/server 流；最符合当前 roadmap 和仓库已有测试体系 | ✓ |
| 只做 package-level 强化 | 继续用 fake/seam 测试，不上真实 FTP 服务端；改动更小，但达不到当前 Phase 4 的真实协议流目标 | |
| 两者都做，但 integration 优先 | 以真实 integration 为主，必要时再补少量 package-level 回归测试 | |
| You decide | 你按现有仓库模式和 roadmap 一致性定 | |

**User's choice:** 真实 integration 流
**Notes:** 用户要求 Phase 4 的验证核心是跑真实 FTP 协议流，而不是停留在 fake/seam 测试层。

---

## 测试环境来源

| Option | Description | Selected |
|--------|-------------|----------|
| 仓库内脚本初始化 | 像 SFTP/MinIO 一样，把 FTP 测试环境准备收进仓库脚本和 CI 步骤；最贴近现有项目模式 | ✓ |
| 依赖开发者本机 FTP 服务 | 文档说明用户自己先起 FTP 服务；实现最省，但 CI 和可重复性最差 | |
| 纯容器方案 | 用 docker/docker-compose 拉起 FTP 服务；可重复，但要看是否与当前仓库脚本风格一致 | |
| You decide | 你按现有 CI/testing 模式定 | |

**User's choice:** 仓库内脚本初始化
**Notes:** 用户要求 FTP integration 环境由项目自身脚本与 CI 初始化，不依赖手工准备的本机服务。

---

## README 暴露程度

| Option | Description | Selected |
|--------|-------------|----------|
| 对齐 SFTP/MinIO 粒度 | 像现有 SFTP/MinIO 一样，在 Usage/示例里把 FTP source 和 destination 都补出来；可发现性更强，也最符合 DOC-01 | ✓ |
| 只加最小示例 | 只补 1-2 个 FTP 命令示例；改动面更小，但 discoverability 较弱 | |
| 加独立 FTP 小节 | 单独给 FTP 开专节，把配置和限制集中讲；更清晰，但版面改动更大 | |
| You decide | 你按现有 README 风格和最小正确原则定 | |

**User's choice:** 对齐 SFTP/MinIO 粒度
**Notes:** 用户希望 FTP 在 README 中的可发现性达到现有 SFTP/MinIO 的同等级别，而不是附带式示例。

---

## 已知限制的呈现方式

| Option | Description | Selected |
|--------|-------------|----------|
| 示例旁 + 明确限制清单 | 在 FTP README 示例附近直接说明关键限制，并集中列一个 limitations 清单；最利于用户避免误用 | ✓ |
| 只放在限制清单 | 集中写在一个 limitations 小节；结构更整洁，但用户可能先看到示例后才发现限制 | |
| 只在注释式说明里提到 | 分散在示例说明和文字里；改动最小，但容易漏读 | |
| You decide | 你按用户预期管理效果定 | |

**User's choice:** 示例旁 + 明确限制清单
**Notes:** 用户希望 Phase 4 用更强的预期管理方式来说明 FTP v1 限制，避免示例先被看到而边界被忽略。

---

## the agent's Discretion

- FTP integration build tag、脚本名和 fixture 布局的具体命名。
- README 中 FTP 示例与限制文字的精确位置与措辞。

## Deferred Ideas

None.
