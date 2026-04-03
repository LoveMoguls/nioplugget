# Phase 2: Content + AI Chat - Research

**Researched:** 2026-04-03
**Domain:** Claude API streaming, exercise content seeding, Socratic chat UX
**Confidence:** HIGH

## Summary

Phase 2 builds two major subsystems: (1) a curriculum-mapped exercise catalog seeded via database migrations, and (2) a streaming Socratic AI chat engine using the official Anthropic Go SDK. The backend proxies Claude API requests using the parent's encrypted API key (already implemented in Phase 1), streams SSE responses to the SvelteKit frontend, and persists all messages in PostgreSQL. Session scoring uses a separate Claude call at session end.

The official `github.com/anthropics/anthropic-sdk-go` SDK (v1.29.0, April 2026) provides first-class streaming support via `client.Messages.NewStreaming()` with an accumulate pattern. The Go backend acts as an SSE proxy: it receives the student's message, decrypts the parent's API key, calls Claude with the exercise system prompt + conversation history, and forwards text deltas as SSE events to the browser. The frontend uses the native `EventSource` API or `fetch` with `ReadableStream` to render bubbles word-by-word.

**Primary recommendation:** Use the official Anthropic Go SDK for streaming, seed exercises via golang-migrate SQL migrations, and keep the system prompt as the single Socratic guardrail layer.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- Strictly never give direct answers — even after repeated student requests, only give hints and leading questions
- Off-topic: short dismissal + redirect ("Bra fraga, men lat oss fokusera pa [amnet]. Sa...")
- No output filtering beyond system prompt — trust the system prompt, keep it simple and cheap
- Always respond in Swedish, regardless of student's language
- System prompt is the single layer of defense — must be well-crafted per exercise
- Bubble-style messages (like iMessage) — student right, AI left, rounded bubbles with color
- Streaming word-by-word via SSE, with typing indicator while AI thinks
- Session end: manual only — student clicks "Avsluta pass", no time limit
- When session ends: AI gives a short summary (2-3 sentences)
- Summary is the final message before session closes
- Detailed system prompts per exercise: specific learning goals, example questions to ask, common misconceptions to address
- Navigation: 3-step hierarchy — Amne -> Omrade -> Ovning
- Claude model: claude-sonnet-4-6 (hardcoded, not configurable in v1)
- Each topic area: 3-5 exercises with increasing difficulty
- AI-assessment: extra Claude call when session ends to evaluate student performance (1-5)
- Score shown to student with short feedback text
- Score feeds into SM-2 algorithm (Phase 3)
- Conversation truncation: sliding window — keep system prompt + last N messages

### Claude's Discretion
- Difficulty indicator visibility (stars vs hidden)
- Exact sliding window size (N messages to keep)
- Typing indicator animation style
- Exercise card layout in navigation views
- Score feedback text tone and format

### Deferred Ideas (OUT OF SCOPE)
- None — discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| CONT-01 | Biologi exercises (Ekologi, Kroppen, Genetik, Cellen) | SQL seed migration with system prompts per exercise |
| CONT-02 | Samhallskunskap exercises (Demokrati, Rattigheter, Ekonomi, Lag & ratt) | SQL seed migration with system prompts per exercise |
| CONT-03 | Matematik exercises (Algebra, Geometri, Statistik, Samband & forandring) | SQL seed migration with system prompts per exercise |
| CONT-04 | 3-5 exercises per area with increasing difficulty | difficulty_order column in exercises table |
| CONT-05 | Per-exercise system prompt following Skolverket | system_prompt TEXT column, crafted per exercise |
| CHAT-01 | Student navigates subject -> area -> exercise and starts session | REST endpoints + SvelteKit routes for 3-step navigation |
| CHAT-02 | AI asks guiding questions, never direct answers | System prompt design with Socratic constraints |
| CHAT-03 | Real-time streaming via SSE | Anthropic Go SDK NewStreaming + SSE proxy handler |
| CHAT-04 | AI adapts level based on student responses | System prompt instruction to adjust complexity |
| CHAT-05 | AI stays on topic, redirects off-topic | System prompt with explicit off-topic redirect instructions |
| CHAT-06 | Student can end session manually | POST /api/sessions/:id/end endpoint + UI button |
| CHAT-07 | All messages persisted per session | messages table with session_id FK |
| CHAT-08 | Conversation truncation for token costs | Sliding window: system prompt + last N messages |
| UI-04 | Chat view with message history and input | Bubble-style chat component with streaming support |
</phase_requirements>

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| github.com/anthropics/anthropic-sdk-go | v1.29.0 | Claude API client with streaming | Official SDK, first-class SSE support, type-safe |
| github.com/go-chi/chi/v5 | v5.2.5 | HTTP router (already in project) | Already used in Phase 1 |
| github.com/jackc/pgx/v5 | v5.9.1 | PostgreSQL driver (already in project) | Already used in Phase 1 |
| sqlc | v1.30.0 | Type-safe SQL queries (already in project) | Already used in Phase 1 |
| golang-migrate | existing | Database migrations (already in project) | Already used in Phase 1 |

