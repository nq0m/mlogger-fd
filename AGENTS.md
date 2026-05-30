<!-- GSD:project-start source:PROJECT.md -->
## Project

**Field Day Logger**

A lightweight, mobile-friendly, multi-user web-based logging application purpose-built for ARRL Field Day. Operators on separate devices log QSOs to a shared local server (Raspberry Pi or laptop) over a LAN. Designed for tent-based deployment with offline-first resilience — logging continues uninterrupted when WiFi or power drops.

**Core Value:** Operators can log QSOs even when the network goes down, with all data syncing automatically when reconnected. The log must never be lost.

### Constraints

- **Offline resilience**: Must log without server connectivity, sync when reconnected
- **Hardware**: Runs on Raspberry Pi 4 (4GB) or old Linux laptop
- **Storage**: SQLite single-file database, WAL mode
- **Network**: LAN-only, no internet dependency, no CORS needed
- **UI**: Must work on phones/tablets with gloves or wet/dirty fingers
- **Deployment**: Single Go binary + static SPA files, systemd service
- **Auth**: None for trusted LAN; simple shared password if open WiFi
- **Browser**: Modern browsers with IndexedDB and Service Worker support
<!-- GSD:project-end -->

<!-- GSD:stack-start source:STACK.md -->
## Technology Stack

Technology stack not yet documented. Will populate after codebase mapping or first phase.
<!-- GSD:stack-end -->

<!-- GSD:conventions-start source:CONVENTIONS.md -->
## Conventions

Conventions not yet established. Will populate as patterns emerge during development.
<!-- GSD:conventions-end -->

<!-- GSD:architecture-start source:ARCHITECTURE.md -->
## Architecture

Architecture not yet mapped. Follow existing patterns found in the codebase.
<!-- GSD:architecture-end -->

<!-- GSD:skills-start source:skills/ -->
## Project Skills

No project skills found. Add skills to any of: `.claude/skills/`, `.agents/skills/`, `.cursor/skills/`, `.github/skills/`, or `.codex/skills/` with a `SKILL.md` index file.
<!-- GSD:skills-end -->

<!-- GSD:workflow-start source:GSD defaults -->
## GSD Workflow Enforcement

Before using Edit, Write, or other file-changing tools, start work through a GSD command so planning artifacts and execution context stay in sync.

Use these entry points:
- `/gsd-quick` for small fixes, doc updates, and ad-hoc tasks
- `/gsd-debug` for investigation and bug fixing
- `/gsd-execute-phase` for planned phase work

Do not make direct repo edits outside a GSD workflow unless the user explicitly asks to bypass it.
<!-- GSD:workflow-end -->



<!-- GSD:profile-start -->
## Developer Profile

> Profile not yet configured. Run `/gsd-profile-user` to generate your developer profile.
> This section is managed by `generate-claude-profile` -- do not edit manually.
<!-- GSD:profile-end -->
