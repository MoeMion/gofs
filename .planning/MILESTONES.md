# Milestones

## v1.0 FTP client sync support (Shipped: 2026-04-27)

**Phases completed:** 4 phases, 9 plans, 18 tasks

**Key accomplishments:**

- FTP endpoint parsing in core.VFS with dedicated ftp query fields, passive-mode/timeout support, and default-port round-trip coverage.
- FTP sync and monitor factories now route `ftp://` endpoints into explicit FTP constructor placeholders with regression coverage instead of generic unsupported-path fallthrough.
- FTP driver integration using github.com/jlaffaye/ftp with conservative metadata fallback, explicit capability errors, and deterministic reconnect coverage.
- FTP disk↔endpoint sync constructors now route through the real driver-backed sync path with deterministic regression coverage instead of Phase 1 deferred placeholder errors.
- FTP→disk source monitoring now runs through a real driver-backed polling monitor with explicit startup rejection when neither sync_once nor sync_cron is configured.
- FTP one-way sync semantics are now locked down with regression tests for routing, delete/rename behavior, and conservative no-op checks that use precise FTP file times when available.
- Repo-owned pyftpdlib bootstrap and tagged FTP fixtures now verify real disk→FTP and FTP→disk flows, including nested paths and delete/rename-relevant behavior, against a live plain-FTP server.
- Tagged real-server FTP push/pull integration coverage now runs through the standard harness and is enforced in GitHub Actions with Ubuntu FTP provisioning.

---
