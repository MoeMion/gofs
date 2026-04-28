# Phase 9: Verification, Examples, and Migration Docs - Context

**Gathered:** 2026-04-28
**Status:** Ready for planning

<domain>
## Phase Boundary

Phase 9 is the final hardening and adoption phase for the reduced FTP sync library. It must prove the final local-module package contract with automated tests, real FTP integration coverage, compilable examples, README usage, and migration documentation.

This phase does not add new sync capabilities. It validates and documents the existing `ftpsync.FTPSyncService` contract after Phase 8 reduced the repository to the FTP library build graph.

</domain>

<decisions>
## Implementation Decisions

### Local module and import path
- **D-01:** The final library is a local module named `ftpsync`, intended to be embedded into other projects as local source rather than published on the internet.
- **D-02:** Phase 9 should change the root `go.mod` module path from `github.com/no-src/gofs` to `ftpsync` so examples and docs can use `import "ftpsync"` directly.
- **D-03:** Documentation should not present `github.com/no-src/gofs/ftpsync` as the final consumer import path. If mentioned, it should be historical/internal context only.

### Real FTP verification
- **D-04:** Real FTP integration tests should use a Go-native test server/fixture started from Go tests, not Docker or Python scripts.
- **D-05:** The real FTP tests must cover both library-based local->FTP and FTP->local one-shot flows without invoking any old CLI/runtime code.
- **D-06:** The test server should be scoped to tests and local loopback only; it should not introduce a production FTP server mode or public API.

### Example format
- **D-07:** Provide both README usage snippets and Go `Example...` tests so examples are human-readable and compiler-checked.
- **D-08:** Examples must cover one-shot local->FTP, one-shot FTP->local, and background disk->FTP usage with typed options.
- **D-09:** Examples should use `import "ftpsync"` and should avoid legacy YAML, CLI flags, URL parser, server/task, SFTP, MinIO, or Docker concepts.

### Migration documentation
- **D-10:** Migration docs should explicitly describe v2.0 as a breaking migration from the old CLI/server application to a local Go library module.
- **D-11:** The docs must clearly state removed/unsupported surfaces: no CLI runtime, no HTTP/gRPC/file server/task/auth/session runtime, no SFTP, no MinIO, no Docker release artifact, no FTPS, no FTP server mode, no FTP<->FTP sync, no FTP->disk background polling, and no bidirectional conflict resolution.
- **D-12:** The docs should explain how callers now configure everything through typed Go options and `FTPSyncService`, not YAML/CLI config files.

### Test coverage priority
- **D-13:** Phase 9 should use a full coverage checklist, not a documentation-only pass.
- **D-14:** Automated tests should cover public service construction, validation, one-shot push, one-shot pull, background disk->FTP lifecycle, cwd safety, FTP path encoding, passive/default behavior, cancellation, and background goroutine shutdown.
- **D-15:** Existing Phase 8 dependency boundary tests should continue to pass after the module path change to `ftpsync`.

### the agent's Discretion
- The exact Go-native FTP test server implementation is left to researcher/planner discretion, provided it is local-loopback, test-only, and does not reintroduce Docker/Python/old runtime dependencies.
- The exact README structure and migration file name are flexible, but README must be updated and migration notes must be discoverable from README.

</decisions>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Roadmap and requirements
- `.planning/ROADMAP.md` §Phase 9 — Defines final verification, examples, and migration-doc success criteria.
- `.planning/REQUIREMENTS.md` §Verification and Documentation — Defines VERIFY-01 through VERIFY-03 and DOC-01 through DOC-03.
- `.planning/PROJECT.md` §Core Value, §Out of Scope, §Current Milestone — Defines the final library shape and deferred capabilities.

### Prior phase contracts
- `.planning/phases/08-ftp-only-package-reduction/08-CONTEXT.md` — Locks Phase 8 reduction and default FTP-only build graph decisions.
- `.planning/phases/08-ftp-only-package-reduction/08-VERIFICATION.md` — Confirms the reduced module currently exposes only `github.com/no-src/gofs/ftpsync`, has no old runtime dependencies, and passes dependency-boundary tests.
- `.planning/phases/07-background-disk-ftp-lifecycle/07-VERIFICATION.md` — Defines verified background lifecycle behavior that docs/tests must preserve.
- `.planning/phases/06-one-shot-disk-ftp-sync-through-library-api/06-03-SUMMARY.md` — Defines FTP->local cwd-safety behavior that Phase 9 must test/document.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `ftpsync/` — The surviving package after Phase 8. Public API includes `Options`, `Endpoint`, `FTPOptions`, `RetryOptions`, `IgnoreRule`, `HookSet`, `FTPSyncService`, `SyncOnce`, and `StartBackground`.
- `ftpsync/oneshot_test.go`, `background_test.go`, `validation_test.go`, `context_test.go`, `public_api_test.go`, `dependency_boundary_test.go` — Existing regression tests and fakes are the baseline for VERIFY-01 and VERIFY-03.
- `ftpsync/internal_ftp.go` — Package-local FTP client and path codec. Useful target for real FTP integration verification and path encoding coverage.
- `.github/workflows/go.yml` — Current reduced CI workflow runs Go 1.24 build/test for `./...`.

### Established Patterns
- Phase 8 reduced `go list ./...` to a single package under the current module. Changing `go.mod` to `module ftpsync` will require updating any tests or docs that assert the old path.
- Public examples should be testable Go examples rather than unchecked README-only snippets.
- Old runtime directories and Docker/release workflows were removed by design; Phase 9 must not reintroduce those surfaces for testing convenience.

### Integration Points
- `go.mod` module path is a direct Phase 9 target because the user selected local module import path `ftpsync`.
- README and migration docs must align with the new module path and local embedding use case.
- Real FTP integration tests should attach to `FTPSyncService.SyncOnce`, not lower-level `internal_ftp` helpers, so they verify the public contract.

</code_context>

<specifics>
## Specific Ideas

- User wants the package imported as `import "ftpsync"` and used as a local module embedded in other projects, not published online.
- User selected changing `go.mod` to `module ftpsync` to make that import path direct.
- User prefers an explicit breaking-change migration note over soft compatibility language.
- User selected README plus Go Example tests for examples.
- User selected a full automated coverage checklist for final verification.

</specifics>

<deferred>
## Deferred Ideas

- Internet publication path and remote module release workflow are out of scope for this milestone.
- FTPS, FTP server mode, FTP->disk background polling, FTP<->FTP sync, bidirectional conflict resolution, and legacy YAML/CLI parser support remain deferred future capabilities.

</deferred>

---

*Phase: 09-verification-examples-and-migration-docs*
*Context gathered: 2026-04-28*
