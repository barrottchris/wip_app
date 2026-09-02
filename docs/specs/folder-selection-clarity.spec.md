---
title: Folder selection clarity in add app flow
status: completed
owner: copilot
last_updated: 2026-09-02
---

# Task

## Context
- WIP is designed as a local, single-user app registry for organizing in-progress apps, per `docs/overview.md`, `docs/mvp-scope.md`, and `docs/architecture.md`.
- The onboarding flow in `frontend/src/js/pages/addApp.js` is the point where a user selects a folder and confirms the app location before creating the app record.
- The existing flow showed the folder browser inline with form fields and did not present a clear, dedicated “selected folder” summary, which made the action feel loose and not obviously persisted.

## Problem / goal
- Users were not given a clear visual container for folder selection.
- The currently selected folder was not prominent enough, so it was hard to tell what would be saved and when the choice was committed.
- The goal is to make the add-app flow read as a deliberate, clearly bounded “Folder selection” step before the user creates the app.

## Questions to answer
- None required; the UX issue is direct and the intended behavior is already clear from the onboarding flow and product intent.

## Requirements
- The folder selection controls must sit inside a dedicated visual panel with a clear section heading.
- The selected folder must be explicitly displayed in a summary area before the user creates the app.
- The folder browser must still allow browsing and selecting a path in a straightforward way.
- The selected path must remain the source of truth for the app creation flow.
- The change must preserve the existing app onboarding model and local-first architecture.

## Proposed implementation plan
1. Review the add-app page and folder picker component to identify the root cause of the loose flow.
2. Wrap the folder chooser in a dedicated panel and add a visual “Selected folder” summary.
3. Clarify the action label on the folder browser so it reads as a deliberate save/confirm step.
4. Refine the styling to make the section feel intentional and distinct from the surrounding form inputs.
5. Validate the relevant frontend and app tests.

## Acceptance criteria
- The add-app page shows a dedicated folder-selection panel instead of a loosely embedded browser.
- The user can clearly see the selected folder before clicking Add app.
- The folder selection UI still supports browsing and choosing a folder path.
- The relevant repository tests pass after the UX adjustment.

## Status log
- 2026-09-02 — `completed` — folder selection was made explicit in the add-app flow and verified with the relevant tests.
