# WIP — Data Model (MVP)

Rough shape only — field names/types are a starting point to react to, not final.

## App Entry

The core object. One per project WIP is tracking.

| Field | Type | Notes |
|---|---|---|
| `id` | string | Internal unique identifier |
| `name` | string | Display name |
| `description` | text | Purpose — what the app is/does |
| `stack` | list of strings | e.g. `["Node.js", "React", "Postgres"]` |
| `status` | enum | Exact set TBD — starting point: `active`, `paused`, `abandoned`, `shipped` |
| `notes` | text (freeform) | Running notes — why paused, next steps, anything unstructured |
| `local_path` | string | Location within the managed WIP storage area on the C drive |
| `repo_url` | string | GitHub repo link |
| `default_branch` | string | e.g. `main` |
| `branches` | list of `Branch` | See below |
| `components` | list of `Component` | See below — the start/stop units |
| `created_at` | datetime | When added to WIP |
| `last_touched_at` | datetime | Derived from git activity — last commit date. Basis for future dormancy/health signals (v1.1) |

## Branch

Tracks the branches that exist within an app's repo.

| Field | Type | Notes |
|---|---|---|
| `name` | string | Branch name |
| `last_commit_at` | datetime | For surfacing which branches are stale |
| `is_default` | boolean | Marks the main/primary branch |

*(MVP: likely read-only, pulled from git — no branch management actions planned for v1.)*

## Component

A named, runnable piece of the app (e.g. "Frontend", "Backend", "Worker"). This is what the Start/Stop buttons act on.

| Field | Type | Notes |
|---|---|---|
| `name` | string | e.g. "Frontend", "Backend", "Docker stack" |
| `start_command` | string (plain text) | The shell command to run |
| `stop_command` | string (plain text) | The shell command to stop it |
| `run_mode` | enum | `docker` or `native` — set per component, not uniform across the app |
| `status` | enum (runtime, not stored) | `running` / `stopped` — reflects current state, computed at runtime rather than persisted |

An app can have one component (simple case) or several (e.g. separate frontend/backend), each started/stopped independently.

## Notes on Scope

- No user/auth model needed for MVP — single local user.
- No fields yet for Jira/Confluence linkage, or brainstorming/idea-tree data — those belong to their respective post-MVP phases and will get their own model additions when designed.
- `status` enum values are a first guess — worth deciding properly once the app registry UI is sketched, since the list/board view will likely be organized by this field.
