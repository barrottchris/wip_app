---
title: App card layout and status fix
status: completed
owner: copilot
last_updated: 2026-09-02
---

# Task

## Context
- Product context is the local-first app registry described in the docs in this repo.
- The card view is the primary entry point for app metadata, git state, and start/stop controls.
- The UI is implemented as static frontend JS with CSS in the frontend folder.

## Problem / goal
- The live output panel is being visually tied to the app controls row instead of staying above them.
- The app running status is appearing twice in the card and is making the right-side status area feel noisy.
- The app action buttons should remain anchored to the left without being displaced by the terminal panel.

## Questions to answer
- None at this stage; the desired layout is specific and consistent with the app card design brief already discussed.

## Requirements
- The live output panel must remain visually above the app controls and should not shift the app action row.
- The running status must appear once in the card's dedicated status area.
- The app control buttons must remain on the left side of the card.
- The URL should remain associated with the running status panel instead of the live-output panel.

## Proposed implementation plan
1. Remove the duplicate status block from the live-output rendering path.
2. Keep the terminal panel as a dedicated row above the app buttons.
3. Adjust the card CSS so the layout no longer reflows the app actions when the log panel is present.
4. Verify the card renders without duplicate status badges and the layout structure remains stable.

## Acceptance criteria
- Only one running status is rendered for the card.
- The terminal panel sits above the app buttons and does not push the controls down.
- The right-side status panel still shows the running state and URL when relevant.

## Status log
- `in-progress` — 2026-09-02 — started card layout and status cleanup.
- `in-progress` — 2026-09-02 — moved the terminal into the header grid so it shares the top row with the app title.
- `completed` — 2026-09-02 — placed connections and controls in the independent left column and removed the duplicate live-row status path.