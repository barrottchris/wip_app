---
title: Main Go entrypoint refactor
status: in-progress
owner: copilot
last_updated: 2026-09-02
---

# Task

## Context
- WIP is a local-first Go application with a static frontend and a small HTTP API.
- The architecture separates storage, runtime tracking, process management, filesystem browsing, and git operations into internal packages.
- `main.go` currently combines process startup, dependency wiring, route registration, HTTP handlers, response projection, activity recording, and shutdown handling.

## Problem / goal
- Make `main.go` a small, readable composition root that starts the application and wires dependencies.
- Move HTTP API behavior into focused files/packages so handlers are easier to navigate and test.
- Preserve the current API paths, JSON contracts, Windows behavior, local SQLite lifecycle, and runtime semantics.

## Questions to answer
- Confirmed 2026-09-02: behavior-preserving structural refactor only; no endpoint or UI changes.
- Confirmed 2026-09-02: introducing `internal/httpapi` is acceptable if it keeps the entrypoint manageable.

## Requirements
- `main.go` should contain only application bootstrap, dependency construction, server startup, and top-level shutdown coordination.
- Existing API routes and response shapes must remain unchanged.
- App lifecycle status and live component runtime status must remain separate.
- Git enrichment, runtime/log enrichment, and response DTOs should have one clear owner.
- Handler groups should be independently testable with `net/http/httptest` where practical.
- Do not add a third-party router unless the standard Go 1.22 `http.ServeMux` cannot express the current routes cleanly.
- Preserve the current local-only and Windows-specific behavior.

## Proposed implementation plan
1. Add focused characterization tests around the current API behavior before moving code: list/get apps, app subroutes, start/stop/terminal, git refresh, onboarding, archive, components, settings, and activity.
2. Extract the computed API view model and enrichment logic from `main.go` into an API/view package, including `EntryWithConnections`, `GitDetails`, and runtime/git/log projection.
3. Extract handlers into focused files or an `internal/httpapi` package:
   - app registry and app subroutes
   - onboarding and filesystem/git helpers
   - settings and activity
   - shared JSON/error helpers
4. Give the API server one constructor and one route-registration method, so dependency wiring is explicit and route setup is no longer interleaved with bootstrap.
5. Reduce `main.go` to database startup, store/config/runtime/process-manager construction, API server construction, static file serving, graceful shutdown, and `ListenAndServe`.
6. Run formatting, the full Go test suite, and focused HTTP handler tests; compare route and JSON behavior against the characterization tests.

## Acceptance criteria
- `main.go` is a concise bootstrap/composition root with no business-specific HTTP handler bodies.
- API behavior and frontend-visible JSON remain compatible.
- Handler tests cover the highest-risk paths, especially start/stop/terminal and computed runtime/git fields.
- `go test ./...` passes.
- The application still starts locally and shuts down the database cleanly on Windows.

## Status log
- `todo` — 2026-09-02 — plan drafted; awaiting confirmation of scope and package boundary.
- `in-progress` — 2026-09-02 — scope and package boundary approved; implementation started.
- `in-progress` — 2026-09-02 — extracted API routing and handlers into `internal/httpapi`; validation remains pending because command execution was unavailable in this session.