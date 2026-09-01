---
title: Live process session and browse URL support
status: completed
owner: copilot
last_updated: 2026-09-01
---

# Task

## Context
- WIP is a local-first app registry for in-progress apps, as described in [docs/overview.md](../overview.md) and [docs/architecture.md](../architecture.md).
- The MVP goal is to organize apps, track git state, and start/stop app components from one place without hunting for commands or losing track of work, per [docs/mvp-scope.md](../mvp-scope.md).
- The runtime model is meant to reflect actual local process execution, and the data model in [docs/data-model.md](../data-model.md) expects components to be runnable units with start/stop commands and a runtime status.
- The live implementation now binds start/stop actions to real OS process sessions and surfaces output and inferred local URLs to the portal.

## Problem / goal
- The portal previously showed a component as running without giving the user any visible process/session evidence.
- The user needed to know whether a started component was actually attached to a live terminal session and whether that session had exited.
- The finished workflow also makes it easy to open a running app directly via a local URL once the process is up.
- This is a requirement for trust and operability in a local app launcher.

## Questions to answer
- Should the portal show an inline console/log panel for each running component, or should it launch a visible terminal window that stays associated with the component session?
- Should the component session be tracked by a live OS process handle plus a bounded log buffer, or do we need a dedicated “terminal session” object in the app model?
- For the app URL link, should the UI display a single inferred browse URL, or multiple candidate URLs if a command exposes more than one local port?

## Requirements
- The running state reflects an actual live process session, not just a boolean toggled on the server.
- If a session exits or the terminal closes, the portal updates the UI state accordingly.
- The user can see a clear process/session output or terminal-like log while a component is running.
- When a component exposes a known local app URL, the portal shows it in an obvious way.
- The implementation preserves the local-first, single-user MVP architecture already documented in the repo.
- The feature is validated with the relevant repo tests and a successful local smoke test.

## Proposed implementation plan
1. Confirm the live-session UX as an inline log panel rather than a separate terminal window.
2. Bind a component to a real spawned process and a live output stream.
3. Update the runtime state so the portal knows when the session exits or is closed.
4. Add clear browse URL inference and display it on running components.
5. Validate with the focused Go tests and a live smoke test for a real process.
6. Close the task once the checks confirm the behavior.

## Acceptance criteria
- A running component is associated with a tracked process session whose termination updates the UI status.
- The portal provides an inline output window for the active session.
- Running components display an obvious local browse URL when one can be inferred.
- The implementation is validated by relevant focused tests and a successful local smoke test.

## Status log
- 2026-09-01 — `todo` — spec created to capture the live terminal/session UX and browse URL requirement
- 2026-09-01 — `in-progress` — implementation aligned to the inline log-panel approach and process lifecycle binding
- 2026-09-01 — `completed` — live process tracking and output/URL display verified in the implemented runtime and UI flow
