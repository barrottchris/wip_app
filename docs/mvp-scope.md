# WIP — MVP Scope

## Goal for v1

> "I can see all my in-progress apps in one organized place, each properly under git, and I can start or stop any of them with one click — without hunting for commands or losing track of what's half-done."

If v1 achieves that, it's a success — everything else is deliberately deferred.

## In Scope for MVP

### 1. App Registry
- A single managed location (replacing scattered C-drive folders) where every in-progress app lives.
- Each app entry has rich metadata:
  - Name
  - Tech stack
  - Purpose / description
  - Status (e.g. active, paused, abandoned, shipped — exact set TBD during design)
  - Notes
  - Git repo link / branch info

### 2. Git Integration
- Apps not currently under git get brought into it as part of onboarding to WIP.
- Branch awareness — an app's various branches are visible and tracked, not just its main line.
- (Open question, non-blocking for MVP design: how *actively* WIP manages git — e.g. surfacing status only, vs. offering actions like branch cleanup — can be decided during build.)

### 3. Start / Stop Controls
- Each app has one or more configurable **start commands**, covering its components (e.g. frontend, backend, docker-compose).
- Commands are plain text (exact storage format — flat config file vs. structured per-component — TBD during build; this is an implementation detail, not a scope decision).
- A corresponding **stop** action to cleanly shut an app's processes down.
- Multiple apps can run independently and simultaneously — starting/stopping one has no effect on others.
- Mix of execution styles supported: Docker containers where useful, native processes otherwise (per-app choice, not uniform).

### 4. Local App
- WIP runs as a local application to start (not a hosted web app yet).

## Explicitly Out of Scope for MVP (Parked for Later)

These are captured and valued, but deliberately excluded from v1 to avoid scope creep:

| Feature | Notes | Target phase |
|---|---|---|
| Brainstorming / idea page | Tree-based visual metaphor — idea starts as a seed, grows into a tree with branches (requirements, functionality, user base) and a health/status indicator showing dormancy or progress | v1.1 |
| Jira integration | One-way, manually-triggered sync from WIP to Jira (descriptions, epics, priority, status, comments); helps PMs track progress. Auth via each user's own AD account (per-user session, not shared service account) | Post-MVP |
| Confluence integration | README-to-Confluence-page mirroring, same one-way/manual sync model as Jira | Post-MVP |
| Company-wide / non-dev-friendly rollout | WIP evolving from personal local tool into something advertised and used across the company, simple enough for non-developers | Post-MVP, later stretch |
| WIP as a hosted web app | Long-term evolution from local app to a proper web app | Post-MVP, later stretch |

## Open Questions to Resolve Before/During Build

Non-blocking for this scope document, but worth revisiting as design gets more concrete:

- Exact status taxonomy for app entries (active/paused/abandoned/etc.)
- How multiple simultaneously-running apps are accessed (fixed ports to remember vs. a proxy/dashboard)
- Exact start/stop command storage format
- Whether Docker orchestration is WIP-generated (e.g. auto compose files) or simply calls existing per-app Dockerfiles/scripts
