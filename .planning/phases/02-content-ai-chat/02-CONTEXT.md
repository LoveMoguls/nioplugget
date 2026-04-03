# Phase 2: Content + AI Chat - Context

**Gathered:** 2026-04-03
**Status:** Ready for planning

<domain>
## Phase Boundary

Seed ~50-60 curriculum-mapped exercises and build the Socratic AI chat engine. Student navigates subject → topic → exercise, starts a streaming dialogue session, and receives a score + summary when done. Covers CONT-01..05, CHAT-01..08, UI-04.

</domain>

<decisions>
## Implementation Decisions

### Socratic guardrails
- Strictly never give direct answers — even after repeated student requests, only give hints and leading questions
- Off-topic: short dismissal + redirect ("Bra fråga, men låt oss fokusera på [ämnet]. Så...")
- No output filtering beyond system prompt — trust the system prompt, keep it simple and cheap
- Always respond in Swedish, regardless of student's language
- System prompt is the single layer of defense — must be well-crafted per exercise

### Chat experience
- Bubble-style messages (like iMessage) — student right, AI left, rounded bubbles with color
- Streaming word-by-word via SSE, with typing indicator while AI thinks
- Session end: manual only — student clicks "Avsluta pass", no time limit
- When session ends: AI gives a short summary (2-3 sentences: "Under det här passet har du lärt dig om...")
- Summary is the final message before session closes

### Exercise content
- Detailed system prompts per exercise: specific learning goals, example questions to ask, common misconceptions to address
- Navigation: 3-step hierarchy — Ämne → Område → Övning
- Difficulty display: Claude's Discretion (visible indicator or hidden)
- Claude model: claude-sonnet-4-6 (hardcoded, not configurable in v1)
- Each topic area: 3-5 exercises with increasing difficulty

### Session scoring
- AI-assessment: extra Claude call when session ends to evaluate student performance (1-5)
- Score shown to student with short feedback text ("Bra jobbat! Du verkar ha koll på grunderna.")
- Score feeds into SM-2 algorithm (Phase 3)
- Conversation truncation: sliding window — keep system prompt + last N messages (e.g., 20)

### Claude's Discretion
- Difficulty indicator visibility (stars vs hidden)
- Exact sliding window size (N messages to keep)
- Typing indicator animation style
- Exercise card layout in navigation views
- Score feedback text tone and format

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `backend/internal/apikey/encrypt.go` — EncryptionService for decrypting parent's Claude API key per request
- `backend/internal/auth/middleware.go` — ChildOnly middleware to gate student routes
- `backend/internal/database/queries/` — sqlc-generated queries for subjects, topics, exercises, sessions, messages tables
- `frontend/src/lib/api.ts` — API client for backend calls
- `frontend/src/lib/stores/auth.ts` — Auth store with role detection

### Established Patterns
- Chi router with middleware groups (backend/cmd/server/main.go)
- sqlc for type-safe DB queries
- JWT auth with role claims (parent/child)
- AES-256-GCM decrypt per request for API key access
- shadcn-svelte components for UI

### Integration Points
- Chat handler needs: decrypt API key → create Claude client → proxy SSE stream
- Exercise seed data via golang-migrate migration files
- New routes: GET /api/subjects, GET /api/subjects/:id/topics, GET /api/topics/:id/exercises, POST /api/sessions, POST /api/sessions/:id/messages, POST /api/sessions/:id/end
- Frontend routes: /study, /study/:subject, /study/:subject/:topic, /chat/:sessionId

</code_context>

<specifics>
## Specific Ideas

- System prompts should use concrete Swedish examples (Swedish nature for biology, Swedish political system for civics)
- Chat should feel like texting a patient tutor — not a formal exam
- Score feedback should be encouraging but honest — no false praise

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 02-content-ai-chat*
*Context gathered: 2026-04-03*
