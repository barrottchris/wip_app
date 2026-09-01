---
title: Live process session and browse URL support
status: todo
owner: copilot
last_updated: 2026-09-01
---

# Task

## Context
- WIP is a local-first app registry for in-progress apps, as described in `docs/overview.md` and `docs/architecture.md`.
- The MVP goal is to organize apps, track git state, and start/stop app components from one place without hunting for commands or losing track of work, per `docs/mvp-scope.md`.
- The runtime model is meant to reflect actual local process execution, and the data model in `docs/data-model.md` expects components to be runnable units with start/stop commands and a runtime status.
- The current implementation starts child processes via Go's `exec.Command`, but it tracks only a boolean running flag and does not link the terminal session to the portal or surface logs or app URLs.

## Problem / goal
- The portal currently shows a component as running without giving the user any visible process/session evidence.
- The user needs to know whether a started component is actually attached to a live terminal session and whether that session has exited.
- The finished workflow should also make it easy to open a running app directly via a local URL once the process is up.
- This is not just a UI polish issue: it is a requirement for trust and operability in a local app launcher.

## Questions to answer
- Should the portal show an inline console/log panel for each running component, or should it launch a visible terminal window that stays associated with the component session?
- Should the component session be tracked by a live OS process handle plus a bounded log buffer, or do we need a dedicated “terminal session” object in the app model?
- For the app URL link, should the UI display a single inferred browse URL, or multiple candidate URLs if a command exposes more than one local port?

## Requirements
- The running state must reflect an actual live process session, not just a boolean toggled on the server.
- If a session exits or the terminal closes, the portal must update the UI state accordingly.
- The user must be able to see a clear process/session output or terminal link while a component is running.
- When a component exposes a known local app URL, the portal must show it in an obvious way.
- The implementation must preserve the local-first, single-user MVP architecture already documented in the repo.
- The feature must be implemented incrementally and validated with the relevant repo tests before completion.

## Proposed implementation plan
1. Confirm the desired live-session UX: visible terminal vs. log panel vs. both.
2. Add a process/session tracking model that binds a component to a real spawned process and a live output stream.
3. Update the runtime state so the portal knows when the session exits or is closed.
4. Add a clear browse URL inference and display it on running components.
5. Validate with the focused Go tests and a live smoke test for a real process.
6. Update this spec to `completed` once verified.

## Acceptance criteria
- A running component is associated with a tracked process session whose termination updates the UI status.
- The portal provides a visible way to monitor output or access the active terminal/session.
- Running components display an obvious local browse URL when one can be inferred.
- The implementation is validated by relevant focused tests and a successful local smoke test.

## Status log
- 2026-09-01 — `todo` — spec created to capture the live terminal/session UX and browse URL requirement
