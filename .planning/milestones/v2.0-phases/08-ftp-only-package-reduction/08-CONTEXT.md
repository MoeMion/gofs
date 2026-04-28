# Phase 8: FTP-Only Package Reduction - Context

**Gathered:** 2026-04-27
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 8 converts the repository root module from a broad `gofs` CLI/server application into a focused FTP sync library module. The phase must remove or isolate old CLI/server/protocol runtime surfaces from the default library build while preserving the FTP sync behavior exposed through `ftpsync.FTPSyncService`.

This phase does not add new sync capabilities. It changes package/module boundaries and dependency graph shape so library consumers no longer pull in CLI entrypoints, flag parsing, daemon/process management, HTTP/gRPC server APIs, task/auth/session runtime, SFTP, MinIO, Docker release surfaces, or their third-party dependencies.

</domain>

<decisions>
## Implementation Decisions

### Reduction strategy
- **D-01:** Use a minimal inline FTP core strategy. Move, copy, or internalize only the disk<->FTP sync pieces `ftpsync` needs instead of continuing to import the old `sync`, `core`, `logger`, `retry`, and broad runtime chains.
- **D-02:** Do not use build tags as the primary reduction mechanism. The default import/build path should be FTP-only without requiring consumers to know special tags.
- **D-03:** Do not attempt a broad reusable refactor of the old application packages. If old packages are not needed by the FTP library, they should be removed rather than preserved for architectural neatness.

### Behavior preservation boundary
- **D-04:** Preserve all user-visible FTP library behavior established in Phases 5-7. Phase 8 changes dependency boundaries, not sync semantics.
- **D-05:** The retained FTP core must keep FTP driver behavior, path encoding, retry semantics, ignore/filtering semantics, rate limiting, cwd safety, best-effort `Result + error` behavior, hook reporting, local->FTP one-shot behavior, FTP->local one-shot behavior, and background disk->FTP debounce/shutdown semantics.
- **D-06:** The public API remains typed-options-only through `ftpsync.FTPSyncService`; old `conf.Config`, `core.VFS`, CLI URL parsing, server/task types, or global logger/reporting types must not leak into the public API.

### Module and code ownership
- **D-07:** Convert the repository root module itself into the FTP sync library. Do not create a nested `ftpsync/` submodule for Phase 8.
- **D-08:** Remove old runtime code that does not belong to the FTP library: CLI entrypoints, flags/config runtime, daemon/process management, HTTP file server, gRPC APIs, task runtime, auth/session, SFTP, MinIO, non-FTP sync/monitor paths, Docker/release surfaces where they only serve the old application runtime.
- **D-09:** Retain only necessary FTP internals behind the library API. The planner should expect file moves/deletions and import rewrites, not just small edits inside `ftpsync`.

### Verification standard
- **D-10:** Phase 8 is complete only when `go list -deps ./...` or the final library package equivalent no longer includes old runtime dependencies on the default build path.
- **D-11:** Add an explicit dependency blacklist check covering at least: `cmd`, `flag`, `daemon`, `server`, `api`, `auth`, `monitor`, `driver/sftp`, `driver/minio`, Gin, gRPC, SFTP, MinIO, Redis/cache, QUIC, OAuth, protobuf, task runtime, and Docker/release-only surfaces.
- **D-12:** `go test ./...` for the reduced root module must pass after pruning. If legacy tests are deleted with old runtime code, replacement tests must still cover the retained FTP library behavior that Phase 8 could break.
- **D-13:** `go mod tidy` should remove unnecessary runtime dependencies from `go.mod`/`go.sum`. The dependency blacklist and tests are the primary proof; `go.mod` cleanup is required evidence, not the only criterion.

