# Architecture Research

**Domain:** AI-driven educational chat app (Go + SvelteKit + PostgreSQL + Claude API)
**Researched:** 2026-04-03
**Confidence:** HIGH (patterns verified via official docs, existing codebase reference, multiple sources)

## Standard Architecture

### System Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                      Browser (SvelteKit SPA)                      │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐  │
│  │  Auth pages  │  │  Study chat  │  │  Dashboard / progress  │  │
│  │  (parent +   │  │  (SSE stream │  │  (spaced repetition    │  │
│  │   child PIN) │  │   + history) │  │   cards + analytics)   │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬─────────────┘  │
└─────────┼─────────────────┼─────────────────────┼────────────────┘
          │ JSON REST        │ fetch + ReadableStream│ JSON REST
          ▼                  ▼                       ▼
┌──────────────────────────────────────────────────────────────────┐
│                       Go (Chi) HTTP Server                        │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────────┐  │
│  │  AuthHandler │  │  ChatHandler │  │   ProgressHandler      │  │
│  │  /api/auth/* │  │  /api/chat/* │  │   /api/subjects/*      │  │
│  └──────┬───────┘  └──────┬───────┘  └──────────┬─────────────┘  │
│         │                 │                       │               │
│  ┌──────▼─────────────────▼───────────────────────▼────────────┐  │
│  │              Service Layer                                    │  │
│  │  AuthService │ ChatService │ EncryptService │ SRSService     │  │
│  └──────┬────────────┬────────────────────────────┬────────────┘  │
│         │            │ (per-request API key)        │              │
│         │            ▼                             │              │
│         │  ┌──────────────────────┐               │              │
│         │  │  Anthropic Go SDK    │               │              │
│         │  │  client.Messages.    │               │              │
│         │  │  NewStreaming(...)   │               │              │
│         │  └──────────────────────┘               │              │
└─────────┼───────────────────────────────────────┼──────────────┘
          │                                         │
          ▼                                         ▼
┌──────────────────────────────────────────────────────────────────┐
│                          PostgreSQL                               │
│  users │ children │ sessions │ messages │ exercises │ srs_cards  │
└──────────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Responsibility | Typical Implementation |
|-----------|----------------|------------------------|
| SvelteKit frontend | All UI, SSE consumption, state management | SvelteKit + shadcn-svelte, no SSR needed (static adapter or SPA mode) |
| Go Chi HTTP server | API routing, auth middleware, request validation | Chi v5, single binary, serves API on `/api/*` |
| AuthHandler | Registration, login, invite-link generation, PIN validation | JWT or session cookie, separate flows for parent vs child |
| ChatHandler | Receive user message, decrypt API key, open SSE stream to Claude, proxy back | SSE response to browser, writes completed message to DB after stream |
| EncryptService | AES-256-GCM encrypt/decrypt of per-user API keys | Go stdlib `crypto/aes` + `crypto/cipher`, key stored in env |
| SRSService | SM-2 algorithm, schedule next review, return due cards | Pure Go, no external dep, runs on session completion |
| ExerciseService | Load exercise definition + system prompt by ID | Read from DB exercises table populated at migration time |
| PostgreSQL | Persistent storage for all entities | Accessed via sqlc (type-safe queries), no ORM |

## Recommended Project Structure

```
nioplugget/
├── cmd/
│   └── server/
│       └── main.go             # Entry point, wire dependencies
├── internal/
│   ├── auth/
│   │   └── auth.go             # JWT/session helpers
│   ├── config/
│   │   └── config.go           # Env-based config struct
│   ├── database/
│   │   ├── db.go               # pgx connection pool
│   │   ├── queries.sql.go      # sqlc-generated
│   │   └── schema.sql          # DDL
│   ├── encrypt/
│   │   └── encrypt.go          # AES-256-GCM service
│   ├── handler/
│   │   ├── auth.go             # Parent register/login, child invite/PIN
│   │   ├── chat.go             # SSE streaming chat endpoint
│   │   ├── exercises.go        # List subjects/areas/exercises
│   │   ├── progress.go         # SRS cards due, stats per subject
│   │   ├── routes.go           # Chi router wiring
│   │   └── middleware.go       # RequireAuth, RequireParent, RequireChild
│   ├── middleware/
│   │   └── auth.go             # JWT validation middleware
│   └── srs/
│       └── sm2.go              # SM-2 algorithm, pure Go
├── migrations/
│   └── 001_initial.sql
├── frontend/
│   ├── src/
│   │   ├── lib/
│   │   │   ├── api.ts          # Typed fetch wrappers
│   │   │   ├── stores/         # Svelte stores (auth, chat, srs)
│   │   │   └── components/     # Reusable UI pieces
│   │   └── routes/
│   │       ├── +layout.svelte  # Root layout (auth guard)
│   │       ├── login/          # Parent login
│   │       ├── register/       # Parent registration
│   │       ├── invite/[token]/ # Child onboarding via invite link
│   │       ├── study/          # Subject/exercise picker
│   │       │   └── [sessionId]/# Active chat session
│   │       └── dashboard/      # Progress, due cards, parent view
│   ├── svelte.config.js
│   └── package.json
├── sqlc.yaml
├── go.mod
├── Makefile                    # make dev (air + vite concurrently)
└── docker-compose.yml          # PostgreSQL for local dev
```

### Structure Rationale

- **internal/encrypt/:** Isolated service — API key crypto is security-critical, should be tested independently and never leak into handler logic directly.
- **internal/srs/:** Pure Go, no DB access — takes session score as input, returns next review date. Can be unit-tested without DB.
- **internal/handler/chat.go:** The most complex handler. Owns the full SSE lifecycle: decrypt key, build messages array with system prompt, stream from Anthropic SDK, accumulate, persist.
- **frontend/routes/study/[sessionId]/:** Session ID is server-generated at session start. Frontend navigates to it; SSE stream is keyed on session ID.

## Architectural Patterns

### Pattern 1: Claude API Proxy via SSE (the core pattern)

**What:** Go handler opens streaming request to Anthropic API using user's decrypted API key, proxies each SSE chunk back to the browser as its own SSE event.

**When to use:** Any time AI response needs to appear token-by-token in the UI. This is the correct approach — never call Anthropic from the browser (key exposure).

**Trade-offs:** Requires Go handler to hold two open connections simultaneously (browser SSE + Anthropic SSE). Go goroutines make this cheap. Messages are persisted after stream completes (on `message_stop` event), not mid-stream.

**Example:**
```go
func (h *ChatHandler) Stream(w http.ResponseWriter, r *http.Request) {
    // 1. Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    flusher := w.(http.Flusher)

    // 2. Decrypt user's API key from DB
    encryptedKey, _ := h.queries.GetAPIKey(r.Context(), userID)
    apiKey, _ := h.encrypt.Decrypt(encryptedKey)

    // 3. Build messages array (history + new user message)
    messages := buildMessageHistory(existingMessages, newUserMessage)

    // 4. Stream from Anthropic with per-request key
    client := anthropic.NewClient(option.WithAPIKey(apiKey))
    stream := client.Messages.NewStreaming(r.Context(), anthropic.MessageNewParams{
        Model:    anthropic.ModelClaude3_5HaikuLatest,
        System:   []anthropic.TextBlockParam{{Text: exerciseSystemPrompt}},
        Messages: messages,
        MaxTokens: 1024,
    })

    var accumulated strings.Builder
    for stream.Next() {
        event := stream.Current()
        if delta, ok := event.AsAny().(anthropic.ContentBlockDeltaEvent); ok {
            if text, ok := delta.Delta.AsAny().(anthropic.TextDelta); ok {
                fmt.Fprintf(w, "data: %s\n\n", text.Text)
                flusher.Flush()
                accumulated.WriteString(text.Text)
            }
        }
    }

    // 5. Persist completed assistant message
    h.queries.InsertMessage(r.Context(), sessionID, "assistant", accumulated.String())
    fmt.Fprintf(w, "data: [DONE]\n\n")
    flusher.Flush()
}
```

### Pattern 2: Per-Request API Key Decryption

**What:** The Anthropic Go SDK supports `option.WithAPIKey(key)` on a per-call basis. The key is never stored in the client struct — decrypted only for the duration of the request, then goes out of scope.

**When to use:** Always. Never instantiate a shared Anthropic client with a hardcoded key. Each request decrypts the current user's key from DB, creates a scoped client.

**Trade-offs:** Slight overhead from key lookup + decrypt per request (microseconds). Worth it — required for multi-user safety.

**Example:**
```go
// EncryptService wraps AES-256-GCM
func (e *EncryptService) Decrypt(ciphertext []byte) (string, error) {
    block, _ := aes.NewCipher(e.masterKey) // masterKey from env
    gcm, _ := cipher.NewGCM(block)
    nonce, ct := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
    plain, err := gcm.Open(nil, nonce, ct, nil)
    return string(plain), err
}
```

### Pattern 3: SM-2 on Backend, Triggered Post-Session

**What:** After a study session ends (user closes chat or clicks "done"), the frontend POSTs the session score (0-5). The backend runs SM-2 to calculate next review date and upserts the SRS card record.

**When to use:** All spaced repetition scheduling logic lives on the backend. Frontend only reads due cards.

**Trade-offs:** The "score" must be determined somehow — either: (a) AI scores the session at end of chat (extra Claude call), or (b) student self-rates 1-5. Simpler to start with self-rating; add AI scoring later.

### Pattern 4: Parent/Child Dual Auth

**What:** Two user types with different auth flows. Parent: email + password → JWT. Child: invite token + PIN → separate short-lived session cookie or JWT with `role: child` claim. Child JWT has restricted permissions (chat only, no settings).

**When to use:** Required by the project's family account model.

**Trade-offs:** Two middleware guards needed: `RequireAuth` (both) and `RequireParent` (settings, API key management).

## Data Flow

### Chat Session Flow

```
User types message
    ↓
SvelteKit: POST /api/chat/sessions/{id}/messages
    { content: "Vad är fotosyntes?" }
    ↓
ChatHandler:
    1. Validate child auth + session ownership
    2. Fetch exercise system prompt from DB
    3. Fetch message history for session
    4. Decrypt parent's API key from DB
    5. Append new user message to history
    6. Insert user message to DB (before streaming)
    ↓
    Open SSE response to browser
    Open Anthropic streaming request (with decrypted key + system prompt)
    ↓
    For each text chunk from Anthropic:
        Write "data: {chunk}\n\n" → browser
        Flush
        Append to accumulator
    ↓
    Stream complete (message_stop event):
        Insert assistant message to DB
        Write "data: [DONE]\n\n" → browser
        Close SSE
    ↓
SvelteKit: ReadableStream reader processes chunks
    Appends tokens to reactive message variable
    Shows typing indicator until [DONE]
```

### Session Completion + SRS Flow

```
User clicks "Avsluta pass" (or session auto-closes)
    ↓
SvelteKit: POST /api/chat/sessions/{id}/complete
    { score: 4 }  (student self-rates 1-5)
    ↓
ChatHandler:
    1. Mark session as complete in DB
    2. Call SRSService.CalculateNextReview(cardID, score)
        → SM-2: update easiness factor, repetition count, interval
        → Returns next_review_date
    3. Upsert srs_cards record
    ↓
Response: { next_review: "2026-04-07" }
    ↓
Frontend navigates to dashboard, shows next due card
```

### Parent API Key Setup Flow

```
Parent registers → Enters Claude API key in settings
    ↓
POST /api/settings/apikey { key: "sk-ant-..." }
    ↓
SettingsHandler:
    1. RequireParent middleware check
    2. EncryptService.Encrypt(apiKey) → []byte (AES-256-GCM)
    3. Store ciphertext + nonce in users.encrypted_api_key (bytea)
    ↓
On every chat request:
    EncryptService.Decrypt(user.encrypted_api_key) → plaintext key
    option.WithAPIKey(key) per Anthropic request
```

### SvelteKit SSE Consumption

```
SvelteKit component:
    const response = await fetch('/api/chat/sessions/{id}/messages', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${jwt}` },
        body: JSON.stringify({ content: userMessage })
    })

    const reader = response.body.getReader()
    const decoder = new TextDecoder()

    while (true) {
        const { done, value } = await reader.read()
        if (done) break
        const chunk = decoder.decode(value)
        for (const line of chunk.split('\n')) {
            if (line.startsWith('data: ')) {
                const text = line.slice(6)
                if (text === '[DONE]') { /* finalize */ break }
                assistantMessage += text  // reactive update → UI re-renders
            }
        }
    }
