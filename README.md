# WIP — Scaffold (local web server version)

This scaffold was switched from Wails (native window) to a **plain Go HTTP
server** serving the frontend over `localhost` in a normal browser tab. This
was a deliberate pivot: Wails needs WebView2/native packaging that may not be
installable on a locked-down work machine, whereas a plain Go binary using
only the standard library needs nothing beyond Go itself.

Everything still runs **locally, on your machine** — that part hasn't
changed. Local git access, starting/stopping local processes, and reading
your local filesystem all still work exactly as designed, because the Go
server is still running where your code lives. Only the way you *view* it
changed (browser tab instead of a native window).

## What's here

- `main.go` — starts an HTTP server on `localhost:34115` (change the port
  freely), serves the frontend as static files, and exposes a small JSON API:
  - `GET /api/apps` — list all tracked apps
  - `GET /api/apps/{id}` — get one app's detail
  - `GET /api/apps/{id}/git` — get branch info for an app
  - `POST /api/apps/{id}/start` — start a component (body: `{"component": "name"}`)
  - `POST /api/apps/{id}/stop` — stop a component (same body shape)
- `internal/app/types.go` — the `Entry`, `Component`, `Branch`, `Status`
  types matching `data-model.md`.
- `internal/app/store.go` — a **placeholder** in-memory store, seeded with
  one sample app. Real persistence (SQLite vs JSON files — still an open
  question) replaces this.
- `frontend/src/` — plain HTML/JS/CSS registry view, calling the API above
  with `fetch()`. No framework or build step.

## Running it

No Wails CLI, no Node.js needed for this version — just Go.

```
go mod tidy
go run .
```

Then open **http://localhost:34115** in your browser.

To build a standalone binary instead of running via `go run`:
```
go build -o wip.exe .
```
Then just run `wip.exe` and open the same URL in your browser.

## What's stubbed / TODO

- `start`/`stop` API calls just print to console right now — real process
  execution (native) and Docker invocation still need building.
- Git status reads from the placeholder store, not real git — actual git
  integration (e.g. via `go-git` or shelling out) is still to do.
- Storage is in-memory only — nothing persists between runs yet.
- No onboarding flow (adding a new app to the registry) yet — the one sample
  app is hardcoded in `store.go`.
- The manual path-parsing router in `main.go` is intentionally simple for
  MVP — worth swapping for a real router if routes grow.

## Next steps

See `mvp-scope.md`, `data-model.md`, and `architecture.md` for the full plan.
Suggested build order:
1. Real persistence (decide SQLite vs JSON, replace the in-memory store)
2. Onboarding flow (add an existing folder as a tracked app)
3. Real git status integration
4. Real start/stop process execution

## Note on future hosting (Kubernetes, team-wide use)

The local git/start-stop functionality needs to stay local, per-machine —
that's inherent to what it does (only your machine has your local repos and
processes). If a hosted, multi-user piece is built later (e.g. the Jira/
Confluence integration, or a team-wide idea tracker), that could reasonably
live on Kubernetes as a separate service — but it would be a different
component from this local server, not a replacement for it.

## App shell (added after initial scaffold)

The frontend now has a proper shell instead of a single flat page:
- **Top banner** — branding, a live running-components count, "Add app" button.
- **Left nav** — Registry (default), Brainstorm (placeholder, v1.1), Activity
  (placeholder), Settings.
- **Settings page** — calls `GET /api/settings` / `POST /api/settings`
  (`internal/config/config.go`). Currently covers the managed apps folder
  path and a GitHub username/token field. The token is never echoed back —
  only whether one is set. Real secret storage (OS credential store or
  encrypted file) still needs building; right now it's an in-memory stub.

Client-side routing between nav pages is a small hand-rolled router in
`main.js` (`navigateTo()`), no framework — consistent with the rest of this
scaffold's "no build step" approach.

## Persistence (app-local SQLite) — added after app shell

WIP stores its metadata in a persistent SQLite database at
`%LOCALAPPDATA%\WIP\wip.db` on Windows. SQLite runs in-process, so WIP does
not require a Postgres installation, server, port, or runtime binary
download. The database is created and migrated automatically on startup.

- `internal/db/db.go` — opens the app-local SQLite file and applies the
  idempotent schema.
- `internal/app/store.go` — reads and writes the `apps` table through
  `database/sql`; list fields are stored as JSON text.
- `internal/config/config.go` — persists settings in the SQLite `settings`
  table.
- `main.go` — opens the database before the HTTP server, seeds one sample
  app on first run only (`SeedIfEmpty`), and closes the connection on exit.

The SQLite driver is `modernc.org/sqlite` and must be resolved through the
repository's approved internal Go module source or vendored before an
offline build.

## Onboarding ("Add app") — added after persistence

The "Add app" button now works. It supports both onboarding flows discussed:

- **Existing folder** — browse to and select a folder that already has code
  in it.
- **Create new** — browse to a parent location and create a fresh empty
  folder there.