### the agent's Discretion
- The exact internal package names, file movement sequence, and helper extraction shape are left to researcher/planner discretion, as long as the root module becomes a focused FTP sync library and all behavior/verification constraints above are met.
- The planner may choose whether to keep package name `ftpsync` at the root or keep a subdirectory package inside the root module if that best preserves import compatibility, but it must explicitly account for the chosen import path in Phase 9 docs/release follow-up.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Roadmap and requirements
- `.planning/ROADMAP.md` §Phase 8 — Defines the FTP-only package reduction goal, dependencies, success criteria, and Phase 9 handoff.
- `.planning/REQUIREMENTS.md` §Package Reduction — Defines PRUNE-01 through PRUNE-04.
- `.planning/PROJECT.md` §Core Value, §Out of Scope, §Key Decisions — Locks the v2.0 shift to a focused FTP sync library and removal/isolation of non-FTP runtime surfaces.

### Prior phase behavior contracts
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-01-SUMMARY.md` — Public typed option surface and constructor boundary.
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-02-SUMMARY.md` — Validation, context, and structured error contracts.
- `.planning/phases/05-public-ftpsyncservice-api-contract/05-03-SUMMARY.md` — Hook contract and no-op defaults.
- `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-01-SUMMARY.md` — One-shot adapter/result scaffolding.
- `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-02-SUMMARY.md` — Local->FTP one-shot semantics.
- `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-03-SUMMARY.md` — FTP->local one-shot and cwd-safety semantics.
- `.planning/phases/07-background-disk-ftp-lifecycle/07-VERIFICATION.md` — Verified background disk->FTP lifecycle truths that Phase 8 must not regress.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `ftpsync/` package — Current public API surface and Phase 5-7 behavior contract. This is the code that survives Phase 8.
- `driver/ftp/` — Existing FTP client backend with path encoding, passive mode, timeout, rate limiting, and FTP file operations. Retain needed behavior, but it may need package movement or dependency trimming.
- `ignore/`, `retry/`, `internal/rate/` — Semantics currently needed by `ftpsync`; retain behavior either by slimming these packages or internalizing minimal equivalents.
- Current tests under `ftpsync/*_test.go` — Useful regression baseline for validation, one-shot, hooks, background lifecycle, cwd safety, and timeout behavior.

### Established Patterns
- `ftpsync/oneshot.go` currently imports old `core`, `ignore`, `logger`, `retry`, and legacy `sync` packages. This is the main dependency leak that pulls in CLI/server/protocol runtime surfaces.
- `sync` directly imports broad runtime packages including API/gRPC, server/client, auth, SFTP, MinIO, progress, report, and old config paths. Keeping a dependency on `sync` likely prevents Phase 8 from passing PRUNE-01/04.
- `core` imports `ssh_config` and legacy logger behavior. If `ftpsync` only needs FTP endpoint/path construction, a smaller typed FTP representation should replace public reliance on `core.VFS`.
- Root `go.mod` currently includes Gin, gRPC, SFTP, MinIO, Redis/cache, QUIC, OAuth, protobuf, and other runtime dependencies. These should disappear from the reduced root module unless a retained FTP library path truly needs them.

### Integration Points
- `FTPSyncService.SyncOnce` and `StartBackground` are the public entrypoints that must continue to work after internal pruning.
- `executeSyncOnce`, FTP push/pull walking, path encoding, ignore matching, retry/rate behavior, and `background.go` are likely the core extraction targets.
- Verification should use Go tooling (`go list -deps`, `go test`, `go mod tidy`) plus an explicit blacklist script/test to prove the default module build graph is FTP-only.

</code_context>

<specifics>
## Specific Ideas

- User selected root-module conversion over a nested `ftpsync/` submodule. Downstream agents should plan for a decisive repository-level package reduction.
- User selected deletion of old runtime code rather than retaining it behind tags or just disconnecting imports.
- User selected a dependency blacklist plus tests as the completion proof.

</specifics>

<deferred>
## Deferred Ideas

- Final docs/release wording for the supported import path belongs primarily to Phase 9, but Phase 8 must leave a clear module/package shape for Phase 9 to document.
- Future FTPS, FTP->disk background polling, FTP<->FTP sync, bidirectional conflict handling, and legacy YAML/CLI parsing remain out of scope.

</deferred>

---

*Phase: 08-ftp-only-package-reduction*
*Context gathered: 2026-04-27*