```

Note: Use `fetch` with `ReadableStream` rather than `EventSource` because the chat endpoint requires `POST` with body. `EventSource` only supports `GET`. (HIGH confidence — EventSource spec limitation.)

### Key Data Flows Summary

1. **Auth flow:** Parent registers → creates child profile → child gets invite link → child sets PIN → child logs in
2. **Chat flow:** Pick exercise → start session → stream AI dialogue → complete session → SM-2 schedules next review
3. **Dashboard flow:** Load due SRS cards → show "Dags att repetera" list → child starts review session

## Database Schema (Core Tables)

```
users (parents)
    id, email, password_hash, encrypted_api_key, created_at

children
    id, parent_id (FK users), name, pin_hash, invite_token, created_at

exercises
    id, subject, area, title, system_prompt, description, curriculum_ref

sessions
    id, child_id (FK children), exercise_id (FK exercises),
    started_at, completed_at, score (0-5 or NULL)

messages
    id, session_id (FK sessions), role (user/assistant),
    content, created_at

srs_cards
    id, child_id (FK children), exercise_id (FK exercises),
    repetitions, easiness_factor, interval_days,
    last_review, next_review, created_at
```

## Scaling Considerations

| Scale | Architecture Adjustments |
|-------|--------------------------|
| 0-500 users | Single Go binary + single Postgres instance. Air for dev. No queues needed. |
| 500-5K users | Add pgBouncer connection pooling. Monitor Anthropic API rate limits (per-key, not per-app). Index srs_cards.next_review + child_id. |
| 5K-50K users | Horizontal Go replicas (stateless — JWT auth, no in-memory session). Read replica for dashboard/progress queries. |
| 50K+ users | Out of scope for MVP. Anthropic per-key rate limits become the real bottleneck, not infra. |

### Scaling Priorities

1. **First bottleneck:** Anthropic API rate limits per user key (not your problem — each user has their own key). Concurrent SSE connections in Go are cheap (goroutines).
2. **Second bottleneck:** PostgreSQL connection count. Each SSE stream holds a connection for its lifetime. Add pgBouncer early if sessions are long.

## Anti-Patterns

### Anti-Pattern 1: Calling Claude API from the Browser

**What people do:** Put the Anthropic API key in frontend env vars (e.g., `VITE_ANTHROPIC_KEY`) and call the API directly from SvelteKit client code.

**Why it's wrong:** API key is exposed to anyone who opens DevTools. Malicious users can extract and abuse the key. Encrypted storage is bypassed entirely.

**Do this instead:** All Claude API calls go through the Go backend. Frontend only talks to `/api/chat/*`. Backend decrypts key, makes Anthropic request, proxies SSE back.

### Anti-Pattern 2: WebSocket for AI Streaming

**What people do:** Implement WebSocket for the chat stream, thinking bidirectional is better.

**Why it's wrong:** Chat streaming is unidirectional (AI → user). WebSocket adds complexity (upgrade handshake, reconnection logic, separate WS library) with no benefit. SSE over HTTP/1.1 is simpler, proxied correctly by most infrastructure, and Go's net/http handles it natively.

**Do this instead:** SSE via `text/event-stream`. Use `fetch` + `ReadableStream` on the frontend (not `EventSource`) because the endpoint is POST.

### Anti-Pattern 3: Storing API Keys Plaintext in DB

**What people do:** Store `claude_api_key TEXT` directly in the users table for simplicity during development.

**Why it's wrong:** DB compromise exposes all user API keys. Any log or debug output that includes the user row leaks keys.

**Do this instead:** AES-256-GCM encrypt at application layer. Store as `bytea`. Master encryption key in environment variable (never in DB). See EncryptService pattern above.

### Anti-Pattern 4: Loading Entire Message History for Every Request

**What people do:** SELECT * FROM messages WHERE session_id = ? and send all messages to Claude every turn.

**Why it's wrong:** Claude context window fills up, latency increases, cost grows linearly. For educational sessions (10-15 min), history can grow large.

**Do this instead:** For this domain (short study sessions), loading the full session history is fine — sessions are bounded. Do add a LIMIT as a safety valve (e.g., last 30 messages). Cross-session history is not needed (SM-2 handles repetition scheduling; each session starts fresh with the exercise system prompt).

### Anti-Pattern 5: Running SM-2 on the Frontend

**What people do:** Implement SM-2 in TypeScript, calculate next review date in the browser, POST it to the server.

**Why it's wrong:** Client controls their own SRS schedule. A student could manipulate intervals to never see hard cards again.

**Do this instead:** SM-2 runs server-side. Frontend POSTs score (0-5). Backend calculates intervals. Frontend is read-only for SRS data.

## Integration Points

### External Services

| Service | Integration Pattern | Notes |
|---------|---------------------|-------|
| Anthropic Claude API | `anthropic-sdk-go` with `option.WithAPIKey(perUserKey)`, streaming via `Messages.NewStreaming()` | Official Go SDK exists (github.com/anthropics/anthropic-sdk-go). Per-request key via `option.WithAPIKey()`. HIGH confidence. |
| PostgreSQL | `pgx/v5` + `sqlc` for type-safe queries | Matches trade-analyst project pattern. No ORM. |

### Internal Boundaries

| Boundary | Communication | Notes |
|----------|---------------|-------|
| SvelteKit ↔ Go API | JSON REST for most endpoints; fetch+ReadableStream for chat streaming | No shared code needed. Types manually kept in sync or via openapi codegen later. |
| ChatHandler ↔ EncryptService | Direct Go function call | Must not log decrypted key; keep in-memory only for request lifetime. |
| ChatHandler ↔ SRSService | Direct Go function call on session completion | SRS only runs after stream completes, not during. |
| ChatHandler ↔ ExerciseService | Fetch exercise system prompt by exercise ID before opening stream | System prompt is the primary guardrail for topic focus. |

## Build Order Implications

The component dependency graph dictates this build order:

1. **Database schema + migrations** — everything depends on this. Define all tables upfront.
2. **EncryptService** — required before any parent registration (stores encrypted key).
3. **AuthHandler (parent)** — register, login, JWT. Needed before anything else works end-to-end.
4. **Exercise data seed** — the ~50-60 exercises with system prompts must exist before chat works.
5. **AuthHandler (child)** — invite flow depends on parent auth existing.
6. **ChatHandler (streaming)** — depends on auth, exercises, encrypt service. Core feature.
7. **SRSService + session completion** — depends on sessions + messages existing from chat.
8. **ProgressHandler + dashboard** — reads SRS state. Depends on SRS existing.
9. **Frontend** — can be built in parallel with backend per-feature. Each route maps to a backend endpoint.

## Sources

- [Claude API Streaming Documentation](https://platform.claude.com/docs/en/build-with-claude/streaming) — Official. SSE event types, Go SDK streaming example. HIGH confidence.
- [anthropic-sdk-go option.WithAPIKey](https://pkg.go.dev/github.com/anthropics/anthropic-sdk-go/option) — Official Go pkg docs. Per-request API key confirmed. HIGH confidence.
- [Go SSE Streaming — oneuptime.com](https://oneuptime.com/blog/post/2026-01-25-server-sent-events-streaming-go/view) — SSE headers, flusher pattern, broker pattern for multiple clients. MEDIUM confidence (third-party, content verified against Go stdlib).
- [SvelteKit SSE patterns — sveltetalk.com](https://sveltetalk.com/posts/building-real-time-sveltekit-apps-with-server-sent-events) — SvelteKit SSE consumption patterns. MEDIUM confidence.
- [AES-256-GCM in Go — Twilio](https://www.twilio.com/en-us/blog/developers/community/encrypt-and-decrypt-data-in-go-with-aes-256) — AES-GCM encrypt/decrypt pattern in Go. MEDIUM confidence.
- [go-srs SM-2 Go implementation](https://github.com/revelaction/go-srs) — Existing Go SM-2 library. MEDIUM confidence.
- [Go + SvelteKit hot reload monorepo](https://1v0.dev/posts/26-go-backend-hotreload/) — Monorepo structure with Air + Concurrently. MEDIUM confidence.
- [trade-analyst project](https://github.com/trollstaven/trade-analyst) — Same team's existing Go Chi + SvelteKit project. HIGH confidence for project conventions.

---
*Architecture research for: AI-driven educational chat app (nioplugget)*
*Researched: 2026-04-03*