Since a browser can't return a real, absolute OS path from a native file
picker (a deliberate browser security restriction), the folder picker is
**server-side**: `internal/fsbrowse/browse.go` lists directories via the Go
backend (which has full local disk access), and the frontend renders
whatever comes back as a clickable list with up/into navigation. This only
works because WIP's backend runs locally — it wouldn't work if this were a
truly remote/hosted service.

Git handling matches what was decided: WIP **never initializes git
silently**. When you select an existing folder with no `.git`, a prompt
appears offering to run `git init`, with an explicit "skip for now" option.

New pieces:
- `internal/fsbrowse/browse.go` — directory listing, with a real passing
  unit test (`browse_smoke_test.go`) confirming hidden folders are excluded
  and parent/current paths are correct.
- `internal/gitutil/gitutil.go` — `HasGit` / `Init`, wrapping the local
  `git` command. Also has a real passing test (`gitutil_smoke_test.go`)
  confirming detection and init both work.
- `internal/app/store.go` — added `Slugify` (name → URL-safe ID) and
  `EnsureUniqueID` (handles name collisions by appending `-2`, `-3`, etc.),
  each with their own test.
- New API endpoints: `GET /api/browse`, `GET /api/git-status`,
  `POST /api/git-init`, and `POST /api/apps` (create).
- `frontend/src/main.js` — new `renderAddAppPage`, wired to the nav's "Add
  app" button, with the existing/new toggle and folder browser.

**Verified in the repository:** the existing filesystem, git, and slug logic
has unit-test coverage. Run `go mod tidy`, `go build`, `go vet`, and
`go test ./...` after the approved internal module source is configured.

## Still not built

- Real git status for the *registry view* (branches/last-commit still come
  from placeholder data — onboarding's git-init is new, but reading real
  branch info back into the app entry isn't wired up yet)
- Real start/stop process execution (still prints to console)
- Editing an app after creation (stack, notes, components all need a way
  to be added post-onboarding, since the creation form is deliberately
  minimal — see mvp-scope.md)
- GitHub token real secret storage
- Branch health check, launch-in-VS-Code/Docker/file actions (flagged
  post-MVP in mvp-scope.md)

## Connections, editing, and real running state — added after onboarding

Several fixes/additions based on using the onboarding flow for real:

- **Folder picker now defaults to a drive root** (`C:\` on Windows, `/`
  elsewhere) instead of the current user's home folder — one click closer
  to anywhere useful, and not biased toward any particular username.
- **"Running" vs. lifecycle status are now genuinely separate.** `Entry.
  Status` (active/paused/abandoned/shipped) is a label *you* set — it does
  not mean "currently running," and editing it never did and never will
  change based on process state. Whether something is actually running is
  computed live from component state (`internal/app/runtime.go`) and shown
  as a distinct "Running" badge that overrides the lifecycle badge on the
  registry card when true. This wasn't a real distinction before — every
  app just always showed "Active" regardless of anything; now start/stop
  actions genuinely flip a tracked state, and the registry reflects it.
  (This state is intentionally *not* persisted — it resets to "not
  running" on restart, same as real processes would.)
- **Connection pills** on each app card (and the detail page): Git,
  Jira, Confluence. Git is checked live against the actual folder each
  time (not cached, so it can't go stale if git gets initialized outside
  WIP). Jira and Confluence are always shown greyed out / "coming soon" —
  they're not built yet, so the UI doesn't pretend otherwise.
- **App detail/edit page** — click an app's name (or a Git/pill) to open
  it. Editable: name, description, status, and folder path (via the same
  server-side folder browser as onboarding, kept in page-local state so it
  never conflicts with the add-app form's state).

New/changed pieces:
- `internal/app/runtime.go` — `RuntimeTracker`, with tests confirming
  running state is tracked correctly and apps don't leak state into each
  other.
- `internal/app/store.go` — added `UpdateApp` for the edit form.
- `main.go` — `EntryWithConnections` wraps entries with live-computed
  `gitConnected` plus the Jira/Confluence coming-soon flags; `PUT
  /api/apps/{id}` added for edits; `start`/`stop` now call
  `runtime.SetRunning` for real instead of only printing.
- `internal/fsbrowse/browse.go` — default path fixed, with a test.

**Verified in this environment:** full `go build`, `go vet`, and `go test
./...` all pass. Frontend JS syntax-checked with Node. The actual browser
click-through against a live server still needs your machine, same caveat
as before.

## Still not built

- Real git status for the registry view's branch/last-commit display
  (still placeholder data — separate from the new live git-*connected*
  check, which only answers yes/no, not branch details)
- Real start/stop process execution (state now tracks correctly, but
  nothing real is actually launched yet)
- Editing stack, notes, or components (the edit page is deliberately
  narrow — name/description/status/path only, per the original scope
  decision to keep forms minimal)
- GitHub token real secret storage
- Remove/archive action for apps
- Branch health check, launch-in-VS-Code/Docker/file actions, Jira/
  Confluence integrations themselves (all flagged post-MVP in
  mvp-scope.md)
