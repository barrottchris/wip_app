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
