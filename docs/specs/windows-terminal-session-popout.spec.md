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
- The terminal UI is a browser popup because the current app has no native window host. It connects to the managed session's stdin and captured output.

## Problem / goal
- The current fix opens a visible Windows command prompt automatically when a component starts.
- The desired UX is for `Start` to run the component silently and for `Open terminal` to be the explicit popout action.
- Windows cannot attach a new `cmd.exe` window to an already-running hidden process. The approved solution is an attached browser terminal view backed by the managed session.

## Questions to answer
- Confirmed: use an attached browser terminal view for the current architecture; it must not start the command again.

## Requirements
- Starting a native component must create one managed ConPTY session without automatically opening a visible terminal window.
- Clicking `Open terminal` must open a terminal view connected to the existing ConPTY session, in the component's app directory and with the component name as its title.
- WIP must continue to know whether the managed process is running or has exited, capture output where technically possible, and stop the same process tree.
- Opening the terminal for an already-running component must not execute its start command again.
- Invalid paths, missing start commands, unsupported platforms, and unavailable terminal sessions must return a clear error to the UI.
- The solution must remain local-only and preserve independent component lifecycle management.
- Add focused Go tests for command construction and lifecycle/launch behavior that can run without requiring an interactive desktop; validate the Windows behavior with the available smoke test.

## Proposed implementation plan
1. Confirmed the ConPTY-based attached-session approach.
2. Choose and implement the terminal host that can render and interact with the ConPTY stream.
3. Move terminal/session ownership into the process manager so `start`, `terminal`, and `stop` address the same process session.
4. Change the terminal route and frontend action to open the attached terminal view without rerunning the configured command.
5. Add focused tests for no-duplicate-launch behavior, command construction, and error cases.
6. Run focused Go tests and the local smoke test, then update this spec with the final behavior and validation.

## Acceptance criteria
- Clicking `Start` does not automatically open a terminal window.
- Clicking `Open terminal` opens a terminal view attached to the existing session and does not start a duplicate process.
- Clicking `Stop` terminates the same managed session and its child processes, and the portal reflects the stopped state.
- Session exit updates runtime state and remains visible through the existing log/status UI.
- Focused tests and the local smoke test pass, and this spec records the completed implementation.

## Status log
- 2026-09-02 — `todo` — spec created from the current split terminal/process implementation; awaiting confirmation of visible-console start behavior and fallback semantics
- 2026-09-02 — `in-progress` — user approved visible `cmd.exe` ownership at start, reuse without duplicate launch, and an error for sessions without an attachable terminal
- 2026-09-02 — `in-progress` — user confirmed the ConPTY-based attached-session approach; terminal-host choice remains an implementation detail to resolve