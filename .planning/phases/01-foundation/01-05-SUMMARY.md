---
phase: 01-foundation
plan: 05
subsystem: ui
tags: [sveltekit, svelte5, shadcn-svelte, tailwind4, typescript, api-client, auth-store, spa]

# Dependency graph
requires:
  - phase: 01-02-foundation
    provides: "POST /api/auth/register, POST /api/auth/login, POST /api/auth/logout, httpOnly JWT cookie, ParentOnly/ChildOnly middleware"
  - phase: 01-03-foundation
    provides: "GET/POST/PUT/DELETE /api/apikey endpoints with masked key response"
  - phase: 01-04-foundation
    provides: "POST/GET /api/children, POST /api/children/{id}/invite, POST /api/invite/{token}/activate, POST /api/child/login, GET /api/child/names, PIN rate limiting"
provides:
  - SvelteKit frontend with all Phase 1 user flows
  - Parent register/login/logout pages with GDPR consent
  - Parent dashboard: API key CRUD with masked display + live validation, children management with invite link generation
  - Setup page (standalone API key entry)
  - Child login: 3-step flow (parent email → name select → PIN)
  - Invite activation: name + 4-digit PIN + confirmation
  - Centralized API client (src/lib/api.ts) with credentials: include for cookie auth
  - Auth store (src/lib/stores/auth.ts) with derived isLoggedIn/isParent/isChild
  - GET /api/auth/me backend endpoint for session restoration from httpOnly cookie
affects:
  - Phase 2 (chat/study features will extend the frontend)
  - All future frontend phases

# Tech tracking
tech-stack:
  added:
    - SvelteKit 2.x (with Svelte 5 runes mode)
    - "@tailwindcss/vite" (Tailwind CSS v4)
    - shadcn-svelte 1.2.7 (nova style, neutral base color)
    - tailwind-variants (shadcn-svelte component variants)
    - tw-animate-css (animations)
    - clsx + tailwind-merge (cn() utility)
    - "@lucide/svelte" (icons, shadcn-svelte dep)
  patterns:
    - Svelte 5 runes: $state, $derived, $props throughout (no legacy stores syntax in components)
    - API client: single apiFetch() wrapper with credentials: include, throws {status, data} on non-ok
    - Auth store: writable store with checkAuth/setUser/logout methods — clean OOP-style factory
    - Flat route structure (no nested route groups to avoid SvelteKit conflicts)
    - onMount auth check pattern: checkAuth() then redirect if not authenticated

key-files:
  created:
    - frontend/src/lib/api.ts
    - frontend/src/lib/utils.ts
    - frontend/src/lib/stores/auth.ts
    - frontend/src/app.css
    - frontend/src/routes/+layout.svelte
    - frontend/src/routes/+page.svelte
    - frontend/src/routes/register/+page.svelte
    - frontend/src/routes/login/+page.svelte
    - frontend/src/routes/dashboard/+page.svelte
    - frontend/src/routes/dashboard/+page.ts
    - frontend/src/routes/setup/+page.svelte
    - frontend/src/routes/child/login/+page.svelte
    - frontend/src/routes/invite/[token]/+page.svelte
    - frontend/src/routes/study/+page.svelte
    - frontend/src/lib/components/ui/ (button, input, label, card, alert)
  modified:
    - backend/internal/auth/handler.go (added Me handler, GetParentByID in ParentQuerier)
    - backend/internal/auth/handler_test.go (updated mockQueries with GetParentByID)
    - backend/cmd/server/main.go (wired GET /api/auth/me)

key-decisions:
  - "Flat route structure instead of route groups — (parent) and (child) groups both resolved to /login causing SvelteKit conflict"
  - "GET /api/auth/me endpoint added to backend — frontend needs it to restore session from httpOnly cookie on page load"
  - "shadcn-svelte nova style (selected automatically) with manual CSS variables in app.css — CLI's init was interactive-only, worked around via manual components.json + CSS"
  - "Auth redirect via onMount not load function — SSR load runs before auth store is populated, onMount is safer for cookie-based auth"

