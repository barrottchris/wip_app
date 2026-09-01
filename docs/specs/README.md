# Specs workflow

This repository uses spec files as the planning layer before implementation.

## Purpose

The goal is to keep work deliberate and aligned with the product direction in the docs folder. Each task starts as a spec, then moves through a status lifecycle before code is implemented.

## Status lifecycle

- `todo` — the task is defined and waiting for clarifying questions or approval
- `in-progress` — work is currently being executed
- `completed` — the work is implemented, validated, and the plan is complete

## File convention

Store each task spec in `docs/specs/` using a filename like:

- `feature-name.spec.md`
- `bug-fix.spec.md`
- `refactor.spec.md`

## Required fields

Each spec should include:

- title
- status
- owner
- last_updated
- context summary from the docs folder
- questions to answer
- requirements
- implementation plan
- acceptance criteria
- status log

## Workflow

1. Read the docs folder and capture the relevant context.
2. Draft or update the spec.
3. Ask the required clarifying questions.
4. Wait for the user to confirm the plan before implementing code.
5. Update the spec status to `in-progress` when implementation begins.
6. Update the spec to `completed` after validation and closure.
