# Roadmap: Nioplugget

## Overview

Nioplugget delivers a Socratic AI study app for Swedish nationella prov (åk 9). The build order is dictated by strict dependencies: parent auth and encrypted API key storage must exist before anything else works; exercise content must be seeded before the chat handler can fetch system prompts; session scores must exist before SM-2 can schedule repetitions; and the dashboard is a pure read layer over data that earlier phases produce. The final phase verifies that the assembled system is mobile-ready, visually calm, and security-hardened end-to-end.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [ ] **Phase 1: Foundation** - Parent auth, child invite flow, API key encryption, and security baseline
- [ ] **Phase 2: Content + AI Chat** - Exercise seed data (~50-60 exercises) and the core Socratic dialogue product
- [ ] **Phase 3: Spaced Repetition** - Session scoring, SM-2 scheduling, and the due-today flywheel
- [ ] **Phase 4: Progress Views** - Read-only strength/weakness and parent overview derived from session history
- [ ] **Phase 5: Polish + UI** - Mobile-first verification, landing page, and UX hardening

## Phase Details

### Phase 1: Foundation
**Goal**: Parents can securely register, store their Claude API key, and onboard a child — the prerequisite for everything else in the product
**Depends on**: Nothing (first phase)
**Requirements**: AUTH-01, AUTH-02, AUTH-03, AUTH-04, AUTH-05, AUTH-06, AUTH-07, AUTH-08, KEY-01, KEY-02, KEY-03, KEY-04, SEC-01, SEC-02, SEC-03
**Success Criteria** (what must be TRUE):
  1. A parent can register with email and password, log in, and remain logged in across page reloads
  2. A parent can enter their Claude API key and it is stored encrypted (AES-256-GCM); the parent can update or delete it
  3. A parent can generate an invite link for a child and the child can activate the account with a self-chosen PIN
  4. A child can log in using their name + PIN; repeated wrong PINs are rate-limited
  5. The Go backend never logs API keys or Authorization headers; invite links are single-use and expire after 72 hours
**Plans**: 5 plans
  - [x] 01-01-PLAN.md — Project scaffolding, DB schema, sqlc, redacting logger
  - [ ] 01-02-PLAN.md — Parent auth (Argon2id, JWT, register/login/logout)
  - [ ] 01-03-PLAN.md — AES-256-GCM encryption + API key CRUD + validation
  - [ ] 01-04-PLAN.md — Child profiles, invite links, activation, PIN login, rate limiting
  - [ ] 01-05-PLAN.md — SvelteKit frontend: all parent and child user flows

### Phase 2: Content + AI Chat
**Goal**: A student can select a subject, area, and exercise, then complete a full Socratic dialogue session where the AI guides with questions and never gives direct answers
**Depends on**: Phase 1
**Requirements**: CONT-01, CONT-02, CONT-03, CONT-04, CONT-05, CHAT-01, CHAT-02, CHAT-03, CHAT-04, CHAT-05, CHAT-06, CHAT-07, CHAT-08, UI-04
**Success Criteria** (what must be TRUE):
  1. All three subjects (Biologi, Samhällskunskap, Matematik) have exercises seeded across their four areas each, with per-exercise system prompts aligned to Skolverkets centrala innehåll
  2. A student can navigate subject → area → exercise and start a session
  3. AI responses stream in real time via SSE and the AI consistently asks guiding questions without revealing direct answers
  4. A student can end a session manually; all messages are persisted in the database
  5. The AI refuses to answer off-topic questions and redirects the student back to the exercise subject
**Plans**: TBD

### Phase 3: Spaced Repetition
**Goal**: Every completed session produces a score that feeds SM-2 scheduling, so returning students see exactly which exercises are due for review
**Depends on**: Phase 2
**Requirements**: SRS-01, SRS-02, SRS-03, SRS-04
**Success Criteria** (what must be TRUE):
  1. A session can be completed with a score (1-5), either via self-rating or AI assessment, and the score is persisted
  2. The SM-2 algorithm calculates the next review date from the score; ease factor never drops below 1.3
  3. The home screen shows "Dags att repetera"-cards for all exercises whose review date has passed
**Plans**: TBD

### Phase 4: Progress Views
**Goal**: Students can see their own strengths and weaknesses per subject, and parents can see an overview of their child's activity — both derived purely from existing session data
**Depends on**: Phase 3
**Requirements**: PROG-01, PROG-02, PROG-03
**Success Criteria** (what must be TRUE):
  1. A student can view a per-subject progress overview showing completed exercises and session scores
  2. A student can identify their weakest areas based on session score history
  3. A parent can view their child's progress and subject overview after logging in
**Plans**: TBD

### Phase 5: Polish + UI
**Goal**: The assembled product is mobile-ready, visually calm, and passes a full end-to-end security and UX review
**Depends on**: Phase 4
**Requirements**: UI-01, UI-02, UI-03
**Success Criteria** (what must be TRUE):
  1. All views are usable on a mobile phone (iOS Safari and Android Chrome) without horizontal scroll or broken layouts
  2. The UI is visually calm — no distracting animations, no streak counters, no gamification pressure
  3. A landing page exists that describes the service and provides registration and login entry points
**Plans**: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 → 2 → 3 → 4 → 5

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. Foundation | 3/5 | In Progress|  |
| 2. Content + AI Chat | 0/? | Not started | - |
| 3. Spaced Repetition | 0/? | Not started | - |
| 4. Progress Views | 0/? | Not started | - |
| 5. Polish + UI | 0/? | Not started | - |