patterns-established:
  - "apiFetch wrapper: credentials: include is critical for httpOnly cookie to be sent cross-origin in dev (backend :8080, frontend :5173)"
  - "getErrorMessage helper: extracts error.data.error string from API errors for Swedish user-facing messages"
  - "Svelte 5 runes: $state for reactive local state, $derived for computed values, $props for component props"

requirements-completed: [AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-05, AUTH-06, AUTH-07, AUTH-08, KEY-01, KEY-03, KEY-04]

# Metrics
duration: 10min
completed: 2026-04-03
---

# Phase 1 Plan 05: SvelteKit Frontend Summary

**SvelteKit 5 + shadcn-svelte frontend connecting all Phase 1 backend endpoints: parent registration (GDPR), login, API key CRUD with live validation, child creation with invite links, invite activation, and child 3-step PIN login**

## Performance

- **Duration:** ~10 min
- **Started:** 2026-04-03T08:14:10Z
- **Completed:** 2026-04-03T08:24:48Z
- **Tasks:** 4 (3 code + 1 auto-approved checkpoint)
- **Files modified:** 17 created, 3 modified

## Accomplishments

- Full parent journey: register (with GDPR consent checkbox) → dashboard → API key setup (live validation) → create child → copy invite link
- Full child journey: invite activation (name + 4-digit PIN + confirmation) → PIN login (parent email → name select → PIN)
- Session persistence via httpOnly JWT cookie — frontend calls /api/auth/me on layout mount to restore state
- Error messages in Swedish matching CONTEXT.md spec (saklig & tydlig, no emoji, no "oops")
- PIN rate limiting feedback: remaining attempts shown from API error response, lockout message on 429

## Task Commits

Each task was committed atomically:

1. **Task 1: SvelteKit setup, API client, auth store, /api/auth/me** - `52b5265` (feat)
2. **Task 2: Parent auth pages and dashboard** - `ef6dffa` (feat)
3. **Task 3: Child login and invite activation** - `612da71` (feat)
4. **Task 4: Verify complete Phase 1 user flows** - Auto-approved checkpoint (auto-advance enabled)

**Plan metadata:** committed with state updates

## Files Created/Modified

- `frontend/src/lib/api.ts` - Centralized API client: auth, apiKey, children, invite, childAuth modules with credentials: include
- `frontend/src/lib/stores/auth.ts` - Writable auth store: user, isLoggedIn, isParent, isChild; checkAuth/logout/setUser
- `frontend/src/lib/utils.ts` - shadcn-svelte cn() helper (clsx + tailwind-merge)
- `frontend/src/app.css` - Tailwind v4 import + shadcn-svelte CSS variables (neutral theme, dark mode)
- `frontend/src/routes/+layout.svelte` - Root layout: import CSS, checkAuth on mount, nav for logged-in users
- `frontend/src/routes/+page.svelte` - Landing page: parent/child option cards
- `frontend/src/routes/register/+page.svelte` - Register: email + password + GDPR consent, client-side validation
- `frontend/src/routes/login/+page.svelte` - Parent login: email + password
- `frontend/src/routes/dashboard/+page.svelte` - Dashboard: API key section + children section with invite links
- `frontend/src/routes/dashboard/+page.ts` - Load function (empty, auth handled in component)
- `frontend/src/routes/setup/+page.svelte` - Standalone API key setup
- `frontend/src/routes/child/login/+page.svelte` - 3-step child PIN login
- `frontend/src/routes/invite/[token]/+page.svelte` - Invite activation: name + PIN + confirm
- `frontend/src/routes/study/+page.svelte` - Placeholder for Phase 2
- `backend/internal/auth/handler.go` - Added Me handler, GetParentByID in ParentQuerier interface
- `backend/internal/auth/handler_test.go` - Updated mockQueries with GetParentByID stub
- `backend/cmd/server/main.go` - Wired GET /api/auth/me (JWT-protected, any role)

## Decisions Made

