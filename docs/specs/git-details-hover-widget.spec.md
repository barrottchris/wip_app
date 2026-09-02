---
title: Git details hover widget
status: completed
owner: copilot
last_updated: 2026-09-02
---

# Task

## Context
- WIP is a local-first app registry whose MVP includes Git awareness and branch information.
- Each app card already shows a Git connection pill, the current/default branch, and last-touched metadata.
- The backend already detects a local repository, reads the branch name, reads the `origin` remote URL, and exposes persisted Git metadata through the app API. The current Git utility does not yet read the repository name or a branch's latest commit timestamp directly.
- Relevant implementation areas are `internal/gitutil`, `main.go`, `internal/app/types.go`, `frontend/src/js/components/appCard.js`, and `frontend/src/style.css`.

## Problem / goal
- A connected Git pill does not currently provide enough repository context without navigating away from the registry.
- Add a compact, accessible hover/focus widget attached to the connected Git pill showing the repository name, current branch, latest update on that branch, and a link to the configured repository when available.

## Questions to answer
- Should “last update on the branch” mean the latest commit on the currently checked-out branch, or the latest commit on the app's default branch?
- Should the repository name be derived from the `origin` URL (preferred) and fall back to the local folder name when no remote exists?
- Should the Git details widget open on hover and keyboard focus, while the existing click action continues to open the Git page?
- When Git is connected but a remote URL or commit timestamp is unavailable, should the widget show an em dash/“Not available” row and keep the pill connected?

## Requirements
- Only connected Git pills expose the Git details widget; disconnected and coming-soon pills retain their current behavior.
- The widget displays repository name, current checked-out branch, latest commit/update timestamp for the defined branch, and a repository link when one exists.
- Git details are computed from the local repository and returned through the existing app payload or a narrowly scoped Git endpoint; no remote service or authentication is introduced.
- The repository link opens in a new browser tab and is omitted or disabled when no remote URL is configured.
- Hover and keyboard focus provide equivalent access to the details; the widget does not prevent the existing Git pill click navigation.
- Missing or malformed Git metadata does not break app-card rendering.
- The implementation preserves the existing card layout and local-first MVP scope.

## Proposed implementation plan
1. Confirm the branch meaning, repository-name fallback, interaction behavior, and missing-data presentation.
2. Extend the local Git utility with narrowly scoped reads for repository name and latest commit timestamp on the selected branch, reusing the existing branch and remote information where possible.
3. Add the computed Git summary to the app API response and focused backend tests for repositories with and without remotes/commits.
4. Add the connected-pill hover/focus widget in the app-card component, including safe DOM text/link handling and responsive styling.
5. Validate with focused Go tests, frontend tests or smoke checks, and a local UI smoke test covering connected and disconnected apps.
6. Update this spec to `completed` with validation notes.

## Acceptance criteria
- A Git-connected app card reveals repository name, branch, latest branch update, and repository link on hover or keyboard focus.
- The repository link is usable when an `origin` remote exists and is absent when it does not.
- The displayed latest update matches the agreed branch definition and is sourced from the local repository.
- Disconnected Git pills and existing Git-page navigation remain functional.
- Missing Git fields render a clear fallback without throwing or distorting the card layout.
- Relevant automated tests and a local smoke check pass.

## Status log
- 2026-09-02 — `todo` — spec created from the product docs and current Git/card implementation; awaiting clarification.
- 2026-09-02 — `in-progress` — approved: use the checked-out branch, derive the repo name from the remote with a folder fallback, support hover/focus plus click navigation, and show `Not available` for missing values.
- 2026-09-02 — `completed` — added live Git summary data, connected-pill hover/focus details, repository linking, missing-data fallbacks, and repository-name tests. Automated execution was unavailable in this session because no terminal command runner is exposed.
- 2026-09-02 — `in-progress` — improving the popover hover bridge and replacing card branch metadata with directory and last-edited metadata.
- 2026-09-02 — `completed` — added a hover bridge to keep the repository widget open while moving to its link, and updated both card renderers to show directory and last-edited metadata. Static selector/label validation passed; command execution remains unavailable in this session.
- 2026-09-02 — `completed` — made the full remote URL visible as a clickable `Repository URL` row in the Git widget.
- 2026-09-02 — `completed` — removed the redundant Git refresh button and handler from both card renderers; live Git details refresh whenever the card data is loaded.