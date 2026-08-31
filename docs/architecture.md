# WIP — Architecture Sketch (MVP)

Rough shape only, to support decisions — not a build spec yet.

## Shape

WIP runs as a **local application** with:

- A **frontend/UI** — the registry view (list/board of apps), app detail view, start/stop controls.
- A **local backend/service layer** — does the actual work: reading git status, running start/stop commands as local processes, reading/writing app metadata.
- **Local storage for metadata** — a lightweight local store (e.g. a local database file or structured config files) holding the App Entry / Branch / Component data from `data-model.md`. Not the app code itself — just WIP's own records about each app.
- **The managed project storage area** — a defined location on the C drive where tracked apps' code actually lives (or is linked to, if apps stay in their current location — TBD).

## Key Interactions

1. **Onboarding an app** — WIP is pointed at an existing folder (or a new one), checks/initializes git, and creates an App Entry with metadata the user fills in.
2. **Viewing the registry** — WIP reads its local metadata store and displays apps, with git status (branch, last commit) pulled live or cached from each repo.
3. **Starting a component** — WIP runs the configured start command as a local process (or triggers `docker` if `run_mode` is `docker`), and tracks that it's running.
4. **Stopping a component** — WIP terminates the corresponding process/container cleanly.
5. **Running multiple apps concurrently** — each app's components are tracked/managed independently; starting or stopping one has no effect on others' running state.

## Deliberately Deferred Architecture Decisions

These don't block MVP design and can be resolved during build:

- **Process management approach** — how WIP tracks and cleanly terminates started processes (e.g. tracking PIDs directly vs. using a process manager).
- **Docker orchestration style** — WIP-generated compose files vs. simply invoking existing Dockerfiles/compose files already in each repo.
- **Access to multiple running apps** — fixed/remembered ports vs. some kind of proxy or unified dashboard linking out to each running app.
- **Metadata storage format** — simple local database vs. structured files (e.g. one config file per app) checked in alongside the app itself.

## Explicitly Not in MVP Architecture

- No network-facing service, hosted backend, or multi-user auth — this is a single-user local tool for v1.
- No Jira/Confluence API integration layer yet.
- No idea-brainstorming data structures or UI (the seed-to-tree concept) — that's a v1.1 addition once the core loop is working.

## Why This Shape

Keeping WIP itself local-only and single-user for MVP avoids the exact trap it's designed to solve: taking on auth, hosting, and multi-user complexity before the core value (organized storage, git tracking, one-click start/stop) is even validated in daily use.
