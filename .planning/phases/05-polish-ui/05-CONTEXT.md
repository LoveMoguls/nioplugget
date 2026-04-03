# Phase 5: Polish + UI - Context

**Gathered:** 2026-04-03
**Status:** Ready for planning

<domain>
## Phase Boundary

Mobile-ready responsive design, calm visual identity, landing page with service info. Covers UI-01..03.

</domain>

<decisions>
## Implementation Decisions

### Landing page
- Plughorse-style long page: hero section, features, "Så funkar det" (3-step), FAQ, CTA
- Main CTA: "Kom igång gratis" — links to registration
- Explain BYOK model clearly: "Du använder din egen Claude API-nyckel"
- Swedish language throughout

### Mobile polish
- Hamburger menu for mobile navigation (not bottom tab bar)
- Touch-friendly tap targets (min 44px)
- Responsive breakpoints: mobile-first design
- Safe area handling for notched devices

### Visual identity
- Color palette: calm blue/green tones — focused, studious feel
- No gamification aesthetics (no badges, trophies, fire streaks)
- Clean typography, generous whitespace
- Follow system dark/light preference (prefers-color-scheme)

### Claude's Discretion
- Exact color values (hex/hsl)
- Font choices (system fonts vs custom)
- Landing page copy/text
- FAQ content
- Animation/transition details
- Hamburger menu behavior (slide-in vs overlay)

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `frontend/src/routes/+page.svelte` — current landing page (basic)
- `frontend/src/routes/+layout.svelte` — layout with navigation
- shadcn-svelte component library already configured
- Tailwind CSS v4 already set up

### Established Patterns
- Tailwind utility classes
- shadcn-svelte components
- SvelteKit layouts and routes

### Integration Points
- Restyle existing pages to match new visual identity
- Update +layout.svelte with hamburger menu
- Rewrite +page.svelte as full landing page

</code_context>

<specifics>
## Specific Ideas

- Landing page should feel trustworthy and educational — parents are the audience
- "Kom igång gratis" — emphasize that it's free (you just need an API key)
- Reference Plughorse's structure but with own identity

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 05-polish-ui*
*Context gathered: 2026-04-03*
