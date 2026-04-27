---
phase: 04-ftp-verification-discoverability
reviewed: 2026-04-24T00:00:00Z
depth: standard
files_reviewed: 15
files_reviewed_list:
  - scripts/ftp/init-ftp.sh
  - scripts/ftp/server.py
  - integration/integration_ftp_test.go
  - integration/testdata/conf/run-gofs-ftp-push-client.yaml
  - integration/testdata/conf/run-gofs-ftp-pull-client.yaml
  - integration/testdata/test/test-gofs-ftp-push.yaml
  - integration/testdata/test/test-gofs-ftp-pull.yaml
  - .github/workflows/go.yml
  - README.md
  - README-CN.md
  - .gitignore
  - driver/ftp/ftp.go
  - driver/ftp/file.go
  - sync/driver_pull_client_sync.go
  - sync/disk_sync.go
findings:
  critical: 0
  warning: 3
  info: 0
  total: 3
status: issues_found
---

# Phase 4: Code Review Report

**Reviewed:** 2026-04-24T00:00:00Z
**Depth:** standard
**Files Reviewed:** 15
**Status:** issues_found

## Summary

Reviewed the Phase 4 FTP verification/discoverability implementation across the repo-owned FTP bootstrap, tagged integration suite, CI wiring, documentation, and the minimal runtime fixes added to make live FTP verification pass.

The implementation is directionally solid: it adds real FTP push/pull coverage, keeps the test server localhost-bound, and wires the suite into GitHub Actions. The main remaining concerns are one documentation mismatch that can cause surprising runtime behavior, one integration-flakiness pattern in the new FTP fixtures, and one CI fragility from installing the FTP server dependency from the public network at test time.

## Warnings

### WR-01: FTP README push examples omit the local-sync behavior that the runtime still enables by default

**File:** `README.md:481-491`, `README-CN.md:461-471`
**Issue:** The new FTP push examples document `-dest="ftp://..."` with only `remote_path`, `ftp_user`, `ftp_pass`, and `ftp_passive=true`. In the shipped runtime, FTP destinations still parse a local `path` and `local_sync_disabled` flag, and push sync keeps local disk sync enabled unless `local_sync_disabled=true` is set. With the documented example, `path` is empty, which resolves to the current working directory path internally, so users can get an unintended local mirror/write target in addition to the FTP upload. That makes the new examples misleading for real usage.
**Fix:** Update both README examples to either disable local mirroring explicitly or document the local mirror path explicitly. For example:

```bash
$ gofs -source="./source" -dest="ftp://127.0.0.1:21?local_sync_disabled=true&remote_path=/gofs_ftp_server&ftp_user=ftp_user&ftp_pass=ftp_pwd&ftp_passive=true"
```

or, if local mirroring is intended:

```bash
$ gofs -source="./source" -dest="ftp://127.0.0.1:21?local_sync_disabled=false&path=./dest&remote_path=/gofs_ftp_server&ftp_user=ftp_user&ftp_pass=ftp_pwd&ftp_passive=true"
```

### WR-02: FTP integration fixtures rely on fixed sleeps instead of completion/state-based synchronization

**File:** `integration/testdata/test/test-gofs-ftp-push.yaml:30`, `integration/testdata/test/test-gofs-ftp-pull.yaml:30`
**Issue:** Both new FTP scenarios wait with a hard-coded `sleep: 10s` before asserting results. This is a classic integration flakiness pattern: fast environments waste time, while slower CI environments can still fail intermittently if sync, server startup, or filesystem visibility takes longer than the fixed delay. Because these are `sync_once` scenarios, the test flow already has a natural completion point and should not need a blind wait.
**Fix:** Replace the fixed sleep with completion-aware synchronization. The most robust options are:

```text
1. Wait for the sync_once client process to complete before running assertions.
2. Or poll for the expected file/dir state with retries and a timeout instead of a blind sleep.
```

If the current harness must stay action-driven, add retrying `is-exist`/`is-equal` support with timeout semantics and use that in place of `sleep: 10s`.

### WR-03: CI FTP bootstrap depends on downloading pyftpdlib during every run

**File:** `scripts/ftp/init-ftp.sh:16-23`, `.github/workflows/go.yml:52-65`
**Issue:** The FTP setup path installs `pyftpdlib` dynamically with `python3 -m pip install ... pyftpdlib` during CI execution. That makes the new FTP verification path depend on PyPI/network availability and package-install behavior at runtime, which is a CI setup risk unrelated to the code under test. A transient package index outage, TLS/intercept issue, or runner-level pip problem can fail the FTP verification job before the suite even starts.
**Fix:** Make the dependency source deterministic for CI. Examples:

```text
1. Cache the pip download/install directory in GitHub Actions.
2. Vendor a pinned wheel/sdist for pyftpdlib under repo-controlled test assets.
3. Build a small reusable CI setup step/image that already contains the FTP test dependency.
```

At minimum, pin the exact package version in the install command to reduce drift.

---

_Reviewed: 2026-04-24T00:00:00Z_
_Reviewer: the agent (gsd-code-reviewer)_
_Depth: standard_
