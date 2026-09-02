---
title: SQLite activity audit trail and Activity page
status: todo
owner: copilot
last_updated: 2026-09-02
---

# Task

## Context
- WIP is a local-first Windows app for organizing projects, surfacing git state, and starting or stopping app components, as described in [docs/overview.md](../overview.md), [docs/architecture.md](../architecture.md), and [docs/mvp-scope.md](../mvp-scope.md).
- App metadata already persists in SQLite, while component running state is intentionally computed from live process state and is not persisted, per [docs/data-model.md](../data-model.md).
- The Activity route currently renders a placeholder in `frontend/src/js/pages/placeholders.js`.
- Relevant event-producing boundaries already exist in the Go server and store: app creation/editing, archive/unarchive, git refresh, and component start/stop.

## Problem / goal
- Replace the Activity placeholder with a useful operational history of work across tracked apps.
- Persist an audit trail in the existing local SQLite database so activity survives restarts and can be queried independently from current app state.
- Show enough context to answer what happened, to which app, on which branch, with what build/check context when available, and whether the operation succeeded or failed.

## Questions to answer
- Should the first release record only WIP actions (app edits, git refresh, start/stop, archive/unarchive), or also poll/import external git commits? Proposed default: record WIP actions first; show git commit activity only when it is already available from an explicit git refresh.
- What does “build” mean in the first release? Proposed default: record a build/check field only for an explicit WIP build action; do not infer builds from start commands or git commits.
- Should failed actions be retained? Proposed default: yes, because failures are important audit and troubleshooting history.
- How much history should the default page load? Proposed default: the newest 100 events, with a simple “load more” path and filters for app, event type, branch, and outcome.

## Requirements
- Add a SQLite-backed `activity_events` record with a stable ID, timestamp, app ID and app name snapshot, event type, summary, branch snapshot, build/check snapshot, lifecycle status snapshot, runtime status snapshot, change summary, outcome, and optional detail/error data.
- Preserve event snapshots where practical so renamed apps, changed branches, and later runtime state do not rewrite historical meaning.
- Record successful and failed events at the server-side operation boundary, including component start/stop and relevant app/git actions.
- Do not treat current component running state as persisted app metadata; capture it as event context only.
- Expose a read-only API endpoint for newest-first activity, with bounded pagination and server-side filters.
- Replace the placeholder with a chronological Activity view grouped by day. Each row should make app, branch, action, outcome, and relative time scannable without opening a detail view.
- Provide filters for app, event type, outcome, and branch where data exists, plus a clear empty state and loading/error states.
- Selecting an event should reveal a detail panel containing exact timestamp, full change/build/status/runtime context, and error/detail text when present.
- Keep the visual language consistent with the existing native-app frontend while making the page operational and dense rather than a marketing dashboard.
- Add focused Go persistence/API tests and frontend rendering/interaction coverage using the repository's existing test setup.

## Proposed implementation plan
1. Confirm the event taxonomy, build meaning, retention/pagination defaults, and proposed Activity layout.
2. Add the activity schema and store methods for append and filtered newest-first reads.
3. Add a small event-recording service/helper and wire it into the existing server action boundaries, including failure paths.
4. Add the Activity API contract and frontend API helper.
5. Replace the placeholder with the grouped feed, filters, event detail panel, and loading/empty/error states.
6. Add focused regression tests, run the relevant Go and frontend checks, and update this spec with validation results.

## Acceptance criteria
- Activity records are stored in the existing SQLite database and survive an application restart.
- Starting, stopping, refreshing git, editing, and archiving/unarchiving an app create accurate success or failure records at the chosen event boundaries.
- The Activity page loads newest-first records and visibly identifies the app, branch when known, action, outcome, and time.
- Event details show build/check context and change/error detail when available without fabricating values.
- Filters and bounded pagination work without exposing archived app data incorrectly or making unbounded database queries.
- The page has deliberate loading, empty, error, and detail states and works at desktop and narrow window widths.
- Focused persistence, API, and frontend tests pass.

## Status log
- 2026-09-02 — `todo` — spec drafted from the Activity placeholder and existing SQLite/server action boundaries; awaiting confirmation of the open questions and proposed UX.