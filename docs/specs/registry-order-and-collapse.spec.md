---
title: Registry app ordering and collapsed cards
status: completed
owner: copilot
last_updated: 2026-09-03
---

# Task

## Context
- WIP is a local-first Windows app registry for managing and resuming in-progress apps.
- The registry page currently renders apps in the backend's `last_touched_at` order.
- App cards already expose the app name, lifecycle/runtime status, metadata, connections, terminal output, and component controls.
- App metadata is persisted in the local SQLite store; runtime component state is computed and must remain live.
- Relevant files are `frontend/src/js/pages/registry.js`, `frontend/src/js/components/appCard.js`, `frontend/src/js/api.js`, `internal/app/types.go`, `internal/app/store.go`, `internal/httpapi/apps.go`, and the frontend/Go tests.

## Problem / goal
- Let the user arrange active registry apps in a deliberate order using animated drag-and-drop.
- Let the user collapse an app card so the registry shows only its name and status, making a long registry easier to scan.

## Questions to answer
- Resolved: app order persists across application restarts.
- Resolved: use animated drag-and-drop reordering; explicit movement buttons are not required for the primary interaction.
- Resolved: the first/last card has no valid drop target beyond the list boundary.
- Resolved: collapsed state is temporary registry-view state and is not persisted.
- Resolved: cards are expanded by default.
- Resolved: ordering affects only the active Registry page; Archived remains a separate view.

## Requirements
- Add a stable persisted ordering value to app entries, with a deterministic fallback for existing records.
- Return active apps in the saved registry order while preserving a deterministic order for ties/new records.
- Add an API operation to save the active registry order, validating app IDs and keeping archived apps out of the active ordering.
- Add animated drag-and-drop behavior that visibly moves cards while dragging and saves the resulting order.
- Re-render or reconcile the registry after a successful order save and preserve live status behavior.
- Add an accessible card collapse/expand control with an understandable label and state.
- When collapsed, a card must show the app name and lifecycle/runtime status while hiding the detailed metadata, connections, terminal output, and component controls.
- Keep the card order and collapse interaction usable at narrow window widths.
- Add focused Go and frontend regression coverage for persistence/order, boundary movement, and collapsed-card rendering/state.

## Proposed implementation plan
1. Confirm the interaction and persistence decisions above.
2. Add the persisted ordering field and migration-safe schema/store behavior, including list ordering and one-step move logic.
3. Add the narrow HTTP/API client surface for saving registry order.
4. Add animated drag-and-drop ordering and collapse/expand state to the existing card/registry rendering path.
5. Add focused tests, run the relevant Go tests and frontend checks, and verify existing app-card behavior is unchanged.
6. Record validation results and mark this spec `completed`.

## Acceptance criteria
- The user can drag any active registry app to another position and see the cards animate/reflow during the interaction.
- The selected order survives reload and application restart.
- The registry has no invalid beyond-the-boundary drop target and archived apps are not reordered.
- The user can collapse and expand each card independently.
- A collapsed card visibly retains the app name and current status, including runtime running state where applicable.
- Existing start/stop, terminal, folder, Git, app-detail, and archived-app interactions continue to work.
- Focused regression tests pass, with any unavailable command validation documented.

## Status log
- 2026-09-03 — `todo` — spec drafted from the registry render path and persisted app model.
- 2026-09-03 — `in-progress` — user approved persisted drag-and-drop ordering, expanded-by-default cards, and non-persisted collapse state.
- 2026-09-03 — `completed` — added SQLite-backed ordering with migration support, validated active-app order saves, animated drag-and-drop registry cards, and per-card collapse/expand state. Added focused store and frontend regression coverage; executable tests were unavailable because no command runner was exposed in this session.