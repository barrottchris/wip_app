---
title: Home page and brand navigation
status: completed
owner: copilot
last_updated: 2026-09-02
---

# Task

## Context
- WIP is a local-first Windows app for managing work-in-progress apps, as described in [docs/overview.md](../overview.md), [docs/architecture.md](../architecture.md), and [docs/mvp-scope.md](../mvp-scope.md).
- The product problem is that AI makes it easy to start apps, while unfinished apps become scattered, hard to resume, and difficult to manage.
- The current frontend opens directly on the Registry route. The `WIP` brand in [frontend/src/index.html](../../frontend/src/index.html) is currently static and does not navigate.
- The current MVP registry remains the operational workspace; this task adds orientation without changing the registry's role.

## Problem / goal
- A first-time or returning user needs a concise explanation of why WIP exists before entering the operational registry.
- The `WIP` brand should provide a reliable way back to that explanation from any primary navigation page.
- The home page may use a visual explanation of the problem, but the product should remain useful without generated artwork.

## Questions to answer
- Resolved: open on Home every time in this early build.
- Resolved: use a CSS/layout-based visual now and defer any generated image; reserve an image slot only if it fits the page.
- Resolved: include clear Registry and Add app shortcuts without duplicating registry content.

## Requirements
- Add a Home route and page that briefly explains the problem WIP solves using language grounded in the product docs.
- Make the `WIP` brand in the top banner navigate to Home and expose an appropriate interactive affordance.
- Keep Registry, Archived, Brainstorm, Activity, Settings, app detail, and add-app navigation working.
- Make the initial route decision explicit and easy to change if the user prefers Registry-first behavior.
- Include a visual problem explanation only if it can be delivered without unapproved image-generation cost; otherwise use intentional typography, layout, and existing UI primitives.
- Keep the page consistent with the existing native-app frontend and usable at narrow window widths.
- Add focused frontend coverage for the Home route, brand navigation, and initial route behavior where supported by the repository's existing test setup.

## Proposed implementation plan
1. Confirm the initial route, Home content emphasis, and whether a visual is wanted after the image-cost decision.
2. Add a Home page module and register its route.
3. Make the top-banner brand navigate to Home and update active navigation state appropriately.
4. Add the approved visual treatment, using a non-generated fallback if image generation is not approved.
5. Add focused regression coverage and validate the frontend behavior.
6. Update this spec with the final decision, validation, and completion status.

## Acceptance criteria
- Clicking `WIP` always navigates to Home from a primary page.
- Home clearly communicates the problem WIP solves and provides a direct path to the Registry.
- The chosen initial route matches the confirmed decision.
- No generated image is created or added without explicit approval after its cost implications are understood.
- Existing routes and the Add app action continue to work.
- The page remains legible and usable on desktop and narrow app windows.
- Focused frontend validation passes.

## Status log
- 2026-09-02 — `todo` — spec drafted from the requested Home page and current static brand/default Registry navigation.
- 2026-09-02 — `in-progress` — user confirmed Home as the initial route, approved Registry/Add app shortcuts, and requested a CSS visual with image generation deferred.
- 2026-09-02 — `completed` — added Home as the initial route, made the WIP brand navigate Home, added Registry/Add app shortcuts, and added a responsive CSS visual with a future image-slot hook. Source-level validation completed; executable tests were unavailable because this session exposes no command runner.