# Nioplugget — Design Spec

## Overview

AI-driven studieapp for nationella proven i åk 9. Eleven chattar med en AI-lärare som ställer ledande frågor istället för att ge svar. Fokus på tre ämnen: Biologi (NO), Samhällskunskap (SO) och Matematik.

Inspirerad av [Plughorse](https://plughorse.com/sv) men specialiserad för nationella provet.

## Tech Stack

- **Backend:** Go (Chi) — API, auth, spaced repetition, Claude-proxy
- **Frontend:** SvelteKit + shadcn-svelte — SSR, UI, chattvy
- **Database:** PostgreSQL
- **AI:** Claude API (användaren kopplar sin egen nyckel)
- **Repo:** GitHub, publikt, open source

## Architecture

```
nioplugget/
├── backend/     # Go (Chi)
├── frontend/    # SvelteKit
└── db/          # PostgreSQL migrations
```

```
┌─────────────┐     ┌──────────────┐     ┌────────────┐
│  SvelteKit   │────▶│  Go (Chi)    │────▶│ PostgreSQL │
│  Frontend    │◀────│  Backend API │◀────│            │
└─────────────┘     └──────┬───────┘     └────────────┘
                           │
                           ▼
                    ┌──────────────┐
                    │  Claude API  │
                    │ (user's key) │
                    └──────────────┘
```

## Data Model

### parents
- id (UUID, PK)
- email (unique)
- password_hash
- created_at

### students
- id (UUID, PK)
- parent_id (FK → parents)
- name
- invite_code (unique)
- created_at

### api_keys
- id (UUID, PK)
- parent_id (FK → parents)
- encrypted_key (AES-256 encrypted Claude API key)

### subjects
- id (UUID, PK)
- name (Biologi, Samhällskunskap, Matematik)

### topics
- id (UUID, PK)
- subject_id (FK → subjects)
- name
- description

### exercises
- id (UUID, PK)
- topic_id (FK → topics)
- title
- system_prompt
- difficulty (1-3)

### sessions
- id (UUID, PK)
- student_id (FK → students)
- exercise_id (FK → exercises)
- started_at
- completed_at
- score (1-5, nullable)

### messages
- id (UUID, PK)
- session_id (FK → sessions)
- role (user/assistant)
- content
- created_at

### review_schedule
- id (UUID, PK)
- student_id (FK → students)
- topic_id (FK → topics)
- next_review (timestamp)
- interval (days)
- ease_factor (float, default 2.5)

## Auth Flow

1. Förälder registrerar sig (e-post + lösenord)
2. Förälder lägger in sin Claude API-nyckel (krypteras med AES-256)
3. Förälder skapar barnprofil → genererar invite-länk
4. Barn klickar länk → väljer namn/pin → loggas in

## AI Dialog & Pedagogy

### System Prompt Strategy

Every exercise has a tailored system prompt instructing Claude to:
- **Never give direct answers** — only ask leading questions
- Adapt difficulty based on student responses
- Stay strictly on topic
- End with a brief summary of what the student learned

Example (Biology — Ecology):
```
Du är en tålmodig NO-lärare som hjälper en elev i åk 9 att förbereda sig
för nationella provet i biologi. Ämnet är ekologi och näringskedjor.

Regler:
- Ge ALDRIG direkta svar. Ställ ledande frågor.
- Om eleven svarar fel, ge en ledtråd och fråga igen.
- Håll dig till ämnet. Om eleven avviker, styr tillbaka.
- Använd konkreta exempel från svensk natur.
- Avsluta passet med en sammanfattning på 2-3 meningar.
```

### Session Lifecycle

1. Student selects subject → topic → exercise
2. Backend creates session, loads system prompt + conversation history
3. Student chats — each message proxied via backend to Claude API with user's key
4. Backend saves each message to `messages` table
5. Session ends (student says "klar" or 15 min) → backend asks Claude to assess performance (1-5) → saves score
6. Score feeds into spaced repetition algorithm → updates `review_schedule`

### Spaced Repetition (SM-2)

- Each topic per student has: `next_review`, `interval`, `ease_factor`
- Score 1-2 → interval resets (needs more practice)
- Score 3 → interval maintained
- Score 4-5 → interval extended
- Homepage shows "Dags att repetera" for topics past `next_review`

## Frontend & UX

### Pages

**Parent:**
- `/` — Landing page (info, register/login)
- `/dashboard` — Overview: children, API key, invite child
- `/dashboard/child/:id` — Child's progress, subject overview

**Student:**
- `/study` — Home: "Dags att repetera" cards + choose subject
- `/study/:subject` — Topic areas
- `/study/:subject/:topic` — Exercise list
- `/chat/:sessionId` — Chat view with AI teacher
- `/progress` — My progress: subject overview, strengths/weaknesses

### Chat View (core experience)

```
┌─────────────────────────────────┐
│ ◀ Biologi > Ekologi            │
│ Övning: Näringskedjor          │
├─────────────────────────────────┤
│                                 │
│  AI: Vad händer med energin     │
│      när en organism äts av     │
│      en annan?                  │
│                                 │
│           Den försvinner? Elev  │
│                                 │
│  AI: Inte riktigt! Tänk på     │
│      vad kroppen använder       │
│      energin till...            │
│                                 │
├─────────────────────────────────┤
│ [Skriv ditt svar...    ] [Skicka]│
│ [Avsluta pass]                  │
└─────────────────────────────────┘
```

### Design Principles
- **Mobile-first** — many students use phones
- **Clean, calm UI** — focus on readability, no distractions
- **shadcn-svelte** component library
- Light/dark theme

## Course Content

Exercises mapped to Skolverket's central content for åk 7-9:

### Biologi (NO)
- **Ekologi** — näringskedjor, ekosystem, kretslopp
- **Kroppen** — organsystem, hälsa, sexualitet
- **Genetik** — arv, DNA, evolution
- **Cellen** — cellens delar, fotosyntes, cellandning

### Samhällskunskap (SO)
- **Demokrati** — riksdag, regering, val, grundlagar
- **Rättigheter** — mänskliga rättigheter, FN, barnkonventionen
- **Ekonomi** — privatekonomi, samhällsekonomi, globalisering
- **Lag & rätt** — rättssystemet, brott, straff

### Matematik
- **Algebra** — ekvationer, uttryck, funktioner
- **Geometri** — area, volym, Pythagoras, likformighet
- **Statistik** — medelvärde, median, diagram, sannolikhet
- **Samband & förändring** — proportionalitet, procent, grafer

Each topic area gets 3-5 exercises with increasing difficulty. ~50-60 exercises total at launch. System prompts stored in database, seeded via migrations.
