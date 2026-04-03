# Project Research Summary

**Project:** Nioplugget
**Domain:** AI-driven educational chat app — Swedish nationella prov (åk 9)
**Researched:** 2026-04-03
**Confidence:** HIGH

## Executive Summary

Nioplugget is a Socratic AI tutoring app built around Swedish national tests for year-9 students. The product is narrow and deep: roughly 50-60 hand-crafted exercises mapped against Skolverket's centrala innehåll for Biologi, Samhällskunskap, and Matematik, each with a per-exercise system prompt that steers Claude to ask guiding questions and never reveal direct answers. The family account model (parent stores API key, child logs in via invite link + PIN) eliminates subscription infrastructure entirely and keeps the cost model transparent — parents pay only for what they use via their own Claude API key. Research confirms this Bring-Your-Own-Key (BYOK) approach is the correct monetization strategy for a pre-validation product in this space.

The recommended stack is Go 1.24+ (Chi) on the backend with SvelteKit (Svelte 5 + shadcn-svelte) on the frontend, backed by PostgreSQL 16+ and the official anthropic-sdk-go v1.29.0. This matches the team's existing trade-analyst project experience, which reduces ramp-up risk significantly. The two technical decisions that must be made deliberately from day one are: (1) AES-256-GCM encryption for the stored Claude API key with a random nonce per encryption, and (2) a server-side SSE proxy pattern where the Go backend decrypts the key per-request and streams Anthropic's SSE response to the browser — the API key must never touch the frontend. Spaced repetition via the SM-2 algorithm, running server-side on session completion, is the learning flywheel.

The biggest risks are not technical — they are pedagogical and security-related. The AI silently abandoning Socratic discipline (answering directly instead of guiding) is the single most dangerous product failure mode; it undermines the entire value proposition without any visible error. Layered defenses are required: per-exercise system prompts with implicit answer guidance, output filtering, and rate limiting on session messages. Secondary risks include API key leakage via application logs, GDPR Article 8 compliance for under-13 users (Swedish DPA priority enforcement area), and PIN brute-force on child accounts. All three must be addressed during the auth and chat phases, not added later.

## Key Findings

### Recommended Stack

The stack is a Go monorepo with a SvelteKit frontend served separately in development, converging to a single deployment in production. Backend uses Chi v5 for routing, pgx v5 + sqlc for type-safe database access, golang-migrate for schema management, and go-chi/jwtauth for JWT middleware. The anthropic-sdk-go v1.29.0 provides native streaming support and per-request API key injection via `option.WithAPIKey()`. Frontend uses Svelte 5 runes (production-stable), shadcn-svelte, and Tailwind v4. All versions are verified against pkg.go.dev as of April 2026.

**Core technologies:**
- Go 1.24+ with Chi v5.2.5: HTTP routing — zero external dependencies, matches team's existing experience
- SvelteKit (Svelte 5) + shadcn-svelte: Frontend — SSR + fine-grained reactivity, Tailwind v4 support confirmed
- PostgreSQL 16+ via pgx v5.9.1 + sqlc v1.30.0: Database — type-safe generated queries, no ORM magic
- anthropic-sdk-go v1.29.0: Claude API — official SDK with streaming and per-request API key support
- golang.org/x/crypto argon2id: Password hashing — OWASP 2025 gold standard over bcrypt
- crypto/aes + crypto/cipher (stdlib): AES-256-GCM for API key encryption — no extra dependencies
- go-chi/jwtauth v5.4.0: JWT auth middleware — first-party Chi integration for dual-role (parent/child) token model

**What not to use:** GORM (hides SQL), lib/pq (maintenance-only), math/rand for crypto nonces, AES-CBC (malleable), WebSockets for chat streaming (SSE is sufficient), EventSource API on frontend (requires GET; use fetch + ReadableStream for POST endpoints).

### Expected Features

The MVP is a complete vertical slice: parent registration with BYOK API key storage, child onboarding via invite link + PIN, subject/exercise navigation, AI Socratic dialogue sessions, session scoring, SM-2 spaced repetition scheduling, and a "due today" dashboard. Approximately 50-60 curriculum-mapped exercises (Bio, SO, Matte) must ship with v1 — without content, the product is an empty shell.

**Must have (table stakes):**
- Parent registration + Claude API key input (encrypted at rest) — nothing works without this
- Child invite link + PIN login — no email required for child
- Subject and exercise selection (3 subjects, 4 areas each, ~50-60 exercises)
- AI Socratic dialogue session per exercise — the product itself
- Session score + persistence — required input for spaced repetition
- SM-2 spaced repetition scheduling — the learning flywheel
- "Due today" dashboard — entry point for return visits
- Subject-locked AI guardrails — parent trust requirement
- Mobile-responsive UI — most students are on phones

