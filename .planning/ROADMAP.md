# Roadmap: gofs

## Milestones

- ✅ **v1.0 FTP client sync support** - Phases 1-4 (shipped 2026-04-27)
- ✅ **v2.0 FTP Sync Library** - Phases 5-9 (shipped 2026-04-28)

## Overview

gofs has completed its transition from a broad CLI/server-centered file synchronization application into a focused local Go FTP sync library. The current shipped package is `ftpsync/ftpsync`, centered on `FTPSyncService` and typed Go options for one-shot disk<->FTP sync plus background disk->FTP sync.

## Phases

<details>
<summary>✅ v1.0 FTP client sync support (Phases 1-4) - SHIPPED 2026-04-27</summary>

See archived milestone artifacts:

- Roadmap: `.planning/milestones/v1.0-ROADMAP.md`
- Requirements: `.planning/milestones/v1.0-REQUIREMENTS.md`
- Audit: `.planning/milestones/v1.0-MILESTONE-AUDIT.md`
- Phase artifacts: `.planning/milestones/v1.0-phases/`

</details>

<details>
<summary>✅ v2.0 FTP Sync Library (Phases 5-9) - SHIPPED 2026-04-28</summary>

See archived milestone artifacts:

- Roadmap: `.planning/milestones/v2.0-ROADMAP.md`
- Requirements: `.planning/milestones/v2.0-REQUIREMENTS.md`
- Audit: `.planning/milestones/v2.0-MILESTONE-AUDIT.md`
- Phase artifacts: `.planning/milestones/v2.0-phases/`

</details>

## Progress

| Phase | Milestone | Plans Complete | Status | Completed |
|-------|-----------|----------------|--------|-----------|
| 1. FTP Endpoint Contract & Routing | v1.0 | - | Complete | 2026-04-27 |
| 2. FTP Driver Backend | v1.0 | - | Complete | 2026-04-27 |
| 3. One-Way FTP Sync Flows | v1.0 | - | Complete | 2026-04-27 |
| 4. FTP Verification & Discoverability | v1.0 | - | Complete | 2026-04-27 |
| 5. Public FTPSyncService API Contract | v2.0 | 3/3 | Complete | 2026-04-27 |
| 6. One-Shot Disk<->FTP Sync Through Library API | v2.0 | 3/3 | Complete | 2026-04-27 |
| 7. Background Disk->FTP Lifecycle | v2.0 | 3/3 | Complete | 2026-04-27 |
| 8. FTP-Only Package Reduction | v2.0 | 4/4 | Complete | 2026-04-28 |
| 9. Verification, Examples, and Migration Docs | v2.0 | 4/4 | Complete | 2026-04-28 |

## Next

No active milestone is currently planned. Start the next cycle with `/gsd-new-milestone` to define fresh requirements and roadmap scope.
