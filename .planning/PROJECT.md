# Nioplugget

## What This Is

En AI-driven studieapp för nationella proven i åk 9. Elever chattar med en AI-lärare som ställer ledande frågor istället för att ge färdiga svar. Fokus på tre ämnen: Biologi (NO), Samhällskunskap (SO) och Matematik. Inspirerad av Plughorse men specialiserad för nationella provet.

## Core Value

Eleven lär sig genom dialog — AI:n ger aldrig direkta svar utan guidar eleven med ledande frågor, anpassade efter elevens nivå.

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] Förälder kan registrera sig och logga in
- [ ] Förälder kan lägga in sin Claude API-nyckel (krypterad)
- [ ] Förälder kan skapa barnprofil via invite-länk
- [ ] Barn kan logga in via invite-länk + pin
- [ ] Elev kan välja ämne, ämnesområde och övning
- [ ] Elev kan chatta med AI-lärare i dialogbaserade övningspass
- [ ] AI ställer ledande frågor, ger aldrig direkta svar
- [ ] Övningspass har skräddarsydda system-prompts per övning
- [ ] Sessioner sparas med meddelanden och score
- [ ] Spaced repetition (SM-2) planerar repetition baserat på score
- [ ] Startsidan visar "Dags att repetera"-kort
- [ ] Färdiga övningspass mappade mot kursplanen (~50-60 st)
- [ ] Mobilanpassat, rent och lugnt UI
- [ ] Progress-vy visar styrkor/svagheter per ämne

### Out of Scope

- Betalning/prenumeration — användaren kopplar sin egen Claude API-nyckel
- Fotouppladdning av läxor — bara färdiga övningspass
- Realtidschat mellan elever — individuell AI-dialog
- Mobilapp — webbaserat, mobilanpassat
- Andra ämnen utöver Bio, Samhälle, Matte — kan läggas till senare

## Context

- Inspirerad av Plughorse (plughorse.com) som erbjuder AI-baserad läxhjälp för åk 7-9
- Nationella proven i åk 9 är viktiga milstolpar — stort behov av övningsmaterial
- Kursinnehåll mappat mot Skolverkets centrala innehåll
- Pedagogisk modell: "Frågor, inte svar" + spaced repetition
- Ämnen: Biologi (Ekologi, Kroppen, Genetik, Cellen), Samhällskunskap (Demokrati, Rättigheter, Ekonomi, Lag & rätt), Matematik (Algebra, Geometri, Statistik, Samband & förändring)

## Constraints

- **Tech stack**: Go (Chi) + SvelteKit (shadcn-svelte) + PostgreSQL + Claude API — matchar befintlig erfarenhet
- **API-nyckel**: Användaren tillhandahåller sin egen Claude API-nyckel, krypterad med AES-256 at rest
- **Säkerhet**: AI:n ska vara strikt begränsad till ämnet, inte kunna utnyttjas för fusk
- **UX**: Mobilanpassat, rent lugnt UI, korta pass (10-15 min)

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Monorepo med Go backend + SvelteKit frontend | Matchar befintlig erfarenhet från trade-analyst, tydlig separation | — Pending |
| Användaren kopplar egen Claude API-nyckel | Inget behov av betalningsinfrastruktur, open source-vänligt | — Pending |
| SM-2 spaced repetition | Beprövad algoritm, enkel att implementera, effektiv för inlärning | — Pending |
| Färdiga övningspass istället för fritext | Mer kontrollerad upplevelse, bättre kvalitet på AI-dialog | — Pending |
| Förälder-barn-modell (som Plughorse) | Föräldern hanterar API-nyckel och överblick, barnet studerar | — Pending |

---
*Last updated: 2026-04-03 after initialization*