**Should have (competitive):**
- Nationella prov-specific curriculum mapping — depth over breadth is the differentiator vs Khanmigo and Plughorse
- Strict Socratic discipline enforced per-exercise via dedicated system prompts — hardest to get right, highest retention value
- Swedish-language UI and AI dialogue throughout — competitors are English-first
- Clean, calm UI — deliberately avoids Duolingo anxiety mechanics (no streaks, no penalty points)
- Progress view (strength/weakness per subject) — add post-validation
- Parent overview dashboard — passive oversight without manual intervention

**Defer to v1.x / v2+:**
- Session history browse — trigger: users ask to revisit past explanations
- Email digest for due exercises — trigger: low return-visit rate
- Additional subjects beyond Bio/SO/Matte
- Multi-language support (English UI)
- Export progress PDF

### Architecture Approach

The architecture is a standard JSON REST API + SSE streaming monorepo. The Go backend owns all Claude API communication — the browser never touches the Anthropic API directly. ChatHandler decrypts the parent's AES-256-GCM encrypted API key per-request, opens an Anthropic streaming request, and proxies each SSE chunk to the browser using Go's http.Flusher interface. SM-2 runs server-side on session completion (client posts score 0-5, backend calculates next review date). The frontend is a SvelteKit SPA that consumes REST endpoints and a fetch+ReadableStream for chat streaming.

**Major components:**
1. AuthHandler (`/api/auth/*`) — parent email/password + JWT; child invite token + PIN validation; dual role claims
2. ChatHandler (`/api/chat/*`) — SSE proxy: decrypt key, build message history with system prompt, stream from Anthropic, persist on completion
3. EncryptService — AES-256-GCM encrypt/decrypt of per-user API keys; security-critical, isolated package
4. SRSService — pure Go SM-2 implementation; no DB access; input: score 0-5; output: next review date
5. ExerciseService — load exercise definition + system prompt by ID from DB
6. SvelteKit frontend — chat UI, SSE ReadableStream consumption, dashboard, auth pages

The component build order is strictly determined by dependencies: schema first, then EncryptService, then parent auth, then exercise seed data, then child auth, then ChatHandler, then SRS, then progress/dashboard. This is not negotiable — each layer depends on the previous.

### Critical Pitfalls

1. **AI gives direct answers instead of guiding questions** — The single highest-risk product failure. Claude's default helpfulness overrides Socratic constraints under student pressure or jailbreaking. Mitigation: per-exercise system prompts with implicit answer guidance (not explicit answer text), output filtering on responses, rate-limiting messages per session, and layered jailbreak detection.

2. **API key leakage via application logs** — Decrypted Claude API keys appear in logs when HTTP client debug logging is active. Mitigation: explicitly exclude `Authorization` header from all request logging; run `grep -r "sk-ant-"` against log output as a deployment gate; never log the decrypted key or any struct containing it.

3. **Invite links that are reusable or don't expire** — Permanent or long-lived invite links allow unauthorized child account creation under a parent's API quota. Mitigation: single-use tokens with 48-72 hour expiry; mark `used=true` atomically in the same DB transaction as child account creation; validate token + parent account ownership together.

4. **Unbounded conversation history causes token cost explosion** — Full session history sent to Claude grows quadratically. Mitigation: cap history at last 10 turns + system prompt from day one; track token usage per session in the DB; surface token budget in parent dashboard; use Anthropic prompt caching for static per-exercise system prompts.

5. **GDPR Article 8 for under-13 users** — Swedish DPA (IMY) has declared children's data protection a priority enforcement area. Implicit consent through parent account creation is not sufficient. Mitigation: explicit consent checkbox at parent registration scoped to child data processing; store `consent_given_at`; no third-party analytics scripts on child views; implement data deletion flow for child profile deletion.

## Implications for Roadmap

Based on research, the feature dependency chain is strict and the build order is dictated by architecture:

```
Parent auth + API key storage
    → Child auth (invite link + PIN)
        → Exercise seed data (content must exist)
            → AI chat session (core product)
                → Session scoring + persistence
                    → SM-2 scheduling
                        → "Due today" dashboard
                            → Progress view (read-only, from history)
```

### Phase 1: Foundation — Database Schema + Auth + API Key Storage

**Rationale:** Every other component depends on the schema being final and parent auth working. EncryptService must exist before parent registration can store API keys. This phase has no shortcuts.
**Delivers:** Working parent registration, login, JWT auth, encrypted API key storage, invite link generation
**Addresses:** Parent registration + API key input (P1), child invite link (P1)
**Avoids:** API key leakage (Pitfall 2), invite link reuse (Pitfall 4), PIN brute force (Pitfall 9), GDPR consent gap (Pitfall 8)
**Uses:** argon2id (password hashing), AES-256-GCM (API key encryption), go-chi/jwtauth (JWT dual-role), golang-migrate + sqlc