### Frontend
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| SvelteKit | 5.x | Frontend framework (already in project) | Already used |
| bits-ui | 2.16.5+ | shadcn-svelte primitives (already in project) | Already used |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Official Anthropic SDK | liushuangls/go-anthropic | Community library, less maintained than official |
| SSE proxy in Go | Direct browser-to-Claude | Exposes API key to client — unacceptable for BYOK model |
| EventSource API | fetch + ReadableStream | EventSource simpler but no POST support; use fetch for POST with streaming response |

## Architecture Patterns

### Recommended Project Structure
```
backend/
├── internal/
│   ├── chat/           # Chat handler, SSE streaming, session management
│   │   ├── handler.go  # HTTP handlers for sessions and messages
│   │   ├── store.go    # Database interface (ChatStore)
│   │   ├── stream.go   # SSE streaming + Claude API proxy
│   │   └── prompt.go   # System prompt builder + sliding window
│   ├── content/        # Content browsing handlers
│   │   ├── handler.go  # GET /subjects, /topics, /exercises
│   │   └── store.go    # ContentStore interface
│   └── database/
│       ├── migrations/
│       │   └── 002_content_and_sessions.up.sql
│       └── queries/
│           ├── subjects.sql
│           ├── topics.sql
│           ├── exercises.sql
│           ├── sessions.sql
│           └── messages.sql
frontend/
├── src/
│   ├── routes/
│   │   ├── study/
│   │   │   ├── +page.svelte          # Subject list
│   │   │   ├── [subject]/
│   │   │   │   └── +page.svelte      # Topic list
│   │   │   └── [subject]/[topic]/
│   │   │       └── +page.svelte      # Exercise list
│   │   └── chat/
│   │       └── [sessionId]/
│   │           └── +page.svelte      # Chat interface
│   └── lib/
│       ├── components/
│       │   ├── chat/
│       │   │   ├── ChatBubble.svelte  # Message bubble
│       │   │   ├── ChatInput.svelte   # Input with send
│       │   │   └── TypingIndicator.svelte
│       │   └── content/
│       │       └── ExerciseCard.svelte
│       └── stores/
│           └── chat.ts               # Chat state management
```

### Pattern 1: SSE Streaming Proxy
**What:** Backend receives student message, calls Claude with streaming, proxies SSE events to browser
**When to use:** Every chat message exchange
**Example:**
```go
// Source: https://platform.claude.com/docs/en/api/sdks/go (Streaming section)
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
    // Set SSE headers
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "streaming not supported", http.StatusInternalServerError)
        return
    }

    // Create client with parent's decrypted API key
    client := anthropic.NewClient(option.WithAPIKey(decryptedKey))
    
    stream := client.Messages.NewStreaming(r.Context(), anthropic.MessageNewParams{
        Model:     anthropic.ModelClaudeSonnet4_6,
        MaxTokens: 1024,
        System:    []anthropic.TextBlockParam{{Text: exercise.SystemPrompt}},
        Messages:  buildMessageHistory(sessionMessages),
    })

    message := anthropic.Message{}
    for stream.Next() {
        event := stream.Current()
        message.Accumulate(event)
        switch ev := event.AsAny().(type) {
        case anthropic.ContentBlockDeltaEvent:
            switch delta := ev.Delta.AsAny().(type) {
            case anthropic.TextDelta:
                fmt.Fprintf(w, "data: %s\n\n", jsonEncode(delta.Text))
                flusher.Flush()
            }
        }
    }
    if stream.Err() != nil {
        // Handle error
    }
    // Save complete AI message to DB
}
```

