---
title: Folder selection clarity and direct path lookup
status: in-progress
owner: copilot
last_updated: 2026-09-03
---

# Task

## Context
- WIP is designed as a local, single-user app registry for organizing in-progress apps, per `docs/overview.md`, `docs/mvp-scope.md`, and `docs/architecture.md`.
- The onboarding flow in `frontend/src/js/pages/addApp.js` and the existing-app edit flow in `frontend/src/js/pages/appDetail.js` are the points where a user selects or changes an app folder.
- The existing flow showed the folder browser inline with form fields and did not present a clear, dedicated “selected folder” summary, which made the action feel loose and not obviously persisted.
- The folder browser currently requires navigating one directory at a time, although the existing browse endpoint can already resolve a supplied path.

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
- The user must be able to enter a folder path directly in a text field.
- Submitting a valid path must browse to that location and show its contents in the existing folder browser.
- Invalid or inaccessible paths must produce a clear local error state without losing the rest of the form.
- The selected path must remain the source of truth for the app creation flow.
- The change must preserve the existing app onboarding model and local-first architecture.

## Proposed implementation plan
1. Add a path input to the dedicated folder-selection panel.
2. On the confirmed lookup trigger, call the existing browse endpoint with the entered path and render the result.
3. Keep browsing and explicit folder confirmation working with the direct-path flow.
4. Add clear loading and invalid-path feedback without clearing the current selection prematurely.
5. Validate the relevant frontend and app tests.

## Acceptance criteria
- The add-app page shows a dedicated folder-selection panel instead of a loosely embedded browser.
- The user can clearly see the selected folder before clicking Add app.
- The folder selection UI still supports browsing and choosing a folder path.
- A directly entered valid path opens in the browser and can be confirmed as the selected folder.
- The existing-app edit page offers the same direct lookup and explicit confirmation when changing folders.
- An invalid path gives clear feedback and does not silently select it.
- The relevant repository tests pass after the UX adjustment.

## Status log
- 2026-09-02 — `completed` — folder selection was made explicit in the add-app flow and verified with the relevant tests.
- 2026-09-03 — `todo` — direct folder-path lookup requested; awaiting confirmation of lookup trigger.
- 2026-09-03 — `in-progress` — approved: add a Find folder button that validates and browses the entered path before selection can be confirmed.
- 2026-09-03 — `in-progress` — extended direct lookup to the existing app edit page.
