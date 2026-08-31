# WIP — Overview

## The Problem

AI has collapsed the cost of *starting* a new app — but not the cost of *finishing* one. The result:

- Apps get started fast, then abandoned mid-way, often in a broken or half-working state.
- Multiple apps end up running/being worked on at once, with no easy way to spin one down cleanly to focus on another.
- Projects live in loose, unorganized folders on the C drive — no consistent structure, no reliable version control.
- Each app can end up with multiple branches or copies ("app-v2-final") with no clear source of truth.
- Starting an app back up requires digging through shell history or scattered notes to remember the right commands.

The core problem isn't lack of ideas or lack of ability to build — it's **lack of a system to manage work-in-progress** so things don't quietly rot.

## The Vision

**WIP** is a tool that manages the *portfolio* of in-progress apps and ideas, not the code itself. It sits above Git, above Docker, above the individual project — giving a single place to see, launch, and maintain everything currently being worked on.

Think: Jira's structure + GitHub's repo-awareness + a personal project launcher, purpose-built for the reality of AI-accelerated, high-concurrency solo development.

## Who It's For

- **Primary user (now):** Chris, personally — managing his own sprawl of side projects.
- **Later:** a dev-team tool, potentially rolled out company-wide. Ambition is for it to be simple enough that non-developers could use it alongside experienced developers.

## Core Principles

1. **Git-first.** Every app should live under proper version control. Apps not currently in git get brought into it.
2. **One managed home.** Replace scattered C-drive folders with a single organized, findable structure.
3. **Run things independently.** Apps can be started and stopped individually — a long-running task in one app shouldn't block work on another.
4. **Low friction to resume.** No more hunting for start commands — one click starts a project's components.
5. **Don't let this tool spiral.** WIP itself is deliberately scoped in phases (MVP → later phases) to avoid becoming exactly the kind of half-finished, over-scoped project it's designed to prevent.

## Phasing Philosophy

- **MVP** — solves the core, daily pain: organized storage, git tracking, and one-click start/stop for apps already in progress.
- **v1.1 and beyond** — richer features (idea brainstorming, team integrations, non-dev-friendly rollout) layered on only once the core loop is proven and used.

See `mvp-scope.md` for the full scope cut line and the parked-for-later list.
