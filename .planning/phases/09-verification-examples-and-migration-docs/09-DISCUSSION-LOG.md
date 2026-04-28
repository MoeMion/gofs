# Phase 9: Verification, Examples, and Migration Docs - Discussion Log

> **Audit trail only.** Do not use as input to planning, research, or execution agents.
> Decisions are captured in CONTEXT.md — this log preserves the alternatives considered.

**Date:** 2026-04-28
**Phase:** 09-Verification, Examples, and Migration Docs
**Areas discussed:** 真实 FTP 验证, 示例形式, 迁移文档, 发布路径措辞, 测试优先级, 模块路径

---

## 真实 FTP 验证

| Option | Description | Selected |
|--------|-------------|----------|
| Go 内置测试服务器 | 用 Go 测试代码启动轻量本地 FTP server/fixture，不依赖 Docker 或外部服务。 | ✓ |
| Python pyftpdlib 脚本 | 复用 Python FTP server 脚本/vendor 思路，但引入 Python 环境依赖。 | |
| Docker FTP server | 接近部署环境，但会重新引入 Phase 8 删除的 Docker surfaces。 | |
| 仅保留 fake 测试 | 不做真实 FTP server，只保留 fakes。 | |

**User's choice:** Go 内置测试服务器
**Notes:** 必须满足 VERIFY-02，覆盖 library-based local->FTP 和 FTP->local，不恢复 CLI/runtime/Docker 测试路径。

---

## 示例形式

| Option | Description | Selected |
|--------|-------------|----------|
| README+Go Examples | README 展示快速用法，`Example...` tests 提供可编译示例。 | ✓ |
| examples/目录 | 提供完整小程序。 | |
| README snippets only | 最轻量但不可编译验证。 | |
| pkg.go.dev 文档优先 | 主要扩展 doc.go 和 Example tests，README 简化。 | |

**User's choice:** README+Go Examples
**Notes:** 需覆盖 one-shot push、one-shot pull、background disk->FTP。

---

## 迁移文档

| Option | Description | Selected |
|--------|-------------|----------|
| 明确破坏性迁移 | 直接说明 v2.0 移除了旧 CLI/server/SFTP/MinIO/HTTP/gRPC/task/auth runtime surfaces。 | ✓ |
| 温和兼容措辞 | 弱化旧功能移除。 | |
| 只列新用法 | 不主动讲旧功能。 | |

**User's choice:** 明确破坏性迁移
**Notes:** 文档必须清楚列出移除和限制，避免用户误以为旧 CLI/server 仍支持。

---

## 发布路径措辞 / 模块路径

| Option | Description | Selected |
|--------|-------------|----------|
| 保持当前导入路径 | 写 `import "github.com/no-src/gofs/ftpsync"`。 | |
| 根路径包名 ftpsync | 移动代码到 module root，用户 import root path。 | |
| 新路径 / 本地模块 | 使用 `import "ftpsync"`，不面向互联网发布。 | ✓ |

**User's choice:** 新路径，直接使用 `import "ftpsync"`；此包不会在互联网发布，只作为本地模块嵌入其他项目代码。
**Follow-up choice:** 改 `go.mod` 为 `module ftpsync`。
**Notes:** Phase 9 planner must account for module path change and update tests/docs accordingly.

---

## 测试优先级

| Option | Description | Selected |
|--------|-------------|----------|
| 全覆盖清单 | 覆盖 construction/validation、one-shot push/pull、real FTP integration、cwd safety、path encoding、passive/defaults、cancellation、background goroutine shutdown。 | ✓ |
| 真实 FTP 优先 | 优先 real FTP 和 README examples，其他回归点只补缺口。 | |
| 文档优先 | 重点 README/examples/migration，测试只运行已有套件。 | |

**User's choice:** 全覆盖清单
**Notes:** Phase 9 is final hardening, not a docs-only pass.

---

## the agent's Discretion

- 具体 Go-native FTP test server implementation 由 researcher/planner 决定，但必须 test-only、loopback-only、无 Docker/Python 依赖。
- README 结构和 migration 文件名可由 planner 决定，但 README 必须链接到迁移说明。

## Deferred Ideas

- 互联网发布路径和远程 module release workflow 暂不处理。
- FTPS、FTP server mode、FTP->disk background polling、FTP<->FTP、bidirectional conflict handling、legacy YAML/CLI parser support 继续延期。
