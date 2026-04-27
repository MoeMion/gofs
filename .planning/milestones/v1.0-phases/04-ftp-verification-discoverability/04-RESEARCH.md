# Phase 4: FTP Verification & Discoverability - Research

**Date:** 2026-04-24
**Status:** Complete

## Research Question

What do we need to know to plan Phase 4 well so FTP is verified through real protocol flows and documented at the same discoverability level as SFTP and MinIO?

## Sources Read

- `.planning/STATE.md`
- `.planning/ROADMAP.md`
- `.planning/REQUIREMENTS.md`
- `.planning/phases/04-ftp-verification-discoverability/04-CONTEXT.md`
- `.planning/phases/02-ftp-driver-backend/02-01-SUMMARY.md`
- `.planning/phases/02-ftp-driver-backend/02-02-SUMMARY.md`
- `.planning/phases/03-one-way-ftp-sync-flows/03-01-SUMMARY.md`
- `.planning/phases/03-one-way-ftp-sync-flows/03-02-SUMMARY.md`
- `.planning/codebase/TESTING.md`
- `integration/integration_test.go`
- `integration/integration_sftp_test.go`
- `integration/integration_minio_test.go`
- `integration/testdata/conf/run-gofs-sftp-push-client.yaml`
- `integration/testdata/conf/run-gofs-sftp-pull-client.yaml`
- `integration/testdata/test/test-gofs-sftp-push.yaml`
- `integration/testdata/test/test-gofs-sftp-pull.yaml`
- `integration/testdata/conf/run-gofs-minio-push-client.yaml`
- `integration/testdata/conf/run-gofs-minio-pull-client.yaml`
- `integration/testdata/test/test-gofs-minio-push.yaml`
- `integration/testdata/test/test-gofs-minio-pull.yaml`
- `.github/workflows/go.yml`
- `README.md`
- `README-CN.md`
- `go.mod`
- pyftpdlib docs: tutorial + API reference

## Key Findings

### 1. Existing repository pattern for realistic backend verification

- Real backend coverage already lives in `integration/` behind dedicated build tags such as `integration_test_sftp` and `integration_test_minio`.
- Shared lifecycle logic is already centralized in `integration/integration_test.go` through `testIntegrationClientServer(...)`.
- Backend-specific suites are intentionally thin: `integration/integration_sftp_test.go` and `integration/integration_minio_test.go` mostly enumerate config/test fixture pairs.
- This directly supports D-01, D-02, and D-05: FTP should follow the same pattern instead of inventing a new harness.

### 2. Existing fixture style we should mirror

- Runtime configs live in `integration/testdata/conf/*.yaml`.
- Action/assertion scenarios live in `integration/testdata/test/*.yaml` and are executed via `github.com/no-src/fsctl/command`.
- SFTP and MinIO both verify push and pull through:
  - a backend-specific run config per direction
  - a backend-specific action script per direction
  - a thin tagged test file in `integration/`
- The push scenarios already include nested file mutations, deletions, hashes, and symlink edge cases; the pull scenarios already include directory creation and delete checks. FTP can reuse this shape with FTP-aware expectations.

### 3. CI integration pattern to preserve

- `.github/workflows/go.yml` provisions backend services on Ubuntu before running their dedicated tagged integration commands.
- SFTP uses `./scripts/sftp/init-sftp.sh`.
- MinIO uses `./scripts/minio/install-minio.sh` and `./scripts/minio/mount-minio.sh`.
- The minimal-change Phase 4 path is therefore:
  1. add `scripts/ftp/` setup,
  2. add an Ubuntu-only workflow step to initialize FTP,
  3. add `go test -v -race -tags=integration_test_ftp ./integration`.

### 4. Recommended FTP test server approach

#### Chosen approach: repo-owned Python/pyftpdlib startup script

Why:

- D-04/D-05 require repository-owned setup, not manual local services.
- `pyftpdlib` is purpose-built for spinning up a local FTP server in tests and supports:
  - explicit users/passwords
  - writable home directory
  - passive mode / passive port range
  - clear-text FTP without dragging FTPS into scope
- The repository already relies on shell setup scripts for protocol services, so a `scripts/ftp/init-ftp.sh` + small Python helper fits the current pattern.

Recommended implementation shape:

- `scripts/ftp/init-ftp.sh`
  - creates a workspace directory for FTP data
  - ensures Python 3 tooling exists on Ubuntu CI
  - installs `pyftpdlib` in a local/ephemeral way
  - starts a background FTP server bound to `127.0.0.1:2121`
  - provisions one test user with full read/write/delete/rename permissions
  - constrains passive ports to a known range for deterministic local operation
  - writes a pid file / startup log so follow-up steps can detect failures
- `scripts/ftp/server.py` (or similar small helper)
  - defines the pyftpdlib `DummyAuthorizer` user
  - points home directory at the FTP workspace
  - enables passive ports
  - uses plain FTP only

Why not FTPS / Docker-first / manual system FTP service:

- FTPS violates D-10/D-11 and the project’s v1 constraint.
- Docker would be heavier than the existing script-based backend init shape.
- System FTP daemons vary by distro and introduce more CI variance than a pinned Python helper.

