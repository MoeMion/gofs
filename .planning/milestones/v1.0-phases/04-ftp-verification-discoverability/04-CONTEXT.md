# Phase 4: FTP Verification & Discoverability - Context

**Gathered:** 2026-04-24
**Status:** Ready for planning

<domain>
## Phase Boundary

This phase proves the FTP path through realistic protocol-level verification and makes FTP usage discoverable to users. It owns real integration coverage against an FTP server plus the README/documentation work needed so users can find, configure, and correctly scope FTP support. It does not expand the FTP feature set beyond v1: no FTPS, no FTP server mode, no FTP↔FTP, and no bidirectional sync semantics.

</domain>

<decisions>
## Implementation Decisions

### Verification style
- **D-01:** Phase 4 should verify FTP using real integration flows rather than stopping at package-level fake/seam tests.
- **D-02:** The verification approach should align with the repository's existing `integration/` + build-tag pattern used for SFTP and MinIO.
- **D-03:** Package-level tests remain useful as supporting regression coverage, but they are not sufficient to satisfy Phase 4's primary verification goal.

### Test environment source
- **D-04:** The FTP test server should be provisioned through repository-owned scripts and CI steps rather than relying on a developer's manually prepared local FTP service.
- **D-05:** The setup should follow the current project habit of dedicated integration setup scripts, similar to existing SFTP and MinIO flows.
- **D-06:** The goal is repeatable project-native verification, not ad hoc local reproduction instructions.

### README and discoverability scope
- **D-07:** README discoverability for FTP should match the granularity and visibility of existing SFTP and MinIO documentation.
- **D-08:** FTP should be represented in user-facing usage/examples for both source and destination directions, not hidden behind a minimal one-off snippet.
- **D-09:** Documentation changes should preserve the established README structure and visual style where possible instead of inventing a separate documentation system.

### Limitation communication
- **D-10:** FTP v1 limitations should be shown close to the usage examples as well as in an explicit limitations list.
- **D-11:** The documentation should clearly state at least: plain FTP only, no FTP↔FTP, no active mode support in v1, and the expectation that backend capability failures are surfaced explicitly.
- **D-12:** Limitation language should manage user expectations directly rather than assuming readers will infer the boundaries from missing examples.

### the agent's Discretion
- The exact build tags, fixture file layout, and script names for FTP integration coverage, as long as they fit the repo's current testing pattern.
- The precise placement of FTP example commands within README, provided discoverability remains comparable to SFTP and MinIO.
- The exact wording of the limitations section, provided the locked constraints above remain explicit.

</decisions>

<specifics>
## Specific Ideas

- Treat Phase 4 as the point where FTP support becomes trustworthy and discoverable, not where new FTP capabilities are added.
- Prefer the repository's existing integration-test ergonomics over introducing a brand new test harness style.
- Make it hard for users to miss the FTP limitations by placing them both near examples and in a dedicated summary list.

</specifics>

<canonical_refs>
## Canonical References

**Downstream agents MUST read these before planning or implementing.**

### Project and roadmap scope
- `.planning/PROJECT.md` — Global FTP constraints and the project's minimal-change objective.
- `.planning/REQUIREMENTS.md` — Phase 4 requirement set `TEST-01`, `TEST-02`, `TEST-03`, `DOC-01`, and `DOC-02`.
- `.planning/ROADMAP.md` §Phase 4 — Phase goal, success criteria, and dependency on Phase 3.
- `.planning/STATE.md` — Current project concerns and prior decisions that documentation must not contradict.

### Prior locked decisions
- `.planning/phases/01-ftp-endpoint-contract-routing/01-CONTEXT.md` — FTP endpoint grammar and config contract that documentation/examples must reflect accurately.
- `.planning/phases/02-ftp-driver-backend/02-CONTEXT.md` — Plain-FTP-only scope, active-mode rejection, conservative metadata, and truthful backend failure behavior.
- `.planning/phases/03-one-way-ftp-sync-flows/03-CONTEXT.md` — One-way-only semantics, long-running FTP source behavior, and explicit failure expectations.
- `.planning/phases/02-ftp-driver-backend/02-01-SUMMARY.md` / `02-02-SUMMARY.md` — What was actually shipped in the driver/backend layer.
- `.planning/phases/03-one-way-ftp-sync-flows/03-01-SUMMARY.md` / `03-02-SUMMARY.md` — What was actually shipped in the monitor/sync semantics layer.

### Existing test and doc patterns
- `.planning/codebase/TESTING.md` — Existing integration build-tag strategy, test organization, and CI patterns.
- `integration/integration_test.go` — Shared process-level integration harness pattern.
- `.github/workflows/go.yml` — Current CI shape for standard, SFTP, and MinIO integration flows.
- `README.md` — Current user-facing protocol sections and example style for SFTP and MinIO.

</canonical_refs>

<code_context>
## Existing Code Insights

### Reusable Assets
- `integration/` already contains tagged integration suites and shared lifecycle helpers that can likely host FTP integration tests without inventing a new framework.
- `.github/workflows/go.yml` already provisions extra services for protocol-specific integration suites, which gives FTP a natural insertion point.
- `README.md` already documents protocol-specific flows for SFTP and MinIO in a style that FTP can mirror.

### Established Patterns
- Real protocol validation in this repository is typically done through dedicated integration build tags and CI setup steps.
- User-facing protocol discoverability is provided through README usage sections and concrete command examples rather than hidden reference docs.
- Protocol limitations are not always centralized today, so Phase 4 should intentionally make FTP's limits more explicit without breaking README tone.

### Integration Points
- New FTP integration coverage will likely land in `integration/` with its own build tag and setup scripts.
- CI will likely need an FTP-specific initialization step parallel to existing SFTP and MinIO setup.
- README will need new FTP source/destination examples inserted near comparable SFTP/MinIO sections and accompanied by explicit limitations text.

</code_context>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope.

</deferred>

---

*Phase: 04-ftp-verification-discoverability*
*Context gathered: 2026-04-24*