### Phase 2: Exercise Content + AI Chat Session

**Rationale:** Content (exercise seed data with system prompts) must exist before the chat handler can function. ChatHandler is the most complex single component — it owns SSE lifecycle, key decryption, message history management, and persistence. This is the core product and requires the most care.
**Delivers:** Working AI dialogue session for all exercises across Bio/SO/Matte; ~50-60 curriculum-mapped exercises seeded; SSE streaming visible in UI
**Addresses:** AI dialogue session (P1), curriculum-mapped exercises (P1), subject-locked guardrails (P1), subject/exercise selection (P1)
**Avoids:** AI gives direct answers (Pitfall 1), student jailbreaking (Pitfall 3), system prompt leakage (Pitfall 7), token cost explosion (Pitfall 5), streaming hang (integration gotcha)
**Implements:** ChatHandler, ExerciseService, SSE proxy pattern

### Phase 3: Session Scoring + Spaced Repetition

**Rationale:** SM-2 depends on sessions existing and having scores. Session completion flow (POST score → SM-2 → upsert srs_card) is a clean unit that can be built and tested independently once chat works.
**Delivers:** End-of-session score submission, SM-2 scheduling, srs_cards table populated, "due today" query working
**Addresses:** Session score + persistence (P1), SM-2 scheduling (P1), "due today" dashboard (P1)
**Avoids:** SM-2 ease hell (Pitfall 6), SM-2 on frontend anti-pattern (Architecture), timezone handling bug (integration gotcha)
**Implements:** SRSService (pure Go, testable without DB), session completion endpoint

### Phase 4: Dashboard + Progress Views

**Rationale:** Pure read layer over existing data. Cannot build until sessions + SRS exist. Lower risk because it introduces no new state mutations.
**Delivers:** "Due today" cards by subject, session history browse, per-subject strength/weakness breakdown, parent child overview
**Addresses:** "Due today" dashboard (P1), progress view (P2), parent overview (P2), session history (P2)
**Avoids:** Flat list overwhelm UX pitfall (group by subject, cap at 5 per group), timezone display bug for due cards
**Implements:** ProgressHandler, dashboard SvelteKit routes

### Phase 5: Polish + Security Hardening

**Rationale:** Security review and UX polish after all features work end-to-end. Deliberately last — not because it's unimportant, but because the attack surface is fully visible only when the whole system is assembled.
**Delivers:** Log audit (no `sk-ant-` in output), IDOR tests for session/child ownership, mobile browser testing (iOS Safari + Android Chrome), rate limiting verified on PIN and chat endpoints, GDPR consent flow verified
**Addresses:** All "looks done but isn't" checklist items from PITFALLS.md
**Avoids:** All 5 critical pitfalls; all security mistake patterns documented in PITFALLS.md

### Phase Ordering Rationale

- Schema must be final before any code is written against it — migrations are the single source of truth
- EncryptService before parent auth — registration stores an encrypted key; no encrypt service = broken registration from day one
- Exercise content before ChatHandler testing — the handler needs exercises in the DB to fetch system prompts
- SM-2 after chat — requires sessions with scores; building it earlier means testing against fixtures only
- Dashboard last among features — purely derived from existing data; no new state mutations
- Security hardening as a dedicated phase — the full attack surface is only visible when the system is complete; partial audits miss cross-component vulnerabilities

### Research Flags

