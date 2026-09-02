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
- The current compact visual size is appropriate, but the surrounding Home layout no longer feels aligned because its fixed-width visual, text column, and principles strip use different visual edges.

## Questions to answer
- Resolved: open on Home every time in this early build.
- Resolved: use a CSS/layout-based visual now and defer any generated image; reserve an image slot only if it fits the page.
- Resolved: include clear Registry and Add app shortcuts without duplicating registry content.
- Resolved: use a centered content frame with a balanced two-column hero and a principles strip aligned to the same frame edges.

## Requirements
- Add a Home route and page that briefly explains the problem WIP solves using language grounded in the product docs.
- Make the `WIP` brand in the top banner navigate to Home and expose an appropriate interactive affordance.
- Keep Registry, Archived, Brainstorm, Activity, Settings, app detail, and add-app navigation working.
- Make the initial route decision explicit and easy to change if the user prefers Registry-first behavior.
- Include a visual problem explanation only if it can be delivered without unapproved image-generation cost; otherwise use intentional typography, layout, and existing UI primitives.
- Keep the page consistent with the existing native-app frontend and usable at narrow window widths.
- Add focused frontend coverage for the Home route, brand navigation, and initial route behavior where supported by the repository's existing test setup.
- Preserve the current visual component sizes while reorganizing the surrounding layout for clearer alignment.
- Keep the visual container compact and align its outer edges with the hero content frame.
- Ensure the principles strip, hero text, and visual share a consistent responsive width and spacing system.

## Proposed implementation plan
1. Confirm the initial route, Home content emphasis, and whether a visual is wanted after the image-cost decision.
2. Add a Home page module and register its route.
3. Make the top-banner brand navigate to Home and update active navigation state appropriately.
4. Add the approved visual treatment, using a non-generated fallback if image generation is not approved.
5. Add focused regression coverage and validate the frontend behavior.
6. Update this spec with the final decision, validation, and completion status.

### Reorganization plan
1. Wrap the Home content in a centered frame with a stable maximum width and consistent horizontal gutters.
2. Keep the introduction and compact visual as the two columns of one hero row, vertically centered and aligned.
3. Keep the visual's internal boards, hub, and text dimensions unchanged; adjust only the container placement and responsive flow.
4. Align the three principles to the same frame and use equal columns on wide windows, stacking cleanly on narrow windows.
5. Validate the page at wide and narrow layouts and update this spec with the result.

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
- 2026-09-02 — `completed` — compacted the CSS visual container while preserving the internal illustration element sizes and responsive page proportions.
- 2026-09-02 — `todo` — user requested a layout reorganization because the compact visual now feels misaligned with the rest of the Home page; plan drafted for confirmation before editing.
- 2026-09-02 — `in-progress` — user approved the reorganization plan; implementation started.
- 2026-09-02 — `completed` — reorganized Home into a centered frame with a shared hero row, aligned principles strip, and responsive narrow-window flow. Source-level validation completed; executable tests were unavailable because no command runner was exposed in this session.
- 2026-09-02 — `completed` — replaced the vague Home headline with concrete language about WIP's findable and runnable project registry.
- 2026-09-02 — `completed` — updated the Home headline to "Your work-in-progress, managed in one place."
- 2026-09-02 — `completed` — added a restrained Planned connections note for GitHub, Jira, and Confluence without implying those integrations are live.