### 5. Scope of realistic FTP integration coverage

Phase 4 must prove more than route selection. The real server suite should cover:

- `disk→FTP` push flow (TEST-01)
- `FTP→disk` pull flow (TEST-02)
- nested directory/file behavior (TEST-03)
- delete propagation and/or rename-related behavior on the FTP path (TEST-03)

The most consistent mapping with existing fixtures is:

- FTP push scenario:
  - create nested files/directories locally
  - update file contents
  - delete one file after initial sync
  - assert remote nested artifacts exist and deleted artifact disappears
- FTP pull scenario:
  - seed FTP server workspace with nested files/directories
  - remove one remote file before/after poll cycle depending on scenario timing
  - assert local destination mirrors nested structure and deletion outcome

### 6. README discoverability expectations

- `README.md` and `README-CN.md` already place SFTP and MinIO as first-class protocol sections in the Usage area.
- D-07/D-08 require FTP to be equally visible, meaning FTP should get both:
  - “FTP Push Client” example
  - “FTP Pull Client” example
- D-10/D-11 require limitations both near the examples and in an explicit limitations list.

Recommended doc changes:

- Add `### FTP Push Client` and `### FTP Pull Client` adjacent to existing SFTP / MinIO protocol examples.
- Use the actual implemented endpoint grammar from earlier phases:
  - `ftp://127.0.0.1:21?...&ftp_user=...&ftp_pass=...&ftp_passive=true`
- Add a short note directly under FTP examples stating:
  - plain FTP only in v1
  - passive mode only / active mode unsupported
  - no FTP↔FTP sync
- Add a dedicated “FTP v1 limitations” bullet list in the protocol docs area.
- Mirror the same additions in `README-CN.md` so discoverability is not English-only.

### 7. Operational constraints from shipped FTP behavior

From prior phases we must document and test against the actual product behavior:

- FTP uses dedicated `ftp_*` parameters.
- Port defaults to 21 when omitted, but examples can still show `:21` explicitly for discoverability.
- Active mode is explicitly unsupported and should be documented as such.
- FTP comparisons are conservative; docs should not overpromise perfect metadata fidelity.
- FTP v1 is one-way only through disk↔FTP, not FTP↔FTP and not bidirectional conflict resolution.

## Standard Stack

- Existing Go `testing` integration framework in `integration/`
- Existing YAML action/config fixtures in `integration/testdata/`
- Existing GitHub Actions workflow in `.github/workflows/go.yml`
- New repo-owned FTP setup under `scripts/ftp/`
- `pyftpdlib` as the temporary local FTP server implementation used by scripts/CI

## Architecture Patterns

- **Use the existing integration harness** — do not create a separate test runner or bespoke CLI.
- **Keep backend-specific test files thin** — enumerate scenarios in `integration/integration_ftp_test.go`, not custom orchestration.
- **Put environment setup in scripts** — CI should call scripts, not inline a large ad hoc FTP bootstrap block.
- **Keep docs in existing README surfaces** — do not create a separate FTP-only doc system.

## Don't Hand-Roll

- Do not write a custom in-Go FTP server just for tests.
- Do not add FTPS, active mode, or FTP↔FTP verification to “make the test harness complete”.
- Do not hide FTP limitations in CI notes only; they must be visible in README examples and limitations text.

## Common Pitfalls

- **Testing only package-level fakes again** would violate D-01 and fail the phase goal.
- **Using a manual local FTP service** would violate D-04/D-06 and make verification non-repeatable.
- **Adding only English docs** would make FTP discoverability partial relative to existing bilingual docs.
- **Forgetting passive port control** can make local FTP data connections flaky in CI.
- **Documenting unsupported features implicitly** instead of explicitly would violate D-10 to D-12.

## Planned Coverage Mapping

| Source Item | Planned Coverage |
|-------------|------------------|
| TEST-01 | Real `disk→FTP` integration fixture + tagged integration test + CI command |
| TEST-02 | Real `FTP→disk` integration fixture + tagged integration test + CI command |
| TEST-03 | FTP fixture scenarios with nested paths and delete/rename-relevant assertions |
| DOC-01 | FTP push/pull examples in `README.md` and `README-CN.md` |
| DOC-02 | Explicit FTP v1 limitations near examples and in a dedicated limitations list |
| D-01 / D-02 / D-03 | Build-tagged `integration/` suite with real FTP server, while keeping package tests as supporting coverage |
| D-04 / D-05 / D-06 | Repo-owned `scripts/ftp/` provisioning + CI step |
| D-07 / D-08 / D-09 | README additions parallel to SFTP/MinIO sections |
| D-10 / D-11 / D-12 | Explicit limitations text in both READMEs |

## Recommended Plan Shape

1. **Foundation plan:** add repo-owned FTP integration bootstrap plus FTP config/test fixtures.
2. **Verification plan:** add tagged FTP integration suite and wire it into CI.
3. **Discoverability plan:** add FTP push/pull examples and explicit limitations to both READMEs.

## Outcome

Phase 4 does need research. The repo already has a strong pattern to follow, and the lowest-risk implementation is to add FTP as one more tagged real-backend integration with script-owned provisioning plus README updates that mirror SFTP and MinIO visibility.