Phases likely needing deeper research during planning:
- **Phase 2 (Exercise content):** The per-exercise system prompt design is the highest-value and hardest-to-get-right work. Needs deliberate design work: how to encode curriculum goals as implicit guidance without including answer text that Claude might leak. Consider a prompt review checklist before seeding.
- **Phase 2 (Socratic guardrails):** Output filtering approach (second Claude call vs regex vs embedding similarity) needs a concrete decision before implementation begins. Research confirms all three are used in production Socratic tutors; the right choice depends on latency tolerance.
- **Phase 3 (SM-2 variant):** Standard SM-2 vs SM-2+ (Blue Raja's variant) should be decided before implementation. SM-2+ handles overdue cards better, which matters for an exam-prep product with uneven study cadence.

Phases with standard patterns (skip research-phase):
- **Phase 1 (Auth + encryption):** AES-256-GCM, argon2id, JWT dual-role — all are well-documented patterns with verified Go implementations. No novel integration.
- **Phase 4 (Dashboard/progress):** Read-only queries over existing data. Standard SvelteKit routing + Chi handlers. No novel patterns.
- **Phase 5 (Security hardening):** Checklist-driven verification of already-implemented security controls. No new research needed.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | All versions verified against pkg.go.dev as of Apr 2026; matches team's existing trade-analyst project |
| Features | MEDIUM-HIGH | Table stakes well-validated by competitor analysis (Khanmigo, Plughorse, Duolingo). SM-2 feature set confirmed by multiple sources. Some UX patterns from community sources only. |
| Architecture | HIGH | SSE proxy pattern verified against official Anthropic docs and Go SDK; per-request API key confirmed via `option.WithAPIKey()`; project structure validated against trade-analyst conventions |
| Pitfalls | MEDIUM-HIGH | Critical security pitfalls (key logging, GDPR, invite link, PIN brute force) sourced from OWASP, IMY official docs, and Anthropic official docs. Socratic failure mode confirmed by peer-reviewed research (arxiv). SM-2 ease hell pattern from widely-cited community analysis. |

**Overall confidence:** HIGH

### Gaps to Address

- **Socratic output filter implementation:** Research confirms the need but does not settle the approach (second Claude call vs regex vs embedding similarity). Decision should be made in Phase 2 planning before any chat code is written. A second Claude call is the simplest and most reliable but adds latency; regex is fast but brittle.
- **Exercise system prompt design process:** ~50-60 exercises with high-quality Socratic prompts is substantial content work. Research does not quantify time needed. This is likely the longest-running work item in Phase 2 and may need to proceed in parallel with ChatHandler development.
- **Session self-rating UX for SM-2:** Architecture research notes that session score can come from student self-rating (1-5) or an AI-generated score (extra Claude call). Self-rating is simpler for v1; AI scoring is more accurate. This decision should be explicit in Phase 3 planning.
- **GDPR data retention policy:** Research identifies the requirement but does not specify retention periods. A concrete policy (e.g., "session messages deleted after 12 months") needs to be defined before launch.

## Sources

### Primary (HIGH confidence)
- [pkg.go.dev/github.com/anthropics/anthropic-sdk-go](https://pkg.go.dev/github.com/anthropics/anthropic-sdk-go) — v1.29.0, Go 1.23+ requirement, streaming API
- [pkg.go.dev/github.com/go-chi/chi/v5](https://pkg.go.dev/github.com/go-chi/chi/v5) — v5.2.5
- [pkg.go.dev/github.com/jackc/pgx/v5](https://pkg.go.dev/github.com/jackc/pgx/v5) — v5.9.1
- [platform.claude.com/docs/en/build-with-claude/streaming](https://platform.claude.com/docs/en/build-with-claude/streaming) — SSE event types, streaming patterns
- [platform.claude.com/docs/en/test-and-evaluate/strengthen-guardrails/reduce-prompt-leak](https://platform.claude.com/docs/en/test-and-evaluate/strengthen-guardrails/reduce-prompt-leak) — system prompt protection
- [OWASP LLM01:2025 Prompt Injection](https://genai.owasp.org/llmrisk/llm01-prompt-injection/) — jailbreak risk classification
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html) — argon2id recommendation
- [IMY Children's Rights on Digital Platforms](https://www.imy.se/globalassets/dokument/rapporter/the-rights-of-children-and-young-people-on-digital-platforms_accessible.pdf) — GDPR Article 8, Swedish enforcement
- [SocraticAI: Transforming LLMs into Guided CS Tutors (arxiv)](https://arxiv.org/abs/2512.03501) — Socratic failure modes, jailbreak patterns
- [docs.sqlc.dev — pgx/v5 target](https://docs.sqlc.dev/en/stable/guides/using-go-and-pgx.html)

### Secondary (MEDIUM confidence)
- [shadcn-svelte.com/docs/installation/sveltekit](https://www.shadcn-svelte.com/docs/installation/sveltekit) — Svelte 5 + Tailwind v4 support
- [SM-2 Algorithm — tegaru.app](https://tegaru.app/en/blog/sm2-algorithm-explained) — algorithm parameters
- [SM-2+ (Blue Raja)](https://www.blueraja.com/blog/477/a-better-spaced-repetition-learning-algorithm-sm2) — overdue card handling improvement
- [Plughorse.com](https://www.plughorse.com) — Swedish competitor feature set
- [AI tutoring RCT in UK classrooms (arxiv)](https://arxiv.org/html/2512.23633v1) — tutoring model validation
- [testcontainers-go Postgres module](https://golang.testcontainers.org/modules/postgres/) — integration test patterns

### Tertiary (LOW confidence)
- [Duolingo AI Innovations (foralink.io)](https://foralink.io/blogs/duolingos-ai-innovations-transforming-language-learning-and-beyond) — gamification patterns to avoid
- [Khanmigo reviews (kidsaitools.com)](https://www.kidsaitools.com/en/articles/khanmigo-review-parents-complete-2026) — competitor UX observations

---
*Research completed: 2026-04-03*
*Ready for roadmap: yes*
