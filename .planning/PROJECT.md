# Nioplugget

## What This Is

En AI-driven studieapp för nationella proven i åk 9. Elever chattar med en AI-lärare som ställer ledande frågor istället för att ge färdiga svar — aldrig direkta svar, alltid dialog. Ämnen: Biologi, Samhällskunskap, Matematik, Kemi, Fysik, Geografi och Historia med NP-kalibrerade övningspass mappade mot Skolverkets centrala innehåll. Förälder tillhandahåller egen Claude API-nyckel (BYOK), hanterar barnprofiler och kan följa barnets progress.

## Current Milestone: v1.1 Fler ämnen

**Goal:** Expand subject coverage from 3 to 7 by adding Kemi, Fysik, Geografi och Historia — same NP-calibrated quality as v1.0.

**Target features:**
- Kemi: 4 ämnesområden × 3 övningar (12 övningar)
- Fysik: 4 ämnesområden × 3 övningar (12 övningar)
- Geografi: 4 ämnesområden × 3 övningar (12 övningar)
- Historia: 4 ämnesområden × 3 övningar (12 övningar)
- Alla med NP-kalibrerade system-prompts (E/C/A verbhierarki)

## Core Value

Eleven lär sig genom dialog — AI:n ger aldrig direkta svar utan guidar eleven med ledande frågor, anpassade efter nationella provets betygsnivåer (E/C/A).

## Requirements

### Validated

- ✓ Förälder kan registrera sig och logga in — v1.0
- ✓ Förälder kan lägga in sin Claude API-nyckel (krypterad AES-256-GCM) — v1.0
- ✓ Förälder kan skapa barnprofil via invite-länk — v1.0
- ✓ Barn kan logga in via invite-länk + PIN (rate-limited) — v1.0
- ✓ Elev kan välja ämne, ämnesområde och övning (3-stegsnavigation) — v1.0
- ✓ Elev kan chatta med AI-lärare i dialogbaserade övningspass — v1.0
- ✓ AI ställer ledande frågor, ger aldrig direkta svar — v1.0
- ✓ 37 övningspass med NP-kalibrerade system-prompts (NP-verbhierarki: Beskriv/Förklara/Resonera) — v1.0 (Phase 6)
- ✓ Sessioner sparas med meddelanden och AI-genererad score (1-5) — v1.0
- ✓ Spaced repetition (SM-2) planerar repetition baserat på score — v1.0
- ✓ Startsidan visar "Dags att repetera"-kort — v1.0
- ✓ Mobilanpassat (mobile-first), responsiv design med hamburger-meny — v1.0
- ✓ Rent, lugnt UI (blå/grön palett, systemets mörkt/ljust, inga distraktioner) — v1.0
- ✓ Landningssida med hero, features, how-it-works, BYOK-förklaring och FAQ — v1.0
- ✓ Progress-vy visar styrkor/svagheter per ämne (elev + förälder) — v1.0

### Active

- [ ] Kemi: 4 ämnesområden med 3 NP-kalibrerade övningar per område
- [ ] Fysik: 4 ämnesområden med 3 NP-kalibrerade övningar per område
- [ ] Geografi: 4 ämnesområden med 3 NP-kalibrerade övningar per område
- [ ] Historia: 4 ämnesområden med 3 NP-kalibrerade övningar per område

### Out of Scope

- Betalning/prenumeration — BYOK-modell, användaren kopplar sin egen Claude API-nyckel
- Fotouppladdning av läxor — bara färdiga övningspass för kontrollerad kvalitet
- Realtidschat mellan elever — individuell AI-dialog, inte socialt
- Mobilapp (native) — web-first, mobilanpassat — native app i framtiden
- Gamification (streaks, poäng, badges) — anti-feature: skapar ångest, inte lärande
- OAuth/social login — e-post/lösenord räcker för v1
- Mörkt/ljust tema-toggle — följ systemets inställning

## Context

**Shipped v1.0 MVP 2026-04-06.**

- ~10,400 LOC (Go + TypeScript + Svelte)
- Stack: Go (Chi) + SvelteKit (shadcn-svelte, Tailwind 4) + PostgreSQL + Claude API
- 37 NP-kalibrerade övningspass: Biologi (12), Samhällskunskap (13), Matematik (12)
- All 39 v1 requirements delivered
- Timeline: 3 days (2026-04-03 → 2026-04-06)

Known issues / technical debt:
- GDPR data retention policy (session messages) not defined — decide before launch
- In-memory PIN rate limiter resets on server restart (acceptable for single-process MVP)
- AI adaptation (CHAT-04) is prompt-based, not ML-driven — may need refinement based on real usage

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Monorepo: Go backend + SvelteKit frontend | Matches existing experience, clear separation | ✓ Good — clean build, no friction |
| BYOK (Bring Your Own Key) | No payment infrastructure, open source-friendly | ✓ Good — simplifies v1 significantly |
| SM-2 spaced repetition | Proven algorithm, simple to implement, effective | ✓ Good — SM-2 (not SM-2+) sufficient for MVP |
| Färdiga övningspass (not free-text) | Controlled experience, better AI dialogue quality | ✓ Good — 37 exercises cover NP curriculum well |
| Förälder-barn-modell | Parent handles API key and oversight, child studies | ✓ Good — clean separation of concerns |
| Socratic guardrail: separate scoring Claude call | Avoids regex fragility for output filtering | ✓ Good — reliable AI-based session scoring |
| SM-2 (not SM-2+ Blue Raja) | Standard SM-2 sufficient for MVP | ✓ Good — simpler, easier to reason about |
| httpOnly + Secure + SameSite=Lax JWT cookie | XSS cannot steal token, CSRF mitigated | ✓ Good — secure auth baseline |
| Flat SvelteKit routes (not route groups) | Route groups caused /login conflict | ✓ Good — avoided fatal routing bug |
| CSS-only bar charts for progress | No chart library dependency | ✓ Good — lightweight, maintainable |
| NP verb hierarchy in prompts (E/C/A) | Aligns with real nationella prov scoring | ✓ Good — Phase 6 upgrade significantly improves quality |
| Goroutine for SM-2 update after session | Never blocks response | ✓ Good — clean async pattern |
| Logger redacts at middleware level | No risk of accidentally logging secrets | ✓ Good — security baseline from day 1 |

## Constraints

- **Tech stack**: Go (Chi) + SvelteKit (shadcn-svelte) + PostgreSQL + Claude API — matches existing experience
- **API key**: User provides own Claude API key, encrypted AES-256-GCM at rest
- **Security**: AI strictly limited to subject — cannot be exploited for cheating
- **UX**: Mobile-first, calm UI, short sessions (10-15 min)
- **Language**: Swedish UI (target audience is Swedish åk 9 students and parents)

---
*Last updated: 2026-04-06 after v1.1 milestone start*
