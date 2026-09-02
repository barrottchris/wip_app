---
title: Replace embedded Postgres with app-local SQLite
status: in-progress
owner: copilot
last_updated: 2026-09-02
---

# Task

## Context
- WIP is a local-first, single-user Windows application; the MVP explicitly excludes a hosted backend or network-facing service.
- The architecture permits a local database file for WIP metadata, and the data model stores app records, settings, and JSON-shaped branch/component data.
- The current implementation starts an embedded Postgres server, downloads its binary when needed, stores its data in temporary directories, and uses PostgreSQL-specific SQL and driver dependencies.
- The relevant implementation surface is `internal/db/db.go`, `internal/app/store.go`, `internal/config/config.go`, `main.go`, `go.mod`, and the integration tests.

## Problem / goal
- Remove the dependency on an embedded Postgres binary and its external download path so WIP can build and run using only approved internal repositories and local machine resources.
- Persist WIP metadata in a SQLite database file stored with the WIP application data, surviving application restarts.
- Preserve the existing store/config APIs and current app-registry behavior while keeping the migration focused on the database layer.

## Questions to answer
- Should SQLite become the only supported backend for this application, or should a temporary PostgreSQL option remain behind configuration for migration/testing?
- Where should the database file live on Windows: beside the executable, or in the user's per-application data directory (recommended to avoid write-permission and upgrade problems)?
- Is an approved internal Go module proxy or vendored dependency available for the SQLite driver? The standard library does not provide a SQLite `database/sql` driver.
- Should existing embedded-Postgres data be migrated, or can this change start a fresh SQLite store for the current development build?

## Requirements
- Remove runtime startup, shutdown, temporary-directory, port, and binary-download behavior associated with embedded Postgres.
- Use a SQLite `database/sql` driver obtainable through the repository's approved internal dependency path.
- Store the database at a stable app-owned path and create its parent directory as needed.
- Keep the existing `db.Start`/`db.Stop` lifecycle usable by the application, or make the smallest corresponding call-site change.
- Convert schema and query SQL to SQLite-compatible syntax, including JSON defaults, timestamps, booleans, and parameter placeholders.
- Preserve persisted app/settings behavior, idempotent schema initialization, seed-on-empty behavior, and archived-app behavior.
- Add focused tests proving startup, schema creation, persistence across close/reopen, and the existing store/config operations.
- Update README and relevant comments so they no longer describe embedded Postgres or a hosted-Postgres connection path.

## Proposed implementation plan
1. Confirm backend policy, database location, dependency source, and migration expectations.
2. Update the module dependency and replace `internal/db` with a persistent SQLite connection and compatible schema initialization.
3. Adapt store/config SQL only where PostgreSQL syntax is incompatible, preserving public APIs and behavior.
4. Update startup messaging, shutdown comments, README, and integration tests.
5. Run focused Go tests, then the full test suite and the local smoke test.
6. Record validation results and mark this spec `completed`.

## Acceptance criteria
- A clean build does not require an embedded Postgres binary, a Postgres server, or a network download at runtime.
- Starting WIP creates or opens one stable SQLite file in the agreed app-data location.
- Data created in one process is available after stopping and starting the database again.
- Existing app registry, settings, archive, seed, and integration-test behavior remains intact.
- The repository documentation accurately describes SQLite local persistence and the approved dependency path.
- Focused tests, the full Go test suite, and the available smoke test pass, or any unrelated failures are documented.

## Status log
- 2026-09-02 — `todo` — spec created; awaiting confirmation of backend policy, file location, dependency source, and migration expectations.
- 2026-09-02 — `in-progress` — approved as SQLite-only with a per-user app-data file, fresh storage, and the repository's approved internal dependency path.
- 2026-09-02 — `in-progress` — implementation applied; command validation and Git branch creation remain unavailable in this session.