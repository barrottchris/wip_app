---
title: Terminal log auto-update
status: completed
owner: copilot
last_updated: 2026-09-02
---

# Task

## Context
- WIP is a local-first app launcher whose MVP includes independently managed component processes and runtime output, per the repository planning docs.
- The registry card currently renders the last few component log lines from the initial app payload.
- The existing card refresh path fetches the app again and replaces the whole card, which is unsuitable for continuously displaying process output.

## Problem / goal
- New process output does not appear automatically in the terminal panel.
- Add incremental log polling that updates the existing panel in place without navigating or replacing the app card/page.

## Questions to answer
- Use a 1-second polling interval for running components unless the user prefers another cadence.
- Poll only while the component is running; stop polling when the card is detached or the component stops.
- On a polling failure, keep the current output and retry on the next interval without showing repeated alerts.

## Requirements
- Fetch the current app/component runtime state through the existing API layer.
- Update only the terminal log container and running-state placeholder when new data arrives.
- Update the app card's URL link in place when runtime log detection discovers a browse URL.
- Keep the log panel bounded to the existing recent-lines behavior and preserve scroll usability.
- Do not call navigation or replace the entire app card during automatic log updates.
- Avoid leaving timers active after a card is removed or a component stops.
- Preserve explicit full-card refreshes for start, stop, git refresh, and other existing actions.

## Proposed implementation plan
1. Add a focused API helper or reuse the existing app lookup for runtime polling.
2. Give the terminal panel a stable update function and start a card-local polling timer for running components.
3. Stop the timer when the card is detached or the component is no longer running.
4. Add or extend a focused frontend regression test for in-place updates and timer cleanup where the repository test setup supports it.
5. Run the focused Go/frontend tests and update this spec with the validation result.

## Acceptance criteria
- New log lines appear in the terminal app card automatically while its component runs.
- The page URL, registry container, and app card DOM identity remain unchanged during polling.
- Polling stops after the card is removed or the process stops.
- Polling errors do not disrupt the current card or produce an alert loop.
- A detected runtime URL appears in the card without a page or card refresh.
- Existing start/stop and explicit refresh actions continue to work.

## Status log
- 2026-09-02 — `todo` — spec created from the terminal card's snapshot-only log rendering; awaiting confirmation of polling defaults
- 2026-09-02 — `in-progress` — user confirmed one-second in-place polling with silent retry and lifecycle cleanup
- 2026-09-02 — `completed` — terminal panel now polls runtime logs in place, preserves scroll position, and stops when detached or no longer running; command execution was unavailable in this session, so Go/frontend tests remain to be run locally
- 2026-09-02 — `in-progress` — extend the same polling response to update the detected browse URL in the existing card status panel
- 2026-09-02 — `completed` — detected browse URLs now appear or disappear in the existing card status panel during polling; source validation confirmed updated polling signatures and URL slot wiring, but executable tests remain unavailable in this session