### Pattern 2: Sliding Window Conversation Truncation
**What:** Keep system prompt + last N messages to control token costs
**When to use:** Before every Claude API call
**Example:**
```go
func buildMessageHistory(messages []Message, windowSize int) []anthropic.MessageParam {
    // System prompt is passed separately, not in messages
    if len(messages) > windowSize {
        messages = messages[len(messages)-windowSize:]
    }
    // Ensure first message is from user (Claude requirement)
    params := make([]anthropic.MessageParam, 0, len(messages))
    for _, m := range messages {
        if m.Role == "user" {
            params = append(params, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
        } else {
            params = append(params, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
        }
    }
    return params
}
```

### Pattern 3: Session Scoring via Separate Claude Call
**What:** When student ends session, make a non-streaming Claude call to evaluate performance
**When to use:** POST /api/sessions/:id/end
**Example:**
```go
func (h *ChatHandler) EndSession(w http.ResponseWriter, r *http.Request) {
    // 1. Generate summary message (streaming, shown to student)
    // 2. Generate score (non-streaming, separate call)
    scoreResp, _ := client.Messages.New(ctx, anthropic.MessageNewParams{
        Model:     anthropic.ModelClaudeSonnet4_6,
        MaxTokens: 256,
        System:    []anthropic.TextBlockParam{{Text: scoringSystemPrompt}},
        Messages:  buildMessageHistory(sessionMessages),
    })
    // Parse score (1-5) + feedback from response
    // 3. Save score + summary to session record
    // 4. Return score + feedback to frontend
}
```

### Anti-Patterns to Avoid
- **Exposing API key to frontend:** Never send the decrypted API key to the browser. Always proxy through backend.
- **Unbounded conversation history:** Always apply sliding window before sending to Claude. Long conversations will hit token limits and cost unnecessarily.
- **Blocking SSE writes:** Always flush after each SSE event. Buffered writes cause the client to receive chunks instead of streaming.
- **Hardcoding system prompts in Go code:** Store in database as part of exercise seed data. Allows iteration without redeployment.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Claude API client | Custom HTTP + SSE parser | anthropic-sdk-go | Official SDK handles auth, retries, streaming protocol, error types |
| SSE parsing in browser | Custom fetch + text parsing | Native EventSource or fetch with getReader() | Well-tested, handles reconnection |
| UUID generation | Custom UUID logic | PostgreSQL gen_random_uuid() | Already established pattern in Phase 1 |
| Database migrations | Manual SQL execution | golang-migrate | Already established pattern in Phase 1 |

## Common Pitfalls

### Pitfall 1: WriteTimeout Kills SSE Streams
**What goes wrong:** Go's default `http.Server.WriteTimeout` (currently 15s in main.go) closes the connection mid-stream
**Why it happens:** SSE streams can last minutes; the server timeout is designed for normal request-response
**How to avoid:** Either remove WriteTimeout for the SSE handler, use `http.ResponseController` to extend deadlines per-request, or set a longer timeout for the SSE route
**Warning signs:** Streams cut off after exactly 15 seconds

### Pitfall 2: Claude Messages Must Alternate Roles
**What goes wrong:** Sending two consecutive user or assistant messages causes API error
**Why it happens:** Claude API requires strict user/assistant alternation
**How to avoid:** When applying sliding window, ensure the first message is from user and messages alternate. Merge or drop consecutive same-role messages if truncation breaks the pattern.
**Warning signs:** 400 Bad Request from Claude API

### Pitfall 3: Frontend EventSource Only Supports GET
**What goes wrong:** Cannot use EventSource API for POST requests (sending user message)
**Why it happens:** The EventSource spec only supports GET
**How to avoid:** Use `fetch()` with `ReadableStream` for the POST endpoint. Parse SSE events manually from the stream.
**Warning signs:** Trying to POST with EventSource fails silently

### Pitfall 4: System Prompt Quality Determines Everything
**What goes wrong:** AI gives direct answers, goes off-topic, or responds in English
**Why it happens:** Weak or vague system prompts don't constrain Claude effectively
**How to avoid:** Detailed system prompts per exercise with explicit rules: always Swedish, never direct answers, specific misconceptions to probe, redirect off-topic attempts
**Warning signs:** AI behavior varies wildly between exercises

### Pitfall 5: Missing Flusher Check
**What goes wrong:** SSE handler panics or silently fails to stream
**Why it happens:** Not all ResponseWriters implement http.Flusher (e.g., behind certain reverse proxies)
**How to avoid:** Always check `w.(http.Flusher)` and return error if not supported
**Warning signs:** No streaming output but no errors

## Code Examples

