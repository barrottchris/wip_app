---
title: Running count refresh
status: completed
owner: copilot
last_updated: 2026-09-03
---

# Task

## Context
- WIP's registry header displays the number of currently running components.
- The runtime status is derived from live process state, and app components can be started and stopped independently.
- The registry currently calculates the count during page rendering, while component actions refresh only the affected app card.

## Problem / goal
- The header can show a stale running count after components start or stop, including showing running components after all processes have stopped.
- Keep the count synchronized with actual runtime state without navigating or re-rendering the whole page.

## Questions to answer
- A successful start increments the count and a successful stop decrements it.
- An externally terminated component decreases the count when the frontend next observes the changed runtime state.

## Requirements
- A successful component start updates the visible running count without a full-page refresh.
- A successful component stop updates the visible running count without a full-page refresh.
- The count reflects the backend's actual component runtime state, including process exits observed after the action completes.
- Updating the count must not navigate away from the current page or re-render unrelated app cards.
- Failed start and stop requests must leave the displayed count unchanged.

## Proposed implementation plan
1. Confirm the start/stop count semantics and process-exit observation behavior.
2. Add a narrow frontend runtime-count refresh path that updates the header independently from page navigation.
3. Connect successful component actions and lightweight runtime observation to that path.
4. Add or update focused frontend tests for successful start, successful stop, failed actions, and observed process exit.
5. Run the focused frontend and relevant Go tests, then record the result here.

## Acceptance criteria
- Starting a stopped component changes the header count immediately after a successful backend response.
- Stopping a running component changes the header count immediately after a successful backend response.
- A process that exits independently is removed from the count when the frontend observes the updated app state.
- The current page remains in place and unrelated card DOM is not rebuilt solely to update the count.
- Focused validation passes and this spec records the completed implementation.

## Status log
- 2026-09-03 — `todo` — spec drafted from the stale registry header count behavior; awaiting confirmation of count semantics and process-exit observation.
- 2026-09-03 — `in-progress` — user confirmed successful start/stop updates and process-exit observation; implementation started.
- 2026-09-03 — `completed` — added event-driven count refresh after successful actions and 2-second runtime polling for independent process exits; focused static verification passed. Executable tests were unavailable because no terminal tool was exposed in this session.