# Roadmap: gofs

## Completed Milestones

- [x] **v1.0 FTP client sync support** — Shipped 2026-04-27. Added FTP as a first-class client sync endpoint with VFS parsing, driver-backed disk<->FTP sync, polling FTP pull monitor support, real-server integration coverage, and user documentation.

Archive:

- Roadmap: `.planning/milestones/v1.0-ROADMAP.md`
- Requirements: `.planning/milestones/v1.0-REQUIREMENTS.md`
- Audit: `.planning/milestones/v1.0-MILESTONE-AUDIT.md`
- Phase artifacts: `.planning/milestones/v1.0-phases/`

## Active Milestone

No active milestone is defined. Start the next milestone before adding new roadmap phases.

## Known Carryover

- Milestone audit accepted one workflow evidence gap: Phase 3 lacks `03-VERIFICATION.md`, while implementation and later integration evidence show the FTP sync flows are complete.
- FTP v1 intentionally remains plain FTP, passive-mode oriented, one-way disk<->FTP only, with no FTPS, active FTP, FTP<->FTP, or bidirectional conflict resolution.
