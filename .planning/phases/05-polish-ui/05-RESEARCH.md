# Phase 5: Polish + UI — Research

**Researched:** 2026-04-03
**Status:** Complete

## Phase Boundary

Phase 5 delivers three things: mobile-responsive design across all views, a calm visual identity, and a landing page with service info + registration/login entry points.

Requirements: UI-01 (mobile-first responsive), UI-02 (clean calm UI), UI-03 (landing page).

## Current State Analysis

### Existing Frontend Stack
- **SvelteKit** with Svelte 5 (runes: `$state`, `$props`)
- **Tailwind CSS v4** via `@tailwindcss/vite` plugin
- **shadcn-svelte** (bits-ui v2) — Button, Card, Input, Label, Alert components
- **tw-animate-css** for animations
- **@lucide/svelte** for icons
- No custom fonts — system font stack (`ui-sans-serif, system-ui, sans-serif`)

### Existing Color System (app.css)
Currently using shadcn defaults — all achromatic (oklch with 0 chroma). No brand colors applied yet. Both light and dark variants exist via `.dark` class with `@custom-variant dark`.

### Existing Pages (10 routes)
| Route | Purpose | Mobile Concern |
|-------|---------|----------------|
| `/` (+page.svelte) | Landing — two-card grid (Parent/Child) | Minimal; `sm:grid-cols-2` already |
| `/login` | Parent login form | Centered card, works on mobile |
| `/register` | Parent registration | Same pattern as login |
| `/child/login` | Child PIN login | Same pattern |
| `/invite/[token]` | Invite activation | Same pattern |
| `/dashboard` | Parent dashboard (API key + children) | `max-w-4xl`, decent but untested |
| `/setup` | Setup flow | Unknown |
| `/study` | Subject list + due reviews | Grid with responsive cols |
| `/study/[subject]` | Topic list | Needs check |
| `/study/[subject]/[topic]` | Exercise list | Needs check |
| `/chat/[sessionId]` | Chat interface | Most complex for mobile |

### Navigation (layout.svelte)
Simple top nav bar when logged in: logo left, Dashboard link + Logga ut right. No hamburger menu. No navigation when logged out (landing page has inline links).

## Technical Findings

### 1. Mobile Responsiveness (UI-01)
**Confidence: HIGH**

**What needs fixing:**
- Navigation needs hamburger menu for mobile (CONTEXT decision)
- Touch targets: buttons/links should be min 44px tap area
- Safe area: `env(safe-area-inset-*)` for notched devices (iPhone X+)
- Viewport meta tag: verify `<meta name="viewport" content="width=device-width, initial-scale=1">` exists
- Chat page is the most complex mobile layout (message bubbles + input bar at bottom)
- No horizontal scroll on any page at 320px viewport width

**Approach:**
- Tailwind's responsive prefixes (`sm:`, `md:`, `lg:`) already used in some places
- Mobile-first: write base styles for mobile, add breakpoints for larger screens
- Test with browser DevTools mobile simulation (320px, 375px, 414px widths)

**Safe area CSS pattern:**
```css
padding-bottom: env(safe-area-inset-bottom, 0);
```

### 2. Calm Visual Identity (UI-02)
**Confidence: HIGH**

**What needs changing:**
- Replace achromatic color palette with calm blue/green tones (CONTEXT decision)
- Keep generous whitespace (already present)
- Remove any gamification aesthetics (none exist currently — good)
- System dark/light via `prefers-color-scheme` media query (CONTEXT decision)

**Color approach:**
- Use oklch format (already in place) for the calm blue/green palette
- Primary: muted teal/blue-green for calm, studious feel
- Keep neutral grays for backgrounds and borders
- Accent: soft complementary color for interactive elements
- Apply to existing CSS custom properties in app.css — no structural changes needed

**Dark mode implementation:**
- Currently uses `.dark` class variant
- Need to switch to `@media (prefers-color-scheme: dark)` for automatic system detection
- Or add JS that sets `.dark` class based on `matchMedia('(prefers-color-scheme: dark)')` on mount

### 3. Landing Page (UI-03)
**Confidence: HIGH**

**What needs building:**
- Replace current two-card landing page with full marketing page
- Sections (from CONTEXT): Hero, Features, "Så funkar det" (3-step), FAQ, CTA
- Swedish language throughout
- Main CTA: "Kom igång gratis" linking to `/register`
- Explain BYOK model: "Du använder din egen Claude API-nyckel"
- Target audience: parents (trustworthy, educational feel)

**Structure (Plughorse-style long page):**
1. Hero: headline, subtitle, CTA button
2. Features: 3-4 feature cards (Socratic method, spaced repetition, etc.)
3. "Så funkar det": numbered steps (1. Register + API key, 2. Add child, 3. Child studies)
4. FAQ: accordion or simple Q&A (pricing? data? how does BYOK work?)
5. Footer CTA: repeat "Kom igång gratis"

**No new dependencies needed** — can build entirely with Tailwind + existing shadcn components.

### 4. Hamburger Menu Implementation
**Confidence: HIGH**

**Approach:**
- Mobile: hamburger icon (Lucide `Menu` / `X` icons) toggles a slide-in or overlay panel
- Desktop: keep current horizontal nav links
- Breakpoint: `md:` (768px) — hamburger below, horizontal above
- State: simple `$state(false)` toggle in layout.svelte
- Accessibility: `aria-expanded`, `aria-controls`, focus trap optional for v1

### 5. Dark Mode via System Preference
**Confidence: HIGH**

**Current:** `.dark` class variant exists in CSS but nothing applies it.

**Implementation:**
- In `+layout.svelte` onMount: check `matchMedia('(prefers-color-scheme: dark)')`
- Set/remove `.dark` class on `document.documentElement`
- Listen for changes with `addEventListener('change', ...)`
- No manual toggle (explicitly out of scope per REQUIREMENTS)

## Risk Assessment

| Risk | Severity | Mitigation |
|------|----------|------------|
| Chat page mobile layout complexity | Medium | Focus on fixed bottom input + scrollable messages |
| Color changes break readability | Low | Use oklch lightness values similar to current; test both modes |
| Landing page scope creep | Low | Fixed sections from CONTEXT; no backend changes |
| Safe area CSS browser support | Low | `env()` supported in iOS 11.2+, all modern Android Chrome |

## Dependencies

- No backend changes required for Phase 5
- No new npm packages needed (Lucide icons already available for menu icon)
- No database changes
- Pure frontend/CSS work

## Recommended Plan Structure

Three plans, all in Wave 1 (independent work):

1. **Visual Identity + Dark Mode** — Update app.css color palette, implement system dark/light detection
2. **Mobile Responsiveness + Navigation** — Hamburger menu, touch targets, safe areas, verify all pages
3. **Landing Page** — Full Plughorse-style landing page replacing current +page.svelte

Wave 1 parallelism works because: color changes are in app.css variables (CSS custom properties), navigation changes are in +layout.svelte, and the landing page is a self-contained +page.svelte rewrite. No conflicts.

## RESEARCH COMPLETE
