# WIP — Project Context

This file summarizes everything discussed so far about "WIP," an app idea, so a fresh conversation (on another account/device) can pick up with full context. Pair this with the four planning docs: `overview.md`, `mvp-scope.md`, `data-model.md`, `architecture.md`.

## The Problem
AI has made it fast to start new apps but hasn't made them faster to finish. Result: many half-done apps running concurrently, no organization, broken/messy versions, apps scattered across the C drive with no consistent git usage, and no easy way to remember how to start something back up.

## The Idea
"WIP" — a tool that manages the *portfolio* of in-progress apps, not the code itself. Think Jira's structure + GitHub's repo-awareness + a personal project launcher, built for high-concurrency AI-era solo development.

## Who It's For
- Now: personal tool, just for Chris.
- Later ambition: dev team tool, potentially company-wide, ideally simple enough for non-devs to use alongside experienced devs.

## Platform Decision
- **This will be a native Windows app** — not a browser/HTML-based UI. (An earlier HTML mockup was only a rough layout sketch, not representative of the real UI.)
- Chris's background: comfortable with HTML, JavaScript, Python, some Go. Wants to use Go in this project partly to learn it more.
- **Framework choice is still undecided** — flagged as a key open decision to revisit. Wails (Go backend + HTML/JS/CSS frontend, packaged as native Windows app) was proposed as a strong fit given the Go-learning goal and reuse potential for a future web version, versus Electron (Node-based, more common/tutorialized, no Go). Not yet decided — worth properly comparing options when picked back up.

## MVP Scope (see mvp-scope.md for full detail)
**Goal:** see all in-progress apps in one organized place, each properly under git, and start/stop any of them with one click.

In scope:
- App registry with rich metadata per app: name, tech stack, purpose, status, notes, repo/branch info
- Git integration — apps not yet in git get brought into it; branches are visible/tracked
- Start/Stop controls — each app can have multiple configurable components (e.g. frontend/backend), each with its own start and stop command (plain text commands leaning favorite, not fully decided); mix of Docker (where useful) and native process execution per component
- Runs as a local, single-user app for v1

## Explicitly Deferred (Post-MVP / v1.1+)
- **Brainstorming/idea page (v1.1):** a tree-based visual metaphor — idea starts as a seed, grows into a tree as requirements/functionality/user base get fleshed out, with a health/status indicator (e.g. showing dormancy). This was deliberately parked as v1.1, not MVP, but must not be forgotten.
- **Jira integration:** one-way (WIP → Jira), manually triggered sync of descriptions, epics, priority, status, comments — intended to help PMs. Auth via each user's own AD account (per-user, not shared service account).
- **Confluence integration:** README-to-Confluence-page mirroring, same one-way/manual model as Jira.
- **Company-wide/non-dev-friendly rollout** and **WIP evolving into a hosted web app** — later stretch goals.

## Data Model (see data-model.md for full detail)
Rough shape: `App Entry` (id, name, description, stack, status, notes, local_path, repo_url, branches, components, timestamps) → has many `Branch` (read-only, from git) and `Component` (name, start_command, stop_command, run_mode: docker/native).

## Architecture (see architecture.md for full detail)
Local app with a UI layer, a local backend/service layer (git operations, process start/stop, metadata read/write), local metadata storage, and a defined managed storage area on the C drive. No network service, hosted backend, or multi-user auth in MVP.

## Open Questions Still to Resolve
- Windows app framework: Wails vs Electron vs other (flagged as a key decision to come back to)
- Exact status taxonomy for app entries (active/paused/abandoned/etc. — first guess only)
- How multiple simultaneously-running apps are accessed (fixed ports vs. proxy/dashboard)
- Exact start/stop command storage format
- Whether Docker orchestration is WIP-generated (auto compose files) or calls existing per-app Dockerfiles/scripts
- Whether "managed storage area" means physically relocating app folders into one place, or just indexing/linking to wherever they currently live
- Whether brainstorming page ships with a full idea-to-project "graduation" flow or something lighter initially (this is v1.1, so lower priority)

## Working Style Note
Chris is deliberately trying to avoid scope creep/spiraling — that's the whole problem this project itself is meant to solve. Preference has been to fully flesh out planning docs before writing any code, and to explicitly park "nice to have" ideas in a clearly labeled later-phase list rather than let them creep into MVP scope.