### Database Schema for Phase 2
```sql
-- Migration: 002_content_and_sessions.up.sql
CREATE TABLE subjects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    slug TEXT NOT NULL UNIQUE,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE topics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subject_id UUID NOT NULL REFERENCES subjects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(subject_id, slug)
);

CREATE TABLE exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    topic_id UUID NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    difficulty_order INT NOT NULL DEFAULT 1,
    system_prompt TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    score INT,  -- 1-5, NULL until session ended
    summary TEXT,  -- AI-generated summary
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ
);

CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('user', 'assistant')),
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_student_id ON sessions(student_id);
CREATE INDEX idx_sessions_exercise_id ON sessions(exercise_id);
CREATE INDEX idx_messages_session_id ON messages(session_id);
```

### SSE Frontend Pattern (fetch + ReadableStream)
```typescript
// Source: Standard Web Streams API pattern
async function streamChat(sessionId: string, message: string, onChunk: (text: string) => void) {
    const res = await fetch(`${API_BASE}/api/sessions/${sessionId}/messages`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: message }),
    });
    
    if (!res.ok) throw new Error('Failed to send message');
    
    const reader = res.body!.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    
    while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        
        buffer += decoder.decode(value, { stream: true });
        const lines = buffer.split('\n');
        buffer = lines.pop() || '';
        
        for (const line of lines) {
            if (line.startsWith('data: ')) {
                const data = line.slice(6);
                if (data === '[DONE]') return;
                onChunk(JSON.parse(data));
            }
        }
    }
}
```

### System Prompt Template
```text
Du ar en pedagogisk AI-larare som hjalper elever i arskurs 7-9 att forsta {amne}: {omrade}.

OVNING: {ovningstitel}
LARANDEMAL: {specifika larandemal fran Skolverkets centrala innehall}

REGLER (bryts ALDRIG):
1. Svara ALLTID pa svenska
2. Ge ALDRIG direkta svar. Stall istallet ledande fragor som hjalper eleven tanka sjalv.
3. Om eleven fragar nagot utanfor amnet: "Bra fraga, men lat oss fokusera pa {omrade}. Sa..."
4. Anpassa spraksvarigheten till elevens niva
5. Om eleven verkar fast, ge en ledtrad, inte svaret

VANLIGA MISSUPPFATTNINGAR ATT UTFORSKA:
- {missuppfattning 1}
- {missuppfattning 2}

EXEMPELFRAGOR ATT STALLA:
- {fraga 1}
- {fraga 2}

Borja med att halsa och stalla en oppningsfrage om {specifikt amnesomrade}.
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| liushuangls/go-anthropic (community) | anthropics/anthropic-sdk-go (official) | 2024 | Official SDK with better streaming support and maintenance |
| WebSocket for real-time | SSE for AI streaming | Standard practice | SSE is simpler, unidirectional (AI->client), no upgrade negotiation needed |
| Full conversation in every request | Sliding window truncation | Standard practice | Controls token costs for long sessions |

## Open Questions

1. **Sliding window size (N)**
   - What we know: Need to keep system prompt + last N messages
   - What's unclear: Optimal N for balance between context and cost
   - Recommendation: Start with N=20 (Claude's Discretion), monitor token usage, adjust later

2. **Scoring prompt design**
   - What we know: Separate Claude call at session end, score 1-5
   - What's unclear: Exact scoring criteria prompt wording
   - Recommendation: Design a generic scoring prompt that receives the conversation + exercise learning goals and returns structured JSON {score: int, feedback: string}

3. **Summary generation approach**
   - What we know: AI gives 2-3 sentence summary when session ends
   - What's unclear: Whether summary and score should be one call or two
   - Recommendation: Single call that returns both summary and score as structured output to save API costs

## Sources

### Primary (HIGH confidence)
- [Anthropic Go SDK official docs](https://platform.claude.com/docs/en/api/sdks/go) - Streaming API, MessageNewParams, system prompt format
- [anthropics/anthropic-sdk-go GitHub](https://github.com/anthropics/anthropic-sdk-go) - v1.29.0, April 2026
- [pkg.go.dev/anthropic-sdk-go](https://pkg.go.dev/github.com/anthropics/anthropic-sdk-go) - Full API reference

### Secondary (MEDIUM confidence)
- Existing codebase analysis - Phase 1 patterns (Chi router, sqlc, JWT auth, encryption service)

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - Official SDK, established project patterns
- Architecture: HIGH - Standard SSE proxy pattern, well-documented SDK
- Pitfalls: HIGH - Well-known Go SSE gotchas, Claude API constraints documented

**Research date:** 2026-04-03
**Valid until:** 2026-05-03
