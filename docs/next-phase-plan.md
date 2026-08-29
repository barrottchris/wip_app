# WIP — Next Phase Plan

Covers the next four pieces of work, in the order decided, each scoped from
questions answered before any code was written. This sits alongside
`mvp-scope.md`, `data-model.md`, and `architecture.md` as a working plan —
not a replacement for them.

## File structure change (applies immediately, no feature attached)

Tests move out of being colocated with source into a dedicated `/tests`
directory, mirroring the `internal/` package layout:

```
wip-scaffold/
├── internal/
│   ├── app/{types,store,runtime}.go
│   ├── config/config.go
│   ├── db/db.go
│   ├── fsbrowse/browse.go
│   └── gitutil/gitutil.go
├── tests/
│   ├── app/{store_test,runtime_test}.go   (package app_test)
│   ├── fsbrowse/browse_test.go             (package fsbrowse_test)
│   └── gitutil/gitutil_test.go             (package gitutil_test)
└── frontend/src/{index.html,main.js,style.css}
```

All existing tests only use exported identifiers, so this works as
"black-box" test packages (`package app_test` importing `wip/internal/app`)
— no loss of test coverage, still runs with `go test ./...`.

## Working agreement (applies going forward)

From this point on, only new or changed files are provided per step — not
a full project rebuild/rezip each time. This is both to save tokens and so
the person building this actually integrates each change by hand and
learns the codebase as it grows.

---

## 1. Remove / Archive an app

**Decided:**
- **Soft archive**, not permanent delete — archived apps are hidden but
  recoverable, not gone.
- **Hidden by default** from the main registry; a separate, dedicated
  **Archived view** is where they live (not just greyed out inline).
- Archiving **optionally offers to delete the folder on disk too** — not
  just untracking it in WIP. This needs to be an explicit, separate
  confirmation from the archive action itself, since it's destructive and
  irreversible in a way archiving alone isn't.

**Shape this implies:**
- `Entry` gains an `Archived` (bool) or similar field — a soft-delete flag,
  separate from `Status` (lifecycle) entirely.
- `ListApps` needs to filter archived apps out by default, with a way to
  query archived-only for the new Archived view.
- A new nav item or sub-view for "Archived", listing archived apps with an
  "Unarchive" action.
- Archive action itself: confirm → set the flag. A **second, separate**
  prompt for "also delete the folder on disk?" — defaulting to no, since
  this is the one truly destructive path in the whole app so far.

## 2. Editing components (start/stop config) on an existing app

**Decided:**
- Docker vs. Native run mode is **selectable per component now**, even
  though Docker execution itself isn't built yet — the UI should support
  choosing it so it's not a second pass later.

**Left open, Claude's recommendation given (not yet confirmed):**
- **Location:** inside the existing app edit page, as a new "Components"
  section — consistent with how the Git section was added there rather
  than spun into a separate page.
- **Add/remove:** free add/remove of any number of components, not fixed
  slots — matches multi-component apps like the SRE agent (frontend +
  backend) without hardcoding a shape that won't fit everything.

**Shape this implies (pending confirmation of the above):**
- A repeatable component row: name, start command, stop command, run mode
  (Docker/Native) — add/remove buttons per row.
- Saving needs to replace the app's component list wholesale (simplest) or
  diff it (more complex) — worth deciding when this is actually built.

## 3. Real git status

**Decided:**
- **Scope: current branch + last commit date only** — not a full branch
  list. (Full branch history/staleness is still the later "branch health
  check" idea, parked post-MVP in `mvp-scope.md`.)
- **Read via shelling out to the `git` command**, not a Go git library —
  for consistency with `gitutil.Init`, which already assumes git is
  installed locally. No new dependency needed.
- **Refresh on-demand via a per-app refresh button** — not live on every
  registry load, to avoid slowing the registry down as more apps get
  tracked.

**Shape this implies:**
- `gitutil` gains something like `CurrentBranch(path)` and
  `LastCommitDate(path)`, each shelling out (`git rev-parse
  --abbrev-ref HEAD`, `git log -1 --format=%cI`).
- A new endpoint (e.g. `GET /api/apps/{id}/git-refresh`) the refresh button
  calls, updating the entry's `DefaultBranch` and a last-commit display
  value — separate from the existing `gitConnected` live check, which
  stays as-is (still checked live/cheaply, no refresh button needed there).

## 4. Real start/stop execution (deferred — needs more testing)

Not scoped yet — explicitly the last of this batch, and flagged as needing
more testing before diving in, likely because this is the piece that
actually touches real processes/Docker rather than metadata. Revisit once
1–3 are done and something real exists to test start/stop against.
