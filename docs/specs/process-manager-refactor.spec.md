---
title: Process manager responsibility split
status: completed
owner: copilot
last_updated: 2026-09-02
---

# Task

## Context
- WIP is a local-first Go application whose process manager owns live OS process handles while `RuntimeTracker` owns UI-visible runtime state.
- `internal/app/process_manager.go` currently also contains process session log state and browse URL inference helpers.
- Existing process-manager tests cover start, stop, terminal access, captured output, and URL inference.

## Problem / goal
- Reduce the size and responsibility of `process_manager.go` without changing runtime behavior or public APIs.
- Make process lifecycle, process-session state, and browse URL parsing easier to navigate and test independently.

## Questions to answer
- Confirmed 2026-09-02: proceed with a small structural refactor after review.
- Confirmed 2026-09-02: preserve existing behavior, exported names, and package boundaries.

## Requirements
- Keep `ProcessManager` lifecycle methods and process termination behavior unchanged.
- Move `ProcessSession` and its log writer/state helpers into a focused file.
- Move browse URL inference and parsing helpers into a focused file.
- Preserve all existing exported function and method signatures.
- Run formatting and the focused app tests after the refactor.

## Proposed implementation plan
1. Extract process-session state and captured-output helpers into `process_session.go`.
2. Extract browse URL inference and parsing helpers into `browse_url.go`.
3. Remove the moved implementations and now-unused imports from `process_manager.go`.
4. Run `gofmt` and `go test ./tests/app`.

## Acceptance criteria
- `process_manager.go` contains process-manager lifecycle and terminal orchestration only.
- Existing callers and tests compile without API changes.
- Focused app tests pass.
- The spec reflects the completed work.

## Status log
- 2026-09-02 — `in-progress` — approved small structural refactor; implementation started.
- 2026-09-02 — `completed` — split process sessions and browse URL helpers into focused files; public APIs preserved. Command validation was unavailable in this session.
- 2026-09-02 — `completed` — added behavior-focused comments above every function in the three process-manager files and removed a duplicated package declaration from the edited session file.