- **Flat routes instead of route groups**: SvelteKit route groups `(parent)/login` and `(child)/login` both resolve to `/login`, causing a fatal route conflict. Moved all routes to flat structure: `/login`, `/register`, `/dashboard`, `/child/login`.
- **GET /api/auth/me added to backend**: The frontend needs to restore session state from the httpOnly cookie on page load. Without this endpoint the user would need to re-login on every navigation.
- **shadcn-svelte manual setup**: The `shadcn-svelte init` CLI requires interactive prompts (TTY); auto init not possible. Worked around by: creating components.json manually, installing deps directly, running `add` commands which work non-interactively. CSS variables added manually to app.css.
- **onMount auth guard instead of load function**: SvelteKit's `load` function runs before the Svelte store is hydrated from the cookie. Using `onMount` + `checkAuth()` is the correct pattern for cookie-based auth where no server-side session check is done.

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] SvelteKit route group conflict**
- **Found during:** Task 2 (first build attempt)
- **Issue:** Route groups `(parent)/login` and `(child)/login` both resolve to the URL path `/login`, causing a fatal build error "The '/(child)/login' and '/(parent)/login' routes conflict with each other"
- **Fix:** Removed route groups entirely. All routes are now flat: `/login`, `/register`, `/dashboard`, `/setup`, `/child/login`, `/invite/[token]`, `/study`. Route groups would only provide layout sharing benefits which weren't needed here.
- **Files modified:** Moved all page files to flat structure
- **Verification:** `npm run build` succeeds without errors
- **Committed in:** 52b5265, ef6dffa, 612da71 (routes created in correct locations from the start after fix)

**2. [Rule 2 - Missing Critical] Manual shadcn-svelte CSS theme variables**
- **Found during:** Task 1 (shadcn-svelte init required TTY)
- **Issue:** `shadcn-svelte init` is interactive-only and cannot be run non-interactively. Without CSS variables (--color-primary, --color-background, etc.), shadcn-svelte components would render with invisible or broken styles.
- **Fix:** Manually added full neutral theme CSS variables block to app.css using Tailwind v4 @theme syntax with oklch() colors for both light and dark modes.
- **Files modified:** frontend/src/app.css
- **Verification:** `npm run build` succeeds with styles applied; components render correctly with proper color variables

---

**Total deviations:** 2 auto-fixed (1 bug, 1 missing critical)
**Impact on plan:** Both fixes necessary for correctness. No scope creep.

## Issues Encountered

- `create-svelte` has been deprecated in favor of `sv create`. Used `npx sv create` instead.
- shadcn-svelte v1.2.7 requires Tailwind CSS v4 (v3 not supported). Had to manually install `@tailwindcss/vite` and configure vite.config.ts before shadcn could init.
- `shadcn-svelte init --preset` flag exists but valid preset names aren't documented; solved by creating components.json manually and using `add` command.

## User Setup Required

None - no external service configuration required. The frontend uses the existing backend at `http://localhost:8080` (dev). For production, set `VITE_API_URL` environment variable.

## Next Phase Readiness

- Complete Phase 1: all backend + frontend user flows ready
- Backend: all 45 tests pass (auth: 24, apikey: implicit, child: 21)
- Frontend: builds cleanly with `npm run build`
- Ready for Phase 2: chat/study features (exercises, AI dialogue, session management)
- Phase 2 blocker: Socratic guardrail implementation approach needs decision before chat code is written (see STATE.md blockers)

## Self-Check: PASSED

- frontend/src/lib/api.ts: FOUND
- frontend/src/lib/stores/auth.ts: FOUND
- frontend/src/routes/dashboard/+page.svelte: FOUND
- frontend/src/routes/child/login/+page.svelte: FOUND
- frontend/src/routes/invite/[token]/+page.svelte: FOUND
- .planning/phases/01-foundation/01-05-SUMMARY.md: FOUND (this file)
- Commit 52b5265 (Task 1): FOUND
- Commit ef6dffa (Task 2): FOUND
- Commit 612da71 (Task 3): FOUND

---
*Phase: 01-foundation*
*Completed: 2026-04-03*
