---
title: Windows terminal session popout
status: in-progress
owner: copilot
last_updated: 2026-09-02
---

# Task

## Context
- WIP is a local-first Windows app launcher whose MVP includes independently managed component processes, per [docs/overview.md](../overview.md), [docs/architecture.md](../architecture.md), and [docs/mvp-scope.md](../mvp-scope.md).
- Components are runtime-managed units with start and stop commands; runtime state is not persisted, per [docs/data-model.md](../data-model.md).
- The current backend tracks the process started by `ProcessManager.Start`, captures its output, and exposes a `terminal` route that separately starts `cmd.exe /K` with the same command.
- The current frontend's `Open terminal` action calls that separate route, so it cannot display or attach to the already-managed process session.

## Problem / goal
- The terminal popout does not provide a usable local Windows terminal for the component's existing managed session.
- Starting a component and opening its terminal currently create two independent executions of the start command, which can cause duplicate servers, port conflicts, and a terminal window unrelated to the process WIP reports as running.
- Make the component terminal a visible Windows `cmd.exe` session that is the same OS process session WIP manages, while retaining portal status, output, and stop behavior.

## Questions to answer
- Confirmed: clicking `Start` launches the component in a visible `cmd.exe` window, and `Open terminal` reuses that existing session rather than starting the command a second time.
- Confirmed: a component started without an attachable visible session should report that no attachable terminal exists.
- Confirmed: the configured command line is run in the `cmd` prompt, including shell operators and redirects.

## Requirements
- Starting a native component must create one managed Windows console session, not one hidden process plus a second process when the terminal button is clicked.
- The visible terminal must open in the component's app directory and use the component name as its window title.
- WIP must continue to know whether the managed process is running or has exited, capture output where technically possible, and stop the same process tree.
- Opening the terminal for an already-running component must not execute its start command again.
- Invalid paths, missing start commands, unsupported platforms, and unavailable terminal sessions must return a clear error to the UI.
- The solution must remain local-only and preserve independent component lifecycle management.
- Add focused Go tests for command construction and lifecycle/launch behavior that can run without requiring an interactive desktop; validate the Windows behavior with the available smoke test.

## Proposed implementation plan
1. Confirmed the visible-console session behavior and fallback for sessions that cannot be attached.
2. Move the terminal/session ownership into the process manager so `start`, `terminal`, and `stop` address the same process session.
3. Use Windows-specific process launch behavior to create or reveal a `cmd.exe` window while preserving the managed process handle and output/status tracking.
4. Change the terminal route and frontend action to reopen the existing session without rerunning the configured start command.
5. Add focused tests for no-duplicate-launch behavior, command construction, and error cases.
6. Run focused Go tests and the local smoke test, then update this spec with the final behavior and validation.

## Acceptance criteria
- Clicking `Start` opens one visible Windows command prompt in the app directory and starts the configured component there.
- Clicking `Open terminal` for that component reuses or brings forward the existing session and does not start a duplicate process.
- Clicking `Stop` terminates the same managed session and its child processes, and the portal reflects the stopped state.
- Session exit updates runtime state and remains visible through the existing log/status UI.
- Focused tests and the local smoke test pass, and this spec records the completed implementation.

## Status log
- 2026-09-02 — `todo` — spec created from the current split terminal/process implementation; awaiting confirmation of visible-console start behavior and fallback semantics
- 2026-09-02 — `in-progress` — user approved visible `cmd.exe` ownership at start, reuse without duplicate launch, and an error for sessions without an attachable terminal