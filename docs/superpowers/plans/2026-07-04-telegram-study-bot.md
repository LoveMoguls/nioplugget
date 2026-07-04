# Telegram Study Bot Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Barnet pluggar via en Telegram-bot (hela flödet: välj övning → sokratisk dialog → stjärnor/XP → SRS) medan uppladdning/paneler stannar i web-GUI:t. Spec: `docs/superpowers/specs/2026-07-04-telegram-bot-design.md`.

**Architecture:** Nytt paket `backend/internal/telegram` i samma process som servern. Long polling mot Telegram Bot API (egen minimal klient, inget nytt beroende). Återanvänder `chat`, `srs`, `challenges` som Go-funktioner. State i Postgres (`telegram_links`, `telegram_link_codes`, `telegram_sessions`, `telegram_reminders`).

**Tech Stack:** Go 1.25, chi, sqlc (pgx/v5), golang-migrate, anthropic-sdk-go, SvelteKit-frontend.

## Global Constraints

- Modulnamn: `github.com/trollstaven/nioplugget/backend` — alla imports utgår från det.
- Tom/saknad `TELEGRAM_BOT_TOKEN` ⇒ Telegram-lagret startar inte; befintlig drift opåverkad.
- All användartext på svenska. Felhantering: logga med zerolog (`github.com/rs/zerolog/log`), krascha aldrig polling-loopen.
- AI-anrop använder alltid förälderns dekrypterade BYOK-nyckel (mönstret i `internal/chat/handler.go:469-486`).
- Telegram-kanalen får ALDRIG skicka LaTeX — extra systemprompt-suffix (Task 7).
- Meddelanden > 4096 tecken delas (Telegrams hårda gräns).
- Kör kommandon från `backend/`-katalogen om inget annat anges. DB måste vara igång för migrate: `docker start nioplugget-db`.
- Migrate-kommando: `set -a && source .env && set +a && migrate -path db/migrations -database "$DATABASE_URL" up`
- Commit efter varje grönt test-steg. Co-Authored-By-trailer enligt repo-praxis.

---

### Task 1: DB-migration + sqlc-queries för Telegram-tabellerna

**Files:**
- Create: `backend/db/migrations/012_telegram.up.sql`
- Create: `backend/db/migrations/012_telegram.down.sql`
- Create: `backend/db/queries/telegram.sql`
- Generated: `backend/internal/database/queries/telegram.sql.go` (via sqlc)

**Interfaces:**
- Consumes: befintliga tabeller `students`, `sessions`.
- Produces: sqlc-genererade metoder som senare tasks anropar: `CreateTelegramLinkCode`, `GetTelegramLinkCode`, `MarkTelegramLinkCodeUsed`, `UpsertTelegramLink`, `GetTelegramLinkByTelegramUserID`, `ListTelegramLinks`, `GetTelegramSession`, `UpsertTelegramSession`, `TryInsertTelegramReminder` samt typerna `queries.TelegramLink`, `queries.TelegramLinkCode`, `queries.TelegramSession` och paramstructar.

- [ ] **Step 1: Skriv up-migrationen**

`backend/db/migrations/012_telegram.up.sql`:

```sql
CREATE TABLE telegram_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL UNIQUE REFERENCES students(id) ON DELETE CASCADE,
    telegram_user_id BIGINT NOT NULL UNIQUE,
    chat_id BIGINT NOT NULL,
    linked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE telegram_link_codes (
    code TEXT PRIMARY KEY,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ
);

CREATE TABLE telegram_sessions (
    chat_id BIGINT PRIMARY KEY,
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    state TEXT NOT NULL DEFAULT 'menu',
    active_session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE telegram_reminders (
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    sent_on DATE NOT NULL,
    PRIMARY KEY (student_id, sent_on)
);
```

- [ ] **Step 2: Skriv down-migrationen**

`backend/db/migrations/012_telegram.down.sql`:

```sql
DROP TABLE telegram_reminders;
DROP TABLE telegram_sessions;
DROP TABLE telegram_link_codes;
DROP TABLE telegram_links;
```

- [ ] **Step 3: Skriv queries**

`backend/db/queries/telegram.sql`:

```sql
-- name: CreateTelegramLinkCode :one
INSERT INTO telegram_link_codes (code, student_id, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTelegramLinkCode :one
SELECT * FROM telegram_link_codes WHERE code = $1;

-- name: MarkTelegramLinkCodeUsed :exec
UPDATE telegram_link_codes SET used_at = NOW() WHERE code = $1;

-- name: UpsertTelegramLink :one
INSERT INTO telegram_links (student_id, telegram_user_id, chat_id)
VALUES ($1, $2, $3)
ON CONFLICT (student_id)
DO UPDATE SET telegram_user_id = $2, chat_id = $3, linked_at = NOW()
RETURNING *;

-- name: GetTelegramLinkByTelegramUserID :one
SELECT * FROM telegram_links WHERE telegram_user_id = $1;

-- name: ListTelegramLinks :many
SELECT * FROM telegram_links;

-- name: GetTelegramSession :one
SELECT * FROM telegram_sessions WHERE chat_id = $1;

-- name: UpsertTelegramSession :one
INSERT INTO telegram_sessions (chat_id, student_id, state, active_session_id, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (chat_id)
DO UPDATE SET student_id = $2, state = $3, active_session_id = $4, updated_at = NOW()
RETURNING *;

-- name: TryInsertTelegramReminder :execrows
INSERT INTO telegram_reminders (student_id, sent_on)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
```

- [ ] **Step 4: Kör migration + generera + bygg**

```bash
docker start nioplugget-db && sleep 3
set -a && source .env && set +a && migrate -path db/migrations -database "$DATABASE_URL" up
sqlc generate
go build ./...
```
Expected: migrationen kör utan fel; `internal/database/queries/telegram.sql.go` skapas; bygget grönt.

- [ ] **Step 5: Commit**

```bash
git add db/ internal/database/queries/
git commit -m "feat(telegram): db schema and queries for links, sessions, reminders"
```

---

### Task 2: Icke-strömmande chatt-svar — `chat.CompleteChatResponse`

**Files:**
- Create: `backend/internal/chat/complete.go`
- Test: `backend/internal/chat/complete_test.go`

**Interfaces:**
- Produces: `func CompleteChatResponse(ctx context.Context, apiKey, systemPrompt string, messages []anthropic.MessageParam, opts ...option.RequestOption) (string, error)` — Telegram-lagrets AI-anrop (Task 7).

- [ ] **Step 1: Skriv failande test**

`backend/internal/chat/complete_test.go`:

```go
package chat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func TestCompleteChatResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"msg_test","type":"message","role":"assistant","model":"claude-sonnet-4-6","content":[{"type":"text","text":"Vad tror du själv?"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`))
	}))
	defer srv.Close()

	got, err := CompleteChatResponse(context.Background(), "test-key", "system",
		[]anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock("hej"))},
		option.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Vad tror du själv?" {
		t.Errorf("got %q, want %q", got, "Vad tror du själv?")
	}
}
```

- [ ] **Step 2: Verifiera att det failar**

Run: `go test ./internal/chat/ -run TestCompleteChatResponse -v`
Expected: FAIL — `undefined: CompleteChatResponse`

- [ ] **Step 3: Implementera**

`backend/internal/chat/complete.go`:

```go
package chat

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// CompleteChatResponse gets a full (non-streaming) Claude reply. Used by
// channels that cannot stream, e.g. the Telegram bot. Extra request options
// (like option.WithBaseURL in tests) may be passed.
func CompleteChatResponse(ctx context.Context, apiKey, systemPrompt string, messages []anthropic.MessageParam, opts ...option.RequestOption) (string, error) {
	clientOpts := append([]option.RequestOption{option.WithAPIKey(apiKey)}, opts...)
	client := anthropic.NewClient(clientOpts...)
	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 1024,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Messages:  messages,
	})
	if err != nil {
		return "", fmt.Errorf("claude completion: %w", err)
	}
	var text string
	for _, block := range resp.Content {
		if b, ok := block.AsAny().(anthropic.TextBlock); ok {
			text += b.Text
		}
	}
	return text, nil
}
```

- [ ] **Step 4: Verifiera grönt**

Run: `go test ./internal/chat/ -run TestCompleteChatResponse -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/chat/complete.go internal/chat/complete_test.go
git commit -m "feat(chat): non-streaming CompleteChatResponse for non-SSE channels"
```

---

### Task 3: Refaktor — extrahera SM-2-uppdatering och exportera stjärnmappning

**Files:**
- Create: `backend/internal/chat/srs_update.go`
- Modify: `backend/internal/chat/handler.go:419-458` (EndSession-goroutinen)
- Modify: `backend/internal/challenges/handler.go:431-440` (`scoreToStars` → `ScoreToStars`)
- Test: `backend/internal/chat/srs_update_test.go`

**Interfaces:**
- Produces: `chat.UpdateReviewSchedule(ctx context.Context, store ChatStore, studentID, exerciseID pgtype.UUID, score int, now time.Time) error` och `challenges.ScoreToStars(score int) int` (1–3 stjärnor). Båda används av Task 7.

- [ ] **Step 1: Skriv failande test**

`backend/internal/chat/srs_update_test.go`:

```go
package chat

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

type fakeSRSStore struct {
	ChatStore // panics on anything else
	existing  *queries.ReviewSchedule
	upserted  *queries.UpsertReviewScheduleParams
}

func (f *fakeSRSStore) GetReviewSchedule(_ context.Context, _ queries.GetReviewScheduleParams) (queries.ReviewSchedule, error) {
	if f.existing == nil {
		return queries.ReviewSchedule{}, errors.New("not found")
	}
	return *f.existing, nil
}

func (f *fakeSRSStore) UpsertReviewSchedule(_ context.Context, arg queries.UpsertReviewScheduleParams) (queries.ReviewSchedule, error) {
	f.upserted = &arg
	return queries.ReviewSchedule{}, nil
}

func TestUpdateReviewScheduleFirstReview(t *testing.T) {
	store := &fakeSRSStore{}
	now := time.Date(2026, 7, 4, 12, 0, 0, 0, time.UTC)
	var student, exercise pgtype.UUID
	student.Scan("11111111-1111-1111-1111-111111111111")
	exercise.Scan("22222222-2222-2222-2222-222222222222")

	if err := UpdateReviewSchedule(context.Background(), store, student, exercise, 4, now); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.upserted == nil {
		t.Fatal("expected UpsertReviewSchedule to be called")
	}
	if store.upserted.RepetitionCount < 1 {
		t.Errorf("repetition count = %d, want >= 1", store.upserted.RepetitionCount)
	}
	if !store.upserted.NextReview.Time.After(now) {
		t.Errorf("next review %v not after now %v", store.upserted.NextReview.Time, now)
	}
}
```

- [ ] **Step 2: Verifiera FAIL** — `go test ./internal/chat/ -run TestUpdateReviewSchedule -v` → `undefined: UpdateReviewSchedule`

- [ ] **Step 3: Implementera**

`backend/internal/chat/srs_update.go`:

```go
package chat

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
	"github.com/trollstaven/nioplugget/backend/internal/srs"
)

// UpdateReviewSchedule recalculates the SM-2 review schedule after a scored
// session. Shared by the HTTP EndSession handler and the Telegram bot.
func UpdateReviewSchedule(ctx context.Context, store ChatStore, studentID, exerciseID pgtype.UUID, score int, now time.Time) error {
	input := srs.SM2Input{
		Score:           score,
		EaseFactor:      2.5, // SM-2 default
		IntervalDays:    1,
		RepetitionCount: 0,
	}
	existing, err := store.GetReviewSchedule(ctx, queries.GetReviewScheduleParams{
		StudentID:  studentID,
		ExerciseID: exerciseID,
	})
	if err == nil {
		input.EaseFactor = float64(existing.EaseFactor)
		input.IntervalDays = int(existing.IntervalDays)
		input.RepetitionCount = int(existing.RepetitionCount)
	}
	output := srs.Calculate(input, now)
	_, err = store.UpsertReviewSchedule(ctx, queries.UpsertReviewScheduleParams{
		StudentID:       studentID,
		ExerciseID:      exerciseID,
		EaseFactor:      float32(output.EaseFactor),
		IntervalDays:    int32(output.IntervalDays),
		RepetitionCount: int32(output.RepetitionCount),
		NextReview:      pgtype.Timestamptz{Time: output.NextReview, Valid: true},
	})
	return err
}
```

Byt sedan ut goroutine-kroppen i `EndSession` (`internal/chat/handler.go:425-457`) mot:

```go
		go func() {
			if err := UpdateReviewSchedule(srsCtx, h.store, srsStudentUUID, srsExerciseID, srsScore, time.Now()); err != nil {
				log.Error().Err(err).Msg("failed to update review schedule")
			}
		}()
```

(Behåll raderna 420-424 med `srsCtx`/`srsStudentUUID`/`srsExerciseID`/`srsScore` — de behövs fortfarande. Ta bort `srs.`-importen ur handler.go om den blir oanvänd.)

Byt namn `scoreToStars` → `ScoreToStars` i `internal/challenges/handler.go` (definition + alla anrop; verifiera med `grep -rn scoreToStars internal/`).

- [ ] **Step 4: Verifiera grönt** — `go test ./... ` → alla paket PASS (befintliga tester får inte gå sönder), `go vet ./...` rent.

- [ ] **Step 5: Commit**

```bash
git add internal/chat/ internal/challenges/
git commit -m "refactor: extract chat.UpdateReviewSchedule, export challenges.ScoreToStars"
```

---

### Task 4: Minimal Telegram Bot API-klient + meddelandedelning

**Files:**
- Create: `backend/internal/telegram/api.go`
- Create: `backend/internal/telegram/split.go`
- Test: `backend/internal/telegram/api_test.go`, `backend/internal/telegram/split_test.go`

**Interfaces:**
- Produces: typerna `Update`, `Message`, `CallbackQuery`, `Chat`, `User`, `InlineKeyboardMarkup`, `InlineKeyboardButton`; `NewAPI(token string) *API` med metoderna `GetUpdates(ctx, offset int64, timeoutSec int) ([]Update, error)`, `SendMessage(ctx, chatID int64, text string, keyboard *InlineKeyboardMarkup) error`, `SendChatAction(ctx, chatID int64, action string) error`, `AnswerCallbackQuery(ctx, id string) error`; `splitMessage(text string, max int) []string`; konstanten `telegramMessageLimit = 4096`.

- [ ] **Step 1: Skriv failande tester**

`backend/internal/telegram/split_test.go`:

```go
package telegram

import (
	"strings"
	"testing"
)

func TestSplitMessageShort(t *testing.T) {
	parts := splitMessage("hej", 4096)
	if len(parts) != 1 || parts[0] != "hej" {
		t.Errorf("got %v", parts)
	}
}

func TestSplitMessageLongBreaksAtNewline(t *testing.T) {
	text := strings.Repeat("a", 50) + "\n" + strings.Repeat("b", 30)
	parts := splitMessage(text, 60)
	if len(parts) != 2 {
		t.Fatalf("got %d parts, want 2: %v", len(parts), parts)
	}
	if parts[0] != strings.Repeat("a", 50)+"\n" {
		t.Errorf("part 0 = %q", parts[0])
	}
	for _, p := range parts {
		if len([]rune(p)) > 60 {
			t.Errorf("part exceeds max: %d", len([]rune(p)))
		}
	}
}

func TestSplitMessageNoBreakpoint(t *testing.T) {
	text := strings.Repeat("x", 130)
	parts := splitMessage(text, 60)
	if len(parts) != 3 {
		t.Fatalf("got %d parts, want 3", len(parts))
	}
}
```

`backend/internal/telegram/api_test.go`:

```go
package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPISendMessageAndError(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer srv.Close()

	api := &API{baseURL: srv.URL + "/botTEST", client: srv.Client()}
	if err := api.SendMessage(context.Background(), 42, "hej", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/botTEST/sendMessage" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["chat_id"].(float64) != 42 || gotBody["text"] != "hej" {
		t.Errorf("body = %v", gotBody)
	}

	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":false,"description":"Bad Request"}`))
	}))
	defer errSrv.Close()
	badAPI := &API{baseURL: errSrv.URL, client: errSrv.Client()}
	if err := badAPI.SendMessage(context.Background(), 42, "hej", nil); err == nil {
		t.Error("expected error on ok=false")
	}
}
```

- [ ] **Step 2: Verifiera FAIL** — `go test ./internal/telegram/ -v` → kompileringsfel (paketet finns inte ännu).

- [ ] **Step 3: Implementera**

`backend/internal/telegram/api.go`:

```go
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Minimal subset of the Telegram Bot API (https://core.telegram.org/bots/api).

type User struct {
	ID int64 `json:"id"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type Message struct {
	MessageID int64  `json:"message_id"`
	From      *User  `json:"from"`
	Chat      Chat   `json:"chat"`
	Text      string `json:"text"`
}

type CallbackQuery struct {
	ID      string   `json:"id"`
	From    User     `json:"from"`
	Message *Message `json:"message"`
	Data    string   `json:"data"`
}

type Update struct {
	UpdateID      int64          `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

// API is a minimal Telegram Bot API client.
type API struct {
	baseURL string
	client  *http.Client
}

func NewAPI(token string) *API {
	return &API{
		baseURL: "https://api.telegram.org/bot" + token,
		// Long polling holds the connection up to 50 s; timeout must exceed it.
		client: &http.Client{Timeout: 65 * time.Second},
	}
}

type apiResponse struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

func (a *API) call(ctx context.Context, method string, payload any, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/"+method, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("telegram %s: decode: %w", method, err)
	}
	if !apiResp.OK {
		return fmt.Errorf("telegram %s: %s", method, apiResp.Description)
	}
	if result != nil {
		return json.Unmarshal(apiResp.Result, result)
	}
	return nil
}

func (a *API) GetUpdates(ctx context.Context, offset int64, timeoutSec int) ([]Update, error) {
	var updates []Update
	err := a.call(ctx, "getUpdates", map[string]any{"offset": offset, "timeout": timeoutSec}, &updates)
	return updates, err
}

func (a *API) SendMessage(ctx context.Context, chatID int64, text string, keyboard *InlineKeyboardMarkup) error {
	payload := map[string]any{"chat_id": chatID, "text": text}
	if keyboard != nil {
		payload["reply_markup"] = keyboard
	}
	return a.call(ctx, "sendMessage", payload, nil)
}

func (a *API) SendChatAction(ctx context.Context, chatID int64, action string) error {
	return a.call(ctx, "sendChatAction", map[string]any{"chat_id": chatID, "action": action}, nil)
}

func (a *API) AnswerCallbackQuery(ctx context.Context, id string) error {
	return a.call(ctx, "answerCallbackQuery", map[string]any{"callback_query_id": id}, nil)
}
```

`backend/internal/telegram/split.go`:

```go
package telegram

// telegramMessageLimit is Telegram's hard cap on message text length.
const telegramMessageLimit = 4096

// splitMessage splits text into chunks of at most max runes, preferring to
// break at a newline, then a space, in the last half of the chunk.
func splitMessage(text string, max int) []string {
	runes := []rune(text)
	if len(runes) <= max {
		return []string{text}
	}
	var parts []string
	for len(runes) > max {
		cut := 0
		for i := max; i > max/2; i-- {
			if runes[i-1] == '\n' {
				cut = i
				break
			}
		}
		if cut == 0 {
			for i := max; i > max/2; i-- {
				if runes[i-1] == ' ' {
					cut = i
					break
				}
			}
		}
		if cut == 0 {
			cut = max
		}
		parts = append(parts, string(runes[:cut]))
		runes = runes[cut:]
	}
	if len(runes) > 0 {
		parts = append(parts, string(runes))
	}
	return parts
}
```

- [ ] **Step 4: Verifiera grönt** — `go test ./internal/telegram/ -v` → PASS

- [ ] **Step 5: Commit**

```bash
git add internal/telegram/
git commit -m "feat(telegram): minimal Bot API client and message splitting"
```

---

### Task 5: Store-interface + länkkod-endpoint

**Files:**
- Create: `backend/internal/telegram/store.go`
- Create: `backend/internal/telegram/util.go`
- Create: `backend/internal/telegram/link.go`
- Test: `backend/internal/telegram/link_test.go`

**Interfaces:**
- Consumes: `chat.ChatStore`, sqlc-metoderna från Task 1, `auth.GetUserIDFromContext`.
- Produces: `type Store interface` (se nedan), `NewQueriesStore(q *queries.Queries) *QueriesStore`, `NewLinkHandler(store Store, botUsername string) *LinkHandler` med `CreateLinkCode(w, r)` (POST, child-only, svar `{"code":"...","link":"https://t.me/<bot>?start=<kod>"}`), helpers `parseUUID`/`uuidToString`, `generateLinkCode() (string, error)`, konstant `linkCodeTTL = 15 * time.Minute`.

- [ ] **Step 1: Skriv `store.go` och `util.go`** (rena deklarationer — testas via senare tasks)

`backend/internal/telegram/util.go` — kopiera mönstret från `internal/chat/handler.go:494-509`:

```go
package telegram

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	b := u.Bytes
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func parseUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}
	}
	return u
}
```

`backend/internal/telegram/store.go`:

```go
package telegram

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/chat"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// Store is the database access the Telegram bot needs.
type Store interface {
	chat.ChatStore

	CreateTelegramLinkCode(ctx context.Context, arg queries.CreateTelegramLinkCodeParams) (queries.TelegramLinkCode, error)
	GetTelegramLinkCode(ctx context.Context, code string) (queries.TelegramLinkCode, error)
	MarkTelegramLinkCodeUsed(ctx context.Context, code string) error
	UpsertTelegramLink(ctx context.Context, arg queries.UpsertTelegramLinkParams) (queries.TelegramLink, error)
	GetTelegramLinkByTelegramUserID(ctx context.Context, telegramUserID int64) (queries.TelegramLink, error)
	ListTelegramLinks(ctx context.Context) ([]queries.TelegramLink, error)
	GetTelegramSession(ctx context.Context, chatID int64) (queries.TelegramSession, error)
	UpsertTelegramSession(ctx context.Context, arg queries.UpsertTelegramSessionParams) (queries.TelegramSession, error)
	TryInsertTelegramReminder(ctx context.Context, arg queries.TryInsertTelegramReminderParams) (int64, error)

	ListSubjects(ctx context.Context) ([]queries.ListSubjectsRow, error)
	GetSubjectBySlug(ctx context.Context, slug string) (queries.GetSubjectBySlugRow, error)
	ListTopicsBySubjectID(ctx context.Context, subjectID pgtype.UUID) ([]queries.ListTopicsBySubjectIDRow, error)
	GetTopicBySlug(ctx context.Context, arg queries.GetTopicBySlugParams) (queries.GetTopicBySlugRow, error)
	ListExercisesByTopicID(ctx context.Context, topicID pgtype.UUID) ([]queries.ListExercisesByTopicIDRow, error)
	ListPublishedChallengesByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.ListPublishedChallengesByParentIDRow, error)
	ListChallengeExercisesWithProgress(ctx context.Context, arg queries.ListChallengeExercisesWithProgressParams) ([]queries.ListChallengeExercisesWithProgressRow, error)
	ListDueReviews(ctx context.Context, studentID pgtype.UUID) ([]queries.ListDueReviewsRow, error)
	GetStudentProgressBySubject(ctx context.Context, studentID pgtype.UUID) ([]queries.GetStudentProgressBySubjectRow, error)
}

// QueriesStore implements Store using sqlc-generated queries.
// chat.QueriesStore provides the chat.ChatStore subset.
type QueriesStore struct {
	*chat.QueriesStore
	q *queries.Queries
}

func NewQueriesStore(q *queries.Queries) *QueriesStore {
	return &QueriesStore{QueriesStore: chat.NewQueriesStore(q), q: q}
}

func (s *QueriesStore) CreateTelegramLinkCode(ctx context.Context, arg queries.CreateTelegramLinkCodeParams) (queries.TelegramLinkCode, error) {
	return s.q.CreateTelegramLinkCode(ctx, arg)
}

func (s *QueriesStore) GetTelegramLinkCode(ctx context.Context, code string) (queries.TelegramLinkCode, error) {
	return s.q.GetTelegramLinkCode(ctx, code)
}

func (s *QueriesStore) MarkTelegramLinkCodeUsed(ctx context.Context, code string) error {
	return s.q.MarkTelegramLinkCodeUsed(ctx, code)
}

func (s *QueriesStore) UpsertTelegramLink(ctx context.Context, arg queries.UpsertTelegramLinkParams) (queries.TelegramLink, error) {
	return s.q.UpsertTelegramLink(ctx, arg)
}

func (s *QueriesStore) GetTelegramLinkByTelegramUserID(ctx context.Context, telegramUserID int64) (queries.TelegramLink, error) {
	return s.q.GetTelegramLinkByTelegramUserID(ctx, telegramUserID)
}

func (s *QueriesStore) ListTelegramLinks(ctx context.Context) ([]queries.TelegramLink, error) {
	return s.q.ListTelegramLinks(ctx)
}

func (s *QueriesStore) GetTelegramSession(ctx context.Context, chatID int64) (queries.TelegramSession, error) {
	return s.q.GetTelegramSession(ctx, chatID)
}

func (s *QueriesStore) UpsertTelegramSession(ctx context.Context, arg queries.UpsertTelegramSessionParams) (queries.TelegramSession, error) {
	return s.q.UpsertTelegramSession(ctx, arg)
}

func (s *QueriesStore) TryInsertTelegramReminder(ctx context.Context, arg queries.TryInsertTelegramReminderParams) (int64, error) {
	return s.q.TryInsertTelegramReminder(ctx, arg)
}

func (s *QueriesStore) ListSubjects(ctx context.Context) ([]queries.ListSubjectsRow, error) {
	return s.q.ListSubjects(ctx)
}

func (s *QueriesStore) GetSubjectBySlug(ctx context.Context, slug string) (queries.GetSubjectBySlugRow, error) {
	return s.q.GetSubjectBySlug(ctx, slug)
}

func (s *QueriesStore) ListTopicsBySubjectID(ctx context.Context, subjectID pgtype.UUID) ([]queries.ListTopicsBySubjectIDRow, error) {
	return s.q.ListTopicsBySubjectID(ctx, subjectID)
}

func (s *QueriesStore) GetTopicBySlug(ctx context.Context, arg queries.GetTopicBySlugParams) (queries.GetTopicBySlugRow, error) {
	return s.q.GetTopicBySlug(ctx, arg)
}

func (s *QueriesStore) ListExercisesByTopicID(ctx context.Context, topicID pgtype.UUID) ([]queries.ListExercisesByTopicIDRow, error) {
	return s.q.ListExercisesByTopicID(ctx, topicID)
}

func (s *QueriesStore) ListPublishedChallengesByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.ListPublishedChallengesByParentIDRow, error) {
	return s.q.ListPublishedChallengesByParentID(ctx, parentID)
}

func (s *QueriesStore) ListChallengeExercisesWithProgress(ctx context.Context, arg queries.ListChallengeExercisesWithProgressParams) ([]queries.ListChallengeExercisesWithProgressRow, error) {
	return s.q.ListChallengeExercisesWithProgress(ctx, arg)
}

func (s *QueriesStore) ListDueReviews(ctx context.Context, studentID pgtype.UUID) ([]queries.ListDueReviewsRow, error) {
	return s.q.ListDueReviews(ctx, studentID)
}

func (s *QueriesStore) GetStudentProgressBySubject(ctx context.Context, studentID pgtype.UUID) ([]queries.GetStudentProgressBySubjectRow, error) {
	return s.q.GetStudentProgressBySubject(ctx, studentID)
}
```

Kontroll: fältnamnen i genererade paramtyper (`queries.UpsertTelegramLinkParams` osv.) — verifiera mot `internal/database/queries/telegram.sql.go` och justera vid avvikelse.

- [ ] **Step 2: Skriv failande test för länkkoden**

`backend/internal/telegram/link_test.go`:

```go
package telegram

import (
	"context"
	"strings"
	"testing"

	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

func TestGenerateLinkCode(t *testing.T) {
	code, err := generateLinkCode()
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 8 {
		t.Errorf("len = %d, want 8", len(code))
	}
	for _, c := range code {
		if !strings.ContainsRune(codeAlphabet, c) {
			t.Errorf("invalid char %q", c)
		}
	}
	code2, _ := generateLinkCode()
	if code == code2 {
		t.Error("two codes should differ")
	}
}

type fakeLinkStore struct {
	Store
	created *queries.CreateTelegramLinkCodeParams
}

func (f *fakeLinkStore) CreateTelegramLinkCode(_ context.Context, arg queries.CreateTelegramLinkCodeParams) (queries.TelegramLinkCode, error) {
	f.created = &arg
	return queries.TelegramLinkCode{Code: arg.Code, StudentID: arg.StudentID, ExpiresAt: arg.ExpiresAt}, nil
}
```

(HTTP-handlern testas via httptest i Step 4 — auth-context sätts med samma mekanism som `auth.GetUserIDFromContext` läser; kolla `internal/auth` efter test-hjälpare eller context-nyckel och injicera studentID i request-contexten. Om context-nyckeln är opriviligerad/oexporterad: testa `generateLinkCode` + storeanropet direkt genom att anropa handlern med en request vars context byggts via `auth`-paketets exporterade funktioner, annars flytta handler-logiken till en intern funktion `createLinkCode(ctx, store, botUsername, studentID) (code, link string, err error)` och testa den.)

- [ ] **Step 3: Implementera `link.go`**

```go
package telegram

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/auth"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

const linkCodeTTL = 15 * time.Minute

// codeAlphabet omits confusable characters (0/O, 1/I).
const codeAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateLinkCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = codeAlphabet[int(b[i])%len(codeAlphabet)]
	}
	return string(b), nil
}

// LinkHandler serves the web GUI's Telegram linking endpoint.
type LinkHandler struct {
	store       Store
	botUsername string
}

func NewLinkHandler(store Store, botUsername string) *LinkHandler {
	return &LinkHandler{store: store, botUsername: botUsername}
}

// CreateLinkCode handles POST /api/telegram/link-code (child only).
func (h *LinkHandler) CreateLinkCode(w http.ResponseWriter, r *http.Request) {
	studentID := auth.GetUserIDFromContext(r.Context())
	if studentID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}
	code, err := generateLinkCode()
	if err != nil {
		http.Error(w, `{"error":"Kunde inte skapa kod"}`, http.StatusInternalServerError)
		return
	}
	if _, err := h.store.CreateTelegramLinkCode(r.Context(), queries.CreateTelegramLinkCodeParams{
		Code:      code,
		StudentID: parseUUID(studentID),
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(linkCodeTTL), Valid: true},
	}); err != nil {
		log.Error().Err(err).Msg("failed to create telegram link code")
		http.Error(w, `{"error":"Kunde inte skapa kod"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"code": code,
		"link": "https://t.me/" + h.botUsername + "?start=" + code,
	})
}
```

- [ ] **Step 4: Verifiera grönt** — `go test ./internal/telegram/ -v && go build ./...` → PASS

- [ ] **Step 5: Commit**

```bash
git add internal/telegram/
git commit -m "feat(telegram): store interface and link-code endpoint"
```

---

### Task 6: Bot-kärna — koppling via /start, huvudmeny, polling-loop

**Files:**
- Create: `backend/internal/telegram/bot.go`
- Create: `backend/internal/telegram/router.go`
- Create: `backend/internal/telegram/menu.go`
- Test: `backend/internal/telegram/router_test.go`

**Interfaces:**
- Consumes: Task 4:s typer/klient, Task 5:s `Store`, `chat.CompleteChatResponse`, `chat.ScoreSession`, `apikey.EncryptionService`.
- Produces: `type Sender interface { SendMessage(...); SendChatAction(...); AnswerCallbackQuery(...) }` (samma signaturer som `*API`); `NewBot(api Sender, store Store, encSvc *apikey.EncryptionService) *Bot`; `(*Bot).HandleUpdate(ctx, Update)`; `Run(ctx, *API, *Bot)`; `(*Bot).sendMainMenu(ctx, queries.TelegramLink)`; callbackdata-schema: `menu`, `study`, `subj:<slug>`, `top:<subjectSlug>:<topicSlug>`, `ex:<uuid>`, `chals`, `chal:<uuid>`, `chex:<uuid>`, `rev`, `prog`, `end` (max 64 bytes — Telegrams gräns för callback_data).

- [ ] **Step 1: Skriv failande tester**

`backend/internal/telegram/router_test.go`:

```go
package telegram

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

type sentMessage struct {
	chatID int64
	text   string
	kb     *InlineKeyboardMarkup
}

type fakeSender struct {
	messages []sentMessage
	actions  []string
	answered []string
}

func (f *fakeSender) SendMessage(_ context.Context, chatID int64, text string, kb *InlineKeyboardMarkup) error {
	f.messages = append(f.messages, sentMessage{chatID, text, kb})
	return nil
}
func (f *fakeSender) SendChatAction(_ context.Context, _ int64, action string) error {
	f.actions = append(f.actions, action)
	return nil
}
func (f *fakeSender) AnswerCallbackQuery(_ context.Context, id string) error {
	f.answered = append(f.answered, id)
	return nil
}

// fakeStore: embedded Store gör att oimplementerade metoder panikar i test.
type fakeStore struct {
	Store
	linkCodes  map[string]queries.TelegramLinkCode
	links      map[int64]queries.TelegramLink // key: telegram_user_id
	tgSessions map[int64]queries.TelegramSession
	students   map[string]queries.Student // key: uuidToString(id)
	usedCodes  []string
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		linkCodes:  map[string]queries.TelegramLinkCode{},
		links:      map[int64]queries.TelegramLink{},
		tgSessions: map[int64]queries.TelegramSession{},
		students:   map[string]queries.Student{},
	}
}

func (f *fakeStore) GetTelegramLinkCode(_ context.Context, code string) (queries.TelegramLinkCode, error) {
	lc, ok := f.linkCodes[code]
	if !ok {
		return queries.TelegramLinkCode{}, errors.New("not found")
	}
	return lc, nil
}
func (f *fakeStore) MarkTelegramLinkCodeUsed(_ context.Context, code string) error {
	f.usedCodes = append(f.usedCodes, code)
	return nil
}
func (f *fakeStore) UpsertTelegramLink(_ context.Context, arg queries.UpsertTelegramLinkParams) (queries.TelegramLink, error) {
	link := queries.TelegramLink{StudentID: arg.StudentID, TelegramUserID: arg.TelegramUserID, ChatID: arg.ChatID}
	f.links[arg.TelegramUserID] = link
	return link, nil
}
func (f *fakeStore) GetTelegramLinkByTelegramUserID(_ context.Context, id int64) (queries.TelegramLink, error) {
	link, ok := f.links[id]
	if !ok {
		return queries.TelegramLink{}, errors.New("not found")
	}
	return link, nil
}
func (f *fakeStore) GetTelegramSession(_ context.Context, chatID int64) (queries.TelegramSession, error) {
	s, ok := f.tgSessions[chatID]
	if !ok {
		return queries.TelegramSession{}, errors.New("not found")
	}
	return s, nil
}
func (f *fakeStore) UpsertTelegramSession(_ context.Context, arg queries.UpsertTelegramSessionParams) (queries.TelegramSession, error) {
	s := queries.TelegramSession{ChatID: arg.ChatID, StudentID: arg.StudentID, State: arg.State, ActiveSessionID: arg.ActiveSessionID}
	f.tgSessions[arg.ChatID] = s
	return s, nil
}
func (f *fakeStore) GetStudentByID(_ context.Context, id pgtype.UUID) (queries.Student, error) {
	s, ok := f.students[uuidToString(id)]
	if !ok {
		return queries.Student{}, errors.New("not found")
	}
	return s, nil
}
func (f *fakeStore) ListDueReviews(_ context.Context, _ pgtype.UUID) ([]queries.ListDueReviewsRow, error) {
	return nil, nil
}

func testUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	u.Scan(s)
	return u
}

func TestStartWithValidCodeLinksAccount(t *testing.T) {
	store := newFakeStore()
	studentID := testUUID("11111111-1111-1111-1111-111111111111")
	store.students[uuidToString(studentID)] = queries.Student{ID: studentID, Name: "Nio"}
	store.linkCodes["ABCD2345"] = queries.TelegramLinkCode{
		Code:      "ABCD2345",
		StudentID: studentID,
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(10 * time.Minute), Valid: true},
	}
	sender := &fakeSender{}
	bot := NewBot(sender, store, nil)

	bot.HandleUpdate(context.Background(), Update{Message: &Message{
		From: &User{ID: 777}, Chat: Chat{ID: 555}, Text: "/start ABCD2345",
	}})

	if _, ok := store.links[777]; !ok {
		t.Fatal("expected link to be created")
	}
	if len(store.usedCodes) != 1 || store.usedCodes[0] != "ABCD2345" {
		t.Errorf("code not marked used: %v", store.usedCodes)
	}
	if len(sender.messages) < 2 {
		t.Fatalf("expected welcome + menu, got %d messages", len(sender.messages))
	}
	if !strings.Contains(sender.messages[0].text, "Nio") {
		t.Errorf("welcome should greet by name: %q", sender.messages[0].text)
	}
	menu := sender.messages[len(sender.messages)-1]
	if menu.kb == nil || len(menu.kb.InlineKeyboard) != 4 {
		t.Errorf("expected 4-row main menu keyboard, got %+v", menu.kb)
	}
}

func TestStartWithExpiredCodeRejects(t *testing.T) {
	store := newFakeStore()
	store.linkCodes["OLDCODE2"] = queries.TelegramLinkCode{
		Code:      "OLDCODE2",
		StudentID: testUUID("11111111-1111-1111-1111-111111111111"),
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(-time.Minute), Valid: true},
	}
	sender := &fakeSender{}
	bot := NewBot(sender, store, nil)

	bot.HandleUpdate(context.Background(), Update{Message: &Message{
		From: &User{ID: 777}, Chat: Chat{ID: 555}, Text: "/start OLDCODE2",
	}})

	if len(store.links) != 0 {
		t.Error("expired code must not link")
	}
	if len(sender.messages) != 1 || !strings.Contains(sender.messages[0].text, "ogiltig") {
		t.Errorf("expected rejection, got %v", sender.messages)
	}
}

func TestUnlinkedUserIsRejected(t *testing.T) {
	sender := &fakeSender{}
	bot := NewBot(sender, newFakeStore(), nil)

	bot.HandleUpdate(context.Background(), Update{Message: &Message{
		From: &User{ID: 999}, Chat: Chat{ID: 888}, Text: "hej boten",
	}})

	if len(sender.messages) != 1 || !strings.Contains(sender.messages[0].text, "kopplingslänk") {
		t.Errorf("expected unlinked reply, got %v", sender.messages)
	}
}

func TestCallbackIsAnswered(t *testing.T) {
	store := newFakeStore()
	studentID := testUUID("11111111-1111-1111-1111-111111111111")
	store.links[777] = queries.TelegramLink{StudentID: studentID, TelegramUserID: 777, ChatID: 555}
	sender := &fakeSender{}
	bot := NewBot(sender, store, nil)

	bot.HandleUpdate(context.Background(), Update{CallbackQuery: &CallbackQuery{
		ID: "cb1", From: User{ID: 777}, Message: &Message{Chat: Chat{ID: 555}}, Data: "menu",
	}})

	if len(sender.answered) != 1 || sender.answered[0] != "cb1" {
		t.Errorf("callback not answered: %v", sender.answered)
	}
	if len(sender.messages) != 1 || sender.messages[0].kb == nil {
		t.Errorf("expected main menu, got %v", sender.messages)
	}
}
```

- [ ] **Step 2: Verifiera FAIL** — `go test ./internal/telegram/ -v` → `undefined: NewBot` m.fl.

- [ ] **Step 3: Implementera `bot.go`**

```go
package telegram

import (
	"context"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/apikey"
	"github.com/trollstaven/nioplugget/backend/internal/chat"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// Sender is the subset of the Telegram API the bot sends through (mockable).
type Sender interface {
	SendMessage(ctx context.Context, chatID int64, text string, keyboard *InlineKeyboardMarkup) error
	SendChatAction(ctx context.Context, chatID int64, action string) error
	AnswerCallbackQuery(ctx context.Context, id string) error
}

// CompleteFunc produces a full AI reply (injectable for tests).
type CompleteFunc func(ctx context.Context, apiKey, systemPrompt string, messages []anthropic.MessageParam) (string, error)

// ScoreFunc scores a finished session (injectable for tests).
type ScoreFunc func(ctx context.Context, apiKey, exerciseTitle string, messages []queries.Message) (*chat.ScoreResult, error)

type Bot struct {
	api      Sender
	store    Store
	encSvc   *apikey.EncryptionService
	complete CompleteFunc
	score    ScoreFunc
}

func NewBot(api Sender, store Store, encSvc *apikey.EncryptionService) *Bot {
	return &Bot{
		api:    api,
		store:  store,
		encSvc: encSvc,
		complete: func(ctx context.Context, apiKey, systemPrompt string, messages []anthropic.MessageParam) (string, error) {
			return chat.CompleteChatResponse(ctx, apiKey, systemPrompt, messages)
		},
		score: chat.ScoreSession,
	}
}

// Run long-polls Telegram until ctx is canceled.
func Run(ctx context.Context, api *API, bot *Bot) {
	var offset int64
	backoff := time.Second
	for ctx.Err() == nil {
		updates, err := api.GetUpdates(ctx, offset, 50)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Error().Err(err).Msg("telegram getUpdates failed")
			time.Sleep(backoff)
			if backoff < time.Minute {
				backoff *= 2
			}
			continue
		}
		backoff = time.Second
		for _, upd := range updates {
			offset = upd.UpdateID + 1
			bot.HandleUpdate(ctx, upd)
		}
	}
}

// send splits long texts and attaches the keyboard to the last part.
func (b *Bot) send(ctx context.Context, chatID int64, text string, kb *InlineKeyboardMarkup) {
	parts := splitMessage(text, telegramMessageLimit)
	for i, part := range parts {
		var partKB *InlineKeyboardMarkup
		if i == len(parts)-1 {
			partKB = kb
		}
		if err := b.api.SendMessage(ctx, chatID, part, partKB); err != nil {
			log.Error().Err(err).Int64("chat_id", chatID).Msg("telegram send failed")
		}
	}
}
```

- [ ] **Step 4: Implementera `router.go`**

```go
package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

const unlinkedReply = "Hej! Jag är Nioplugget-boten. Hämta en kopplingslänk på nioplugget-webben (knappen \"Koppla Telegram\") för att komma igång."

// HandleUpdate routes a single Telegram update. Errors are logged, never fatal.
func (b *Bot) HandleUpdate(ctx context.Context, upd Update) {
	switch {
	case upd.CallbackQuery != nil:
		b.handleCallback(ctx, *upd.CallbackQuery)
	case upd.Message != nil && upd.Message.From != nil && upd.Message.Text != "":
		b.handleMessage(ctx, *upd.Message)
	}
}

func (b *Bot) handleMessage(ctx context.Context, msg Message) {
	if strings.HasPrefix(msg.Text, "/start") {
		b.handleStart(ctx, msg)
		return
	}
	link, err := b.store.GetTelegramLinkByTelegramUserID(ctx, msg.From.ID)
	if err != nil {
		b.send(ctx, msg.Chat.ID, unlinkedReply, nil)
		return
	}
	if msg.Text == "/avsluta" {
		b.endActiveSession(ctx, link)
		return
	}
	tgSession, err := b.store.GetTelegramSession(ctx, link.ChatID)
	if err == nil && tgSession.State == "in_session" && tgSession.ActiveSessionID.Valid {
		b.handleDialogMessage(ctx, link, tgSession, msg.Text)
		return
	}
	b.sendMainMenu(ctx, link)
}

func (b *Bot) handleStart(ctx context.Context, msg Message) {
	chatID := msg.Chat.ID
	parts := strings.Fields(msg.Text)
	if len(parts) == 2 {
		code := strings.ToUpper(parts[1])
		lc, err := b.store.GetTelegramLinkCode(ctx, code)
		if err != nil || lc.UsedAt.Valid || time.Now().After(lc.ExpiresAt.Time) {
			b.send(ctx, chatID, "Koden är ogiltig eller har gått ut. Hämta en ny länk från webben.", nil)
			return
		}
		link, err := b.store.UpsertTelegramLink(ctx, queries.UpsertTelegramLinkParams{
			StudentID:      lc.StudentID,
			TelegramUserID: msg.From.ID,
			ChatID:         chatID,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to link telegram account")
			b.send(ctx, chatID, "Något gick fel vid kopplingen. Försök igen.", nil)
			return
		}
		if err := b.store.MarkTelegramLinkCodeUsed(ctx, code); err != nil {
			log.Error().Err(err).Msg("failed to mark link code used")
		}
		name := ""
		if student, err := b.store.GetStudentByID(ctx, link.StudentID); err == nil {
			name = " " + student.Name
		}
		b.send(ctx, chatID, fmt.Sprintf("Kopplat! 🎉 Hej%s, nu kan du plugga direkt här i Telegram.", name), nil)
		b.sendMainMenu(ctx, link)
		return
	}
	link, err := b.store.GetTelegramLinkByTelegramUserID(ctx, msg.From.ID)
	if err != nil {
		b.send(ctx, chatID, unlinkedReply, nil)
		return
	}
	b.sendMainMenu(ctx, link)
}

func (b *Bot) handleCallback(ctx context.Context, cb CallbackQuery) {
	if err := b.api.AnswerCallbackQuery(ctx, cb.ID); err != nil {
		log.Error().Err(err).Msg("failed to answer callback query")
	}
	if cb.Message == nil {
		return
	}
	chatID := cb.Message.Chat.ID
	link, err := b.store.GetTelegramLinkByTelegramUserID(ctx, cb.From.ID)
	if err != nil {
		b.send(ctx, chatID, unlinkedReply, nil)
		return
	}
	data := cb.Data
	switch {
	case data == "menu":
		b.sendMainMenu(ctx, link)
	case data == "study":
		b.sendSubjects(ctx, link)
	case strings.HasPrefix(data, "subj:"):
		b.sendTopics(ctx, link, strings.TrimPrefix(data, "subj:"))
	case strings.HasPrefix(data, "top:"):
		if segs := strings.SplitN(strings.TrimPrefix(data, "top:"), ":", 2); len(segs) == 2 {
			b.sendExercises(ctx, link, segs[0], segs[1])
		}
	case strings.HasPrefix(data, "ex:"):
		b.startExercise(ctx, link, strings.TrimPrefix(data, "ex:"))
	case data == "chals":
		b.sendChallenges(ctx, link)
	case strings.HasPrefix(data, "chal:"):
		b.sendChallengeExercises(ctx, link, strings.TrimPrefix(data, "chal:"))
	case strings.HasPrefix(data, "chex:"):
		b.startChallengeExercise(ctx, link, strings.TrimPrefix(data, "chex:"))
	case data == "rev":
		b.sendDueReviews(ctx, link)
	case data == "prog":
		b.sendProgress(ctx, link)
	case data == "end":
		b.endActiveSession(ctx, link)
	}
}
```

- [ ] **Step 5: Implementera `menu.go` med huvudmenyn + tomma stubbar**

För att Task 6 ska kompilera och testerna gå gröna: `sendMainMenu` komplett, övriga meny-/dialogfunktioner som stubbar som fylls i Task 7 (`sendSubjects`, `sendTopics`, `sendExercises`, `sendChallenges`, `sendChallengeExercises`, `sendDueReviews`, `sendProgress`, `startExercise`, `startChallengeExercise`, `endActiveSession`, `handleDialogMessage`). Stubbar skickar "Kommer strax!" så att inget anrop panikar:

```go
package telegram

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

func (b *Bot) sendMainMenu(ctx context.Context, link queries.TelegramLink) {
	due, err := b.store.ListDueReviews(ctx, link.StudentID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list due reviews for menu")
	}
	kb := &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
		{{Text: "📚 Plugga", CallbackData: "study"}},
		{{Text: fmt.Sprintf("⭐ Repetera (%d)", len(due)), CallbackData: "rev"}},
		{{Text: "🏆 Utmaningar", CallbackData: "chals"}},
		{{Text: "📊 Min progress", CallbackData: "prog"}},
	}}
	b.send(ctx, link.ChatID, "Vad vill du göra?", kb)
}

// Stubs — implemented in the browse/dialog task.
func (b *Bot) sendSubjects(ctx context.Context, link queries.TelegramLink) {
	b.send(ctx, link.ChatID, "Kommer strax!", nil)
}
```

(…samma mönster för övriga stubbar; `handleDialogMessage(ctx, link, tgSession, text)` och `endActiveSession(ctx, link)` tar parametrarna som router.go anropar med.)

- [ ] **Step 6: Verifiera grönt** — `go test ./internal/telegram/ -v && go build ./...` → PASS

- [ ] **Step 7: Commit**

```bash
git add internal/telegram/
git commit -m "feat(telegram): bot core with linking, main menu, polling loop"
```

---

### Task 7: Bläddring + dialog + avslut med rättning, stjärnor och SM-2

**Files:**
- Modify: `backend/internal/telegram/menu.go` (ersätt stubbarna)
- Create: `backend/internal/telegram/dialog.go`
- Test: `backend/internal/telegram/dialog_test.go`

**Interfaces:**
- Consumes: `chat.BuildMessageHistory`, `chat.DefaultWindowSize`, `chat.UpdateReviewSchedule` (Task 3), `challenges.ScoreToStars` (Task 3), `b.complete`/`b.score` (Task 6), Store-metoderna.
- Produces: färdiga `sendSubjects/sendTopics/sendExercises/sendChallenges/sendChallengeExercises/sendDueReviews/sendProgress/startExercise/startChallengeExercise/handleDialogMessage/endActiveSession` + `endKeyboard()` + konstant `telegramPromptSuffix`.

- [ ] **Step 1: Skriv failande tester**

`backend/internal/telegram/dialog_test.go` (utökar fakeStore från router_test.go med sessions/messages/övningar):

```go
package telegram

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/chat"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// dialogStore extends fakeStore with session/message/exercise state.
type dialogStore struct {
	*fakeStore
	sessions  map[string]queries.GetSessionByIDRow
	messages  []queries.Message
	exercises map[string]queries.GetExerciseByIDRow
	apiKeys   map[string]queries.ApiKey // key: uuidToString(parentID)
	ended     *queries.EndSessionParams
	schedule  *queries.UpsertReviewScheduleParams
}

func newDialogStore() *dialogStore {
	return &dialogStore{
		fakeStore: newFakeStore(),
		sessions:  map[string]queries.GetSessionByIDRow{},
		exercises: map[string]queries.GetExerciseByIDRow{},
		apiKeys:   map[string]queries.ApiKey{},
	}
}

func (d *dialogStore) CreateSession(_ context.Context, arg queries.CreateSessionParams) (queries.CreateSessionRow, error) {
	id := testUUID("33333333-3333-3333-3333-333333333333")
	d.sessions[uuidToString(id)] = queries.GetSessionByIDRow{ID: id, StudentID: arg.StudentID, ExerciseID: arg.ExerciseID}
	return queries.CreateSessionRow{ID: id, StudentID: arg.StudentID, ExerciseID: arg.ExerciseID}, nil
}
func (d *dialogStore) GetSessionByID(_ context.Context, id pgtype.UUID) (queries.GetSessionByIDRow, error) {
	s, ok := d.sessions[uuidToString(id)]
	if !ok {
		return queries.GetSessionByIDRow{}, errors.New("not found")
	}
	return s, nil
}
func (d *dialogStore) CreateMessage(_ context.Context, arg queries.CreateMessageParams) (queries.Message, error) {
	m := queries.Message{SessionID: arg.SessionID, Role: arg.Role, Content: arg.Content}
	d.messages = append(d.messages, m)
	return m, nil
}
func (d *dialogStore) ListMessagesBySessionID(_ context.Context, _ pgtype.UUID) ([]queries.Message, error) {
	return d.messages, nil
}
func (d *dialogStore) GetExerciseByID(_ context.Context, id pgtype.UUID) (queries.GetExerciseByIDRow, error) {
	e, ok := d.exercises[uuidToString(id)]
	if !ok {
		return queries.GetExerciseByIDRow{}, errors.New("not found")
	}
	return e, nil
}
func (d *dialogStore) GetAPIKeyByParentID(_ context.Context, parentID pgtype.UUID) (queries.ApiKey, error) {
	k, ok := d.apiKeys[uuidToString(parentID)]
	if !ok {
		return queries.ApiKey{}, errors.New("not found")
	}
	return k, nil
}
func (d *dialogStore) EndSession(_ context.Context, arg queries.EndSessionParams) (queries.EndSessionRow, error) {
	d.ended = &arg
	s := d.sessions[uuidToString(arg.ID)]
	s.EndedAt = pgtype.Timestamptz{Valid: true}
	d.sessions[uuidToString(arg.ID)] = s
	return queries.EndSessionRow{ID: arg.ID}, nil
}
func (d *dialogStore) GetReviewSchedule(_ context.Context, _ queries.GetReviewScheduleParams) (queries.ReviewSchedule, error) {
	return queries.ReviewSchedule{}, errors.New("not found")
}
func (d *dialogStore) UpsertReviewSchedule(_ context.Context, arg queries.UpsertReviewScheduleParams) (queries.ReviewSchedule, error) {
	d.schedule = &arg
	return queries.ReviewSchedule{}, nil
}

func setupLinkedStudent(d *dialogStore) queries.TelegramLink {
	studentID := testUUID("11111111-1111-1111-1111-111111111111")
	parentID := testUUID("44444444-4444-4444-4444-444444444444")
	d.students[uuidToString(studentID)] = queries.Student{ID: studentID, ParentID: parentID, Name: "Nio"}
	link := queries.TelegramLink{StudentID: studentID, TelegramUserID: 777, ChatID: 555}
	d.links[777] = link
	return link
}

func TestStartExerciseCreatesSessionAndState(t *testing.T) {
	d := newDialogStore()
	link := setupLinkedStudent(d)
	exID := testUUID("22222222-2222-2222-2222-222222222222")
	d.exercises[uuidToString(exID)] = queries.GetExerciseByIDRow{ID: exID, Title: "Fotosyntes", Description: "Förklara fotosyntesen", SystemPrompt: "Du är en sokratisk lärare."}
	sender := &fakeSender{}
	bot := NewBot(sender, d, nil)

	bot.startExercise(context.Background(), link, uuidToString(exID))

	tg := d.tgSessions[555]
	if tg.State != "in_session" || !tg.ActiveSessionID.Valid {
		t.Fatalf("telegram session not in_session: %+v", tg)
	}
	if len(sender.messages) == 0 || !strings.Contains(sender.messages[0].text, "Fotosyntes") {
		t.Errorf("intro should mention exercise title: %v", sender.messages)
	}
}

func TestDialogMessageCallsClaudeWithTelegramSuffix(t *testing.T) {
	d := newDialogStore()
	link := setupLinkedStudent(d)
	exID := testUUID("22222222-2222-2222-2222-222222222222")
	d.exercises[uuidToString(exID)] = queries.GetExerciseByIDRow{ID: exID, Title: "Fotosyntes", SystemPrompt: "Sokratisk lärare."}
	sessionID := testUUID("33333333-3333-3333-3333-333333333333")
	d.sessions[uuidToString(sessionID)] = queries.GetSessionByIDRow{ID: sessionID, StudentID: link.StudentID, ExerciseID: exID}
	// Fake encryptable API key path: bypass encryption by stubbing decryption
	// via a bot whose encSvc is replaced — instead we stub complete/score and
	// return the key straight from a fake encSvc-less path is not possible,
	// so store an encrypted key with a real EncryptionService:
	sender := &fakeSender{}
	bot := NewBot(sender, d, testEncSvc(t, d, link))

	var gotSystem string
	bot.complete = func(_ context.Context, _, systemPrompt string, _ []anthropic.MessageParam) (string, error) {
		gotSystem = systemPrompt
		return "Bra fråga! Vad tror du själv?", nil
	}

	tgSession := queries.TelegramSession{ChatID: 555, StudentID: link.StudentID, State: "in_session", ActiveSessionID: sessionID}
	bot.handleDialogMessage(context.Background(), link, tgSession, "Vad är klorofyll?")

	if !strings.Contains(gotSystem, "LaTeX") {
		t.Errorf("system prompt must include telegram suffix, got %q", gotSystem)
	}
	if len(d.messages) != 2 || d.messages[0].Role != "user" || d.messages[1].Role != "assistant" {
		t.Errorf("expected saved user+assistant messages, got %+v", d.messages)
	}
	if len(sender.actions) == 0 || sender.actions[0] != "typing" {
		t.Errorf("expected typing action, got %v", sender.actions)
	}
	last := sender.messages[len(sender.messages)-1]
	if !strings.Contains(last.text, "Vad tror du själv?") || last.kb == nil {
		t.Errorf("reply should carry end keyboard: %+v", last)
	}
}

func TestEndActiveSessionScoresAndSchedules(t *testing.T) {
	d := newDialogStore()
	link := setupLinkedStudent(d)
	exID := testUUID("22222222-2222-2222-2222-222222222222")
	d.exercises[uuidToString(exID)] = queries.GetExerciseByIDRow{ID: exID, Title: "Fotosyntes"}
	sessionID := testUUID("33333333-3333-3333-3333-333333333333")
	d.sessions[uuidToString(sessionID)] = queries.GetSessionByIDRow{ID: sessionID, StudentID: link.StudentID, ExerciseID: exID}
	d.tgSessions[555] = queries.TelegramSession{ChatID: 555, StudentID: link.StudentID, State: "in_session", ActiveSessionID: sessionID}
	sender := &fakeSender{}
	bot := NewBot(sender, d, testEncSvc(t, d, link))
	bot.score = func(_ context.Context, _, _ string, _ []queries.Message) (*chat.ScoreResult, error) {
		return &chat.ScoreResult{Score: 5, Summary: "Toppen!", Feedback: "Snyggt!"}, nil
	}

	bot.endActiveSession(context.Background(), link)

	if d.ended == nil || d.ended.Score.Int32 != 5 {
		t.Fatalf("session not ended with score: %+v", d.ended)
	}
	if d.schedule == nil {
		t.Fatal("SM-2 schedule not updated")
	}
	if d.tgSessions[555].State != "menu" {
		t.Errorf("telegram session should reset to menu, got %q", d.tgSessions[555].State)
	}
	var starsMsg string
	for _, m := range sender.messages {
		if strings.Contains(m.text, "⭐") {
			starsMsg = m.text
		}
	}
	if !strings.Contains(starsMsg, "⭐⭐⭐") || !strings.Contains(starsMsg, "+30 XP") {
		t.Errorf("expected 3 stars +30 XP, got %q", starsMsg)
	}
}
```

Lägg också till test-hjälparen (samma fil):

```go
// testEncSvc creates a real EncryptionService with a test key and stores an
// encrypted API key for the linked student's parent.
func testEncSvc(t *testing.T, d *dialogStore, link queries.TelegramLink) *apikey.EncryptionService {
	t.Helper()
	// OBS: kontrollera vilket nyckelformat apikey.NewEncryptionService kräver
	// (se internal/apikey/) — 32 byte, sannolikt hex eller base64 — och justera.
	svc, err := apikey.NewEncryptionService(strings.Repeat("a", 64))
	if err != nil {
		t.Fatalf("enc svc: %v", err)
	}
	student := d.students[uuidToString(link.StudentID)]
	enc, err := svc.Encrypt([]byte("sk-ant-test"))
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	d.apiKeys[uuidToString(student.ParentID)] = queries.ApiKey{ParentID: student.ParentID, EncryptedKey: enc}
	return svc
}
```

(Verifiera `Encrypt`-signaturen och `queries.ApiKey.EncryptedKey`-fältets typ i `internal/apikey/` och justera hjälparen — inte produktionskoden — vid behov. Importera `apikey`-paketet i testfilen.)

- [ ] **Step 2: Verifiera FAIL** — `go test ./internal/telegram/ -v` → udefinierade funktioner/fel beteende i stubbar.

- [ ] **Step 3: Implementera `dialog.go`**

```go
package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/chat"
	"github.com/trollstaven/nioplugget/backend/internal/challenges"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// telegramPromptSuffix is appended to every system prompt in the Telegram
// channel: Telegram cannot render LaTeX.
const telegramPromptSuffix = "\n\nVIKTIGT FÖR DENNA KANAL: Du chattar via Telegram. Skriv aldrig LaTeX. Skriv matematik som enkel text med unicode, t.ex. x² + 3x = 10, √16, ½."

func endKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
		{{Text: "✅ Klar — avsluta och rätta", CallbackData: "end"}},
	}}
}

func (b *Bot) startExercise(ctx context.Context, link queries.TelegramLink, exerciseID string) {
	exercise, err := b.store.GetExerciseByID(ctx, parseUUID(exerciseID))
	if err != nil {
		b.send(ctx, link.ChatID, "Övningen hittades inte.", nil)
		return
	}
	session, err := b.store.CreateSession(ctx, queries.CreateSessionParams{
		StudentID:  link.StudentID,
		ExerciseID: exercise.ID,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create session from telegram")
		b.send(ctx, link.ChatID, "Kunde inte starta övningen. Försök igen.", nil)
		return
	}
	b.setActiveSession(ctx, link, session.ID)
	b.send(ctx, link.ChatID, fmt.Sprintf("📚 %s\n\n%s\n\nSkriv ditt svar eller din fundering — jag guidar dig. Skriv /avsluta när du känner dig klar.", exercise.Title, exercise.Description), endKeyboard())
}

func (b *Bot) startChallengeExercise(ctx context.Context, link queries.TelegramLink, chexIDStr string) {
	chexID := parseUUID(chexIDStr)
	chex, err := b.store.GetChallengeExerciseByID(ctx, chexID)
	if err != nil {
		b.send(ctx, link.ChatID, "Övningen hittades inte.", nil)
		return
	}
	challenge, err := b.store.GetChallengeByID(ctx, chex.ChallengeID)
	if err != nil {
		b.send(ctx, link.ChatID, "Övningen hittades inte.", nil)
		return
	}
	student, err := b.store.GetStudentByID(ctx, link.StudentID)
	if err != nil || challenge.ParentID != student.ParentID || !challenge.Published {
		b.send(ctx, link.ChatID, "Övningen hittades inte.", nil)
		return
	}
	session, err := b.store.CreateChallengeSession(ctx, queries.CreateChallengeSessionParams{
		StudentID:           link.StudentID,
		ChallengeExerciseID: chexID,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create challenge session from telegram")
		b.send(ctx, link.ChatID, "Kunde inte starta övningen. Försök igen.", nil)
		return
	}
	b.setActiveSession(ctx, link, session.ID)
	b.send(ctx, link.ChatID, fmt.Sprintf("🏆 %s\n\n%s\n\nSkriv ditt svar eller din fundering — jag guidar dig. Skriv /avsluta när du känner dig klar.", chex.Title, chex.Description), endKeyboard())
}

func (b *Bot) setActiveSession(ctx context.Context, link queries.TelegramLink, sessionID pgtype.UUID) {
	if _, err := b.store.UpsertTelegramSession(ctx, queries.UpsertTelegramSessionParams{
		ChatID:          link.ChatID,
		StudentID:       link.StudentID,
		State:           "in_session",
		ActiveSessionID: sessionID,
	}); err != nil {
		log.Error().Err(err).Msg("failed to set telegram session state")
	}
}

func (b *Bot) clearTelegramSession(ctx context.Context, link queries.TelegramLink) {
	if _, err := b.store.UpsertTelegramSession(ctx, queries.UpsertTelegramSessionParams{
		ChatID:    link.ChatID,
		StudentID: link.StudentID,
		State:     "menu",
	}); err != nil {
		log.Error().Err(err).Msg("failed to reset telegram session")
	}
}

func (b *Bot) decryptedAPIKey(ctx context.Context, studentID pgtype.UUID) (string, error) {
	student, err := b.store.GetStudentByID(ctx, studentID)
	if err != nil {
		return "", fmt.Errorf("student not found: %w", err)
	}
	rec, err := b.store.GetAPIKeyByParentID(ctx, student.ParentID)
	if err != nil {
		return "", fmt.Errorf("API key not found: %w", err)
	}
	plaintext, err := b.encSvc.Decrypt(rec.EncryptedKey)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}
	return string(plaintext), nil
}

func (b *Bot) handleDialogMessage(ctx context.Context, link queries.TelegramLink, tgSession queries.TelegramSession, text string) {
	sessionID := tgSession.ActiveSessionID
	session, err := b.store.GetSessionByID(ctx, sessionID)
	if err != nil || session.EndedAt.Valid {
		b.clearTelegramSession(ctx, link)
		b.sendMainMenu(ctx, link)
		return
	}
	if _, err := b.store.CreateMessage(ctx, queries.CreateMessageParams{
		SessionID: sessionID, Role: "user", Content: text,
	}); err != nil {
		log.Error().Err(err).Msg("failed to save user message")
		b.send(ctx, link.ChatID, "Något gick fel. Försök igen.", nil)
		return
	}
	var systemPrompt string
	if session.ExerciseID.Valid {
		ex, err := b.store.GetExerciseByID(ctx, session.ExerciseID)
		if err != nil {
			b.send(ctx, link.ChatID, "Övningen hittades inte.", nil)
			return
		}
		systemPrompt = ex.SystemPrompt
	} else {
		chex, err := b.store.GetChallengeExerciseByID(ctx, session.ChallengeExerciseID)
		if err != nil {
			b.send(ctx, link.ChatID, "Övningen hittades inte.", nil)
			return
		}
		systemPrompt = chex.SystemPrompt
	}
	systemPrompt += telegramPromptSuffix

	apiKey, err := b.decryptedAPIKey(ctx, link.StudentID)
	if err != nil {
		log.Error().Err(err).Msg("api key unavailable for telegram chat")
		b.send(ctx, link.ChatID, "Förälderns API-nyckel saknas eller är ogiltig.", nil)
		return
	}
	msgs, err := b.store.ListMessagesBySessionID(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list messages")
		b.send(ctx, link.ChatID, "Något gick fel. Försök igen.", nil)
		return
	}
	history := chat.BuildMessageHistory(msgs, chat.DefaultWindowSize)
	if err := b.api.SendChatAction(ctx, link.ChatID, "typing"); err != nil {
		log.Error().Err(err).Msg("failed to send typing action")
	}
	reply, err := b.complete(ctx, apiKey, systemPrompt, history)
	if err != nil {
		log.Error().Err(err).Msg("claude completion failed")
		b.send(ctx, link.ChatID, "AI-tjänsten svarade inte. Försök igen om en stund.", nil)
		return
	}
	if _, err := b.store.CreateMessage(ctx, queries.CreateMessageParams{
		SessionID: sessionID, Role: "assistant", Content: reply,
	}); err != nil {
		log.Error().Err(err).Msg("failed to save assistant message")
	}
	b.send(ctx, link.ChatID, reply, endKeyboard())
}

func (b *Bot) endActiveSession(ctx context.Context, link queries.TelegramLink) {
	tgSession, err := b.store.GetTelegramSession(ctx, link.ChatID)
	if err != nil || tgSession.State != "in_session" || !tgSession.ActiveSessionID.Valid {
		b.sendMainMenu(ctx, link)
		return
	}
	sessionID := tgSession.ActiveSessionID
	session, err := b.store.GetSessionByID(ctx, sessionID)
	if err != nil || session.EndedAt.Valid {
		b.clearTelegramSession(ctx, link)
		b.sendMainMenu(ctx, link)
		return
	}
	msgs, err := b.store.ListMessagesBySessionID(ctx, sessionID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list messages for scoring")
	}
	var title string
	if session.ExerciseID.Valid {
		if ex, err := b.store.GetExerciseByID(ctx, session.ExerciseID); err == nil {
			title = ex.Title
		}
	} else if chex, err := b.store.GetChallengeExerciseByID(ctx, session.ChallengeExerciseID); err == nil {
		title = chex.Title
	}

	result := &chat.ScoreResult{Score: 3, Summary: "Passet avslutades utan bedömning.", Feedback: "Bra jobbat att du övade!"}
	if apiKey, err := b.decryptedAPIKey(ctx, link.StudentID); err == nil {
		if scored, err := b.score(ctx, apiKey, title, msgs); err == nil && scored != nil {
			result = scored
		}
	}

	if _, err := b.store.CreateMessage(ctx, queries.CreateMessageParams{
		SessionID: sessionID, Role: "assistant", Content: result.Summary,
	}); err != nil {
		log.Error().Err(err).Msg("failed to save summary message")
	}
	if _, err := b.store.EndSession(ctx, queries.EndSessionParams{
		ID:      sessionID,
		Score:   pgtype.Int4{Int32: int32(result.Score), Valid: true},
		Summary: pgtype.Text{String: result.Summary, Valid: true},
	}); err != nil {
		log.Error().Err(err).Msg("failed to end session")
		b.send(ctx, link.ChatID, "Kunde inte avsluta passet. Försök igen.", nil)
		return
	}
	if session.ExerciseID.Valid {
		if err := chat.UpdateReviewSchedule(ctx, b.store, link.StudentID, session.ExerciseID, result.Score, time.Now()); err != nil {
			log.Error().Err(err).Msg("failed to update review schedule")
		}
	}
	b.clearTelegramSession(ctx, link)

	stars := challenges.ScoreToStars(result.Score)
	xp := stars * 10
	b.send(ctx, link.ChatID, fmt.Sprintf("%s %d stjärnor! +%d XP\n\n%s\n\n%s",
		strings.Repeat("⭐", stars), stars, xp, result.Summary, result.Feedback), nil)
	b.sendMainMenu(ctx, link)
}
```

- [ ] **Step 4: Ersätt meny-stubbarna i `menu.go`**

```go
func (b *Bot) sendSubjects(ctx context.Context, link queries.TelegramLink) {
	subjects, err := b.store.ListSubjects(ctx)
	if err != nil || len(subjects) == 0 {
		b.send(ctx, link.ChatID, "Kunde inte hämta ämnen just nu.", nil)
		return
	}
	var rows [][]InlineKeyboardButton
	for _, s := range subjects {
		rows = append(rows, []InlineKeyboardButton{{Text: s.Name, CallbackData: "subj:" + s.Slug}})
	}
	rows = append(rows, []InlineKeyboardButton{{Text: "⬅️ Meny", CallbackData: "menu"}})
	b.send(ctx, link.ChatID, "Välj ämne:", &InlineKeyboardMarkup{InlineKeyboard: rows})
}

func (b *Bot) sendTopics(ctx context.Context, link queries.TelegramLink, subjectSlug string) {
	subject, err := b.store.GetSubjectBySlug(ctx, subjectSlug)
	if err != nil {
		b.send(ctx, link.ChatID, "Ämnet hittades inte.", nil)
		return
	}
	topics, err := b.store.ListTopicsBySubjectID(ctx, subject.ID)
	if err != nil || len(topics) == 0 {
		b.send(ctx, link.ChatID, "Inga områden hittades.", nil)
		return
	}
	var rows [][]InlineKeyboardButton
	for _, t := range topics {
		rows = append(rows, []InlineKeyboardButton{{Text: t.Name, CallbackData: "top:" + subjectSlug + ":" + t.Slug}})
	}
	rows = append(rows, []InlineKeyboardButton{{Text: "⬅️ Ämnen", CallbackData: "study"}})
	b.send(ctx, link.ChatID, "Välj område:", &InlineKeyboardMarkup{InlineKeyboard: rows})
}

func (b *Bot) sendExercises(ctx context.Context, link queries.TelegramLink, subjectSlug, topicSlug string) {
	topic, err := b.store.GetTopicBySlug(ctx, queries.GetTopicBySlugParams{SubjectSlug: subjectSlug, TopicSlug: topicSlug})
	if err != nil {
		b.send(ctx, link.ChatID, "Området hittades inte.", nil)
		return
	}
	exercises, err := b.store.ListExercisesByTopicID(ctx, topic.ID)
	if err != nil || len(exercises) == 0 {
		b.send(ctx, link.ChatID, "Inga övningar hittades.", nil)
		return
	}
	var rows [][]InlineKeyboardButton
	for _, e := range exercises {
		rows = append(rows, []InlineKeyboardButton{{Text: e.Title, CallbackData: "ex:" + uuidToString(e.ID)}})
	}
	rows = append(rows, []InlineKeyboardButton{{Text: "⬅️ Områden", CallbackData: "subj:" + subjectSlug}})
	b.send(ctx, link.ChatID, "Välj övning:", &InlineKeyboardMarkup{InlineKeyboard: rows})
}

func (b *Bot) sendChallenges(ctx context.Context, link queries.TelegramLink) {
	student, err := b.store.GetStudentByID(ctx, link.StudentID)
	if err != nil {
		b.send(ctx, link.ChatID, "Något gick fel.", nil)
		return
	}
	chals, err := b.store.ListPublishedChallengesByParentID(ctx, student.ParentID)
	if err != nil || len(chals) == 0 {
		b.send(ctx, link.ChatID, "Inga utmaningar ännu — be din förälder ladda upp en läxa på webben!", nil)
		return
	}
	var rows [][]InlineKeyboardButton
	for _, c := range chals {
		rows = append(rows, []InlineKeyboardButton{{Text: c.CoverEmoji + " " + c.Title, CallbackData: "chal:" + uuidToString(c.ID)}})
	}
	rows = append(rows, []InlineKeyboardButton{{Text: "⬅️ Meny", CallbackData: "menu"}})
	b.send(ctx, link.ChatID, "Välj utmaning:", &InlineKeyboardMarkup{InlineKeyboard: rows})
}

func (b *Bot) sendChallengeExercises(ctx context.Context, link queries.TelegramLink, challengeIDStr string) {
	exercises, err := b.store.ListChallengeExercisesWithProgress(ctx, queries.ListChallengeExercisesWithProgressParams{
		ChallengeID: parseUUID(challengeIDStr),
		StudentID:   link.StudentID,
	})
	if err != nil || len(exercises) == 0 {
		b.send(ctx, link.ChatID, "Inga övningar i den här utmaningen.", nil)
		return
	}
	var rows [][]InlineKeyboardButton
	for _, e := range exercises {
		label := e.Title
		if e.Score.Valid {
			label = strings.Repeat("⭐", challenges.ScoreToStars(int(e.Score.Int32))) + " " + label
		}
		rows = append(rows, []InlineKeyboardButton{{Text: label, CallbackData: "chex:" + uuidToString(e.ID)}})
	}
	rows = append(rows, []InlineKeyboardButton{{Text: "⬅️ Utmaningar", CallbackData: "chals"}})
	b.send(ctx, link.ChatID, "Välj övning:", &InlineKeyboardMarkup{InlineKeyboard: rows})
}

func (b *Bot) sendDueReviews(ctx context.Context, link queries.TelegramLink) {
	due, err := b.store.ListDueReviews(ctx, link.StudentID)
	if err != nil {
		b.send(ctx, link.ChatID, "Kunde inte hämta repetitioner.", nil)
		return
	}
	if len(due) == 0 {
		b.send(ctx, link.ChatID, "Inget att repetera just nu — snyggt jobbat! 🎉", nil)
		b.sendMainMenu(ctx, link)
		return
	}
	var rows [][]InlineKeyboardButton
	for _, r := range due {
		rows = append(rows, []InlineKeyboardButton{{Text: r.SubjectName + ": " + r.ExerciseTitle, CallbackData: "ex:" + uuidToString(r.ExerciseID)}})
	}
	rows = append(rows, []InlineKeyboardButton{{Text: "⬅️ Meny", CallbackData: "menu"}})
	b.send(ctx, link.ChatID, "Dags att repetera:", &InlineKeyboardMarkup{InlineKeyboard: rows})
}

func (b *Bot) sendProgress(ctx context.Context, link queries.TelegramLink) {
	rows, err := b.store.GetStudentProgressBySubject(ctx, link.StudentID)
	if err != nil || len(rows) == 0 {
		b.send(ctx, link.ChatID, "Ingen progress ännu — kör din första övning! 💪", nil)
		return
	}
	var sb strings.Builder
	sb.WriteString("📊 Din progress:\n\n")
	for _, r := range rows {
		fmt.Fprintf(&sb, "%s: %d pass, snitt %.1f/5, %d övningar\n", r.Name, r.TotalSessions, r.AvgScore, r.UniqueExercises)
	}
	kb := &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{{{Text: "⬅️ Meny", CallbackData: "menu"}}}}
	b.send(ctx, link.ChatID, sb.String(), kb)
}
```

(Uppdatera imports i menu.go: lägg till `strings` och `challenges`-paketet.)

- [ ] **Step 5: Verifiera grönt** — `go test ./internal/telegram/ -v && go test ./... && go build ./...` → PASS

- [ ] **Step 6: Commit**

```bash
git add internal/telegram/
git commit -m "feat(telegram): browse menus, socratic dialog, scoring with stars and SM-2"
```

---

### Task 8: Pushar — SRS-påminnelser + notis vid publicerad utmaning

**Files:**
- Create: `backend/internal/telegram/notify.go`
- Modify: `backend/internal/challenges/handler.go` (Notifier-hook i `Publish`)
- Test: `backend/internal/telegram/notify_test.go`

**Interfaces:**
- Consumes: `ListTelegramLinks`, `ListDueReviews`, `TryInsertTelegramReminder`, `GetChallengeByID`, `GetStudentByID`.
- Produces: `RunReminderLoop(ctx context.Context, bot *Bot)`; `(*Bot).sendDueReminders(ctx, now time.Time)` (exponerad för test); `(*Bot).ChallengePublished(ctx context.Context, challengeID pgtype.UUID)`; i challenges-paketet: `type Notifier interface { ChallengePublished(ctx context.Context, challengeID pgtype.UUID) }` + `(*ChallengeHandler).SetNotifier(n Notifier)`.

- [ ] **Step 1: Skriv failande tester**

`backend/internal/telegram/notify_test.go`:

```go
package telegram

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

type notifyStore struct {
	*fakeStore
	due       map[string][]queries.ListDueReviewsRow // key: uuidToString(studentID)
	reminders map[string]bool                        // key: studentID+date
}

func newNotifyStore() *notifyStore {
	return &notifyStore{fakeStore: newFakeStore(), due: map[string][]queries.ListDueReviewsRow{}, reminders: map[string]bool{}}
}

func (n *notifyStore) ListTelegramLinks(_ context.Context) ([]queries.TelegramLink, error) {
	var links []queries.TelegramLink
	for _, l := range n.links {
		links = append(links, l)
	}
	return links, nil
}
func (n *notifyStore) ListDueReviews(_ context.Context, studentID pgtype.UUID) ([]queries.ListDueReviewsRow, error) {
	return n.due[uuidToString(studentID)], nil
}
func (n *notifyStore) TryInsertTelegramReminder(_ context.Context, arg queries.TryInsertTelegramReminderParams) (int64, error) {
	key := uuidToString(arg.StudentID) + arg.SentOn.Time.Format("2006-01-02")
	if n.reminders[key] {
		return 0, nil
	}
	n.reminders[key] = true
	return 1, nil
}

func TestReminderSentOncePerDayInsideWindow(t *testing.T) {
	store := newNotifyStore()
	studentID := testUUID("11111111-1111-1111-1111-111111111111")
	store.links[777] = queries.TelegramLink{StudentID: studentID, TelegramUserID: 777, ChatID: 555}
	store.due[uuidToString(studentID)] = []queries.ListDueReviewsRow{{ExerciseTitle: "Fotosyntes"}}
	sender := &fakeSender{}
	bot := NewBot(sender, store, nil)

	at16 := time.Date(2026, 7, 4, 16, 0, 0, 0, time.Local)
	bot.sendDueReminders(context.Background(), at16)
	bot.sendDueReminders(context.Background(), at16.Add(time.Hour))

	if len(sender.messages) != 1 {
		t.Fatalf("expected exactly 1 reminder, got %d", len(sender.messages))
	}
	if sender.messages[0].kb == nil {
		t.Error("reminder should carry a 'Kör nu' button")
	}
}

func TestNoReminderOutsideWindow(t *testing.T) {
	store := newNotifyStore()
	studentID := testUUID("11111111-1111-1111-1111-111111111111")
	store.links[777] = queries.TelegramLink{StudentID: studentID, TelegramUserID: 777, ChatID: 555}
	store.due[uuidToString(studentID)] = []queries.ListDueReviewsRow{{ExerciseTitle: "Fotosyntes"}}
	sender := &fakeSender{}
	bot := NewBot(sender, store, nil)

	bot.sendDueReminders(context.Background(), time.Date(2026, 7, 4, 9, 0, 0, 0, time.Local))

	if len(sender.messages) != 0 {
		t.Errorf("expected no reminder at 09:00, got %v", sender.messages)
	}
}

func TestChallengePublishedNotifiesFamilyChildren(t *testing.T) {
	store := newNotifyStore()
	parentID := testUUID("44444444-4444-4444-4444-444444444444")
	otherParent := testUUID("55555555-5555-5555-5555-555555555555")
	kid1 := testUUID("11111111-1111-1111-1111-111111111111")
	kid2 := testUUID("22222222-2222-2222-2222-222222222222")
	store.students[uuidToString(kid1)] = queries.Student{ID: kid1, ParentID: parentID}
	store.students[uuidToString(kid2)] = queries.Student{ID: kid2, ParentID: otherParent}
	store.links[701] = queries.TelegramLink{StudentID: kid1, TelegramUserID: 701, ChatID: 501}
	store.links[702] = queries.TelegramLink{StudentID: kid2, TelegramUserID: 702, ChatID: 502}
	chalID := testUUID("66666666-6666-6666-6666-666666666666")
	store.challenge = &queries.Challenge{ID: chalID, ParentID: parentID, Title: "Glosor v.28", CoverEmoji: "📖", Published: true}
	sender := &fakeSender{}
	bot := NewBot(sender, store, nil)

	bot.ChallengePublished(context.Background(), chalID)

	if len(sender.messages) != 1 || sender.messages[0].chatID != 501 {
		t.Fatalf("expected exactly one notification to chat 501, got %v", sender.messages)
	}
}
```

Lägg till i `notifyStore`:

```go
func (n *notifyStore) GetChallengeByID(_ context.Context, _ pgtype.UUID) (queries.Challenge, error) {
	if n.challenge == nil {
		return queries.Challenge{}, errors.New("not found")
	}
	return *n.challenge, nil
}
```
(och fältet `challenge *queries.Challenge` i structen + `errors`-import).

- [ ] **Step 2: Verifiera FAIL** — `go test ./internal/telegram/ -run 'Reminder|ChallengePublished' -v`

- [ ] **Step 3: Implementera `notify.go`**

```go
package telegram

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// RunReminderLoop checks hourly for due reviews and sends at most one
// reminder per student per day, inside the 15:00–20:00 local window.
func RunReminderLoop(ctx context.Context, bot *Bot) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bot.sendDueReminders(ctx, time.Now())
		}
	}
}

func (b *Bot) sendDueReminders(ctx context.Context, now time.Time) {
	if now.Hour() < 15 || now.Hour() >= 20 {
		return
	}
	links, err := b.store.ListTelegramLinks(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to list telegram links")
		return
	}
	for _, link := range links {
		due, err := b.store.ListDueReviews(ctx, link.StudentID)
		if err != nil || len(due) == 0 {
			continue
		}
		inserted, err := b.store.TryInsertTelegramReminder(ctx, queries.TryInsertTelegramReminderParams{
			StudentID: link.StudentID,
			SentOn:    pgtype.Date{Time: now, Valid: true},
		})
		if err != nil || inserted == 0 {
			continue // already reminded today
		}
		kb := &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "⭐ Kör nu", CallbackData: "rev"}},
		}}
		b.send(ctx, link.ChatID, fmt.Sprintf("Dags att repetera! 🧠 Du har %d övning(ar) som väntar, t.ex. %s.", len(due), due[0].ExerciseTitle), kb)
	}
}

// ChallengePublished notifies all linked students in the challenge's family.
// Implements challenges.Notifier.
func (b *Bot) ChallengePublished(ctx context.Context, challengeID pgtype.UUID) {
	challenge, err := b.store.GetChallengeByID(ctx, challengeID)
	if err != nil || !challenge.Published {
		return
	}
	links, err := b.store.ListTelegramLinks(ctx)
	if err != nil {
		return
	}
	for _, link := range links {
		student, err := b.store.GetStudentByID(ctx, link.StudentID)
		if err != nil || student.ParentID != challenge.ParentID {
			continue
		}
		kb := &InlineKeyboardMarkup{InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🏆 Öppna", CallbackData: "chal:" + uuidToString(challenge.ID)}},
		}}
		b.send(ctx, link.ChatID, fmt.Sprintf("Ny läxutmaning: %s %s", challenge.CoverEmoji, challenge.Title), kb)
	}
}
```

- [ ] **Step 4: Hooka publicering i challenges**

I `internal/challenges/handler.go`: lägg till nära handler-structen:

```go
// Notifier is told when a challenge is published (e.g. the Telegram bot).
type Notifier interface {
	ChallengePublished(ctx context.Context, challengeID pgtype.UUID)
}

func (h *ChallengeHandler) SetNotifier(n Notifier) { h.notifier = n }
```

Lägg fältet `notifier Notifier` i `ChallengeHandler`-structen. I `Publish`-metoden, direkt efter lyckad publicering (leta upp stället där `PublishChallenge` returnerat utan fel, före svaret skrivs):

```go
	if h.notifier != nil {
		go h.notifier.ChallengePublished(context.WithoutCancel(r.Context()), challenge.ID)
	}
```

(`challenge` = raden som `PublishChallenge` returnerar; kontrollera variabelnamnet på plats. Lägg till `context`-import om den saknas. OBS: endast `Publish`-endpointen — barn-skapade challenges som auto-publiceras ska INTE notifiera, barnet skapade den själv.)

- [ ] **Step 5: Verifiera grönt** — `go test ./... && go build ./...` → PASS

- [ ] **Step 6: Commit**

```bash
git add internal/telegram/ internal/challenges/
git commit -m "feat(telegram): SRS reminders and challenge-published notifications"
```

---

### Task 9: Wiring i main.go + miljövariabler + README

**Files:**
- Modify: `backend/cmd/server/main.go`
- Modify: `README.md` (kort Telegram-sektion under setup)

**Interfaces:**
- Consumes: allt ovan. Env: `TELEGRAM_BOT_TOKEN`, `TELEGRAM_BOT_USERNAME` (utan @).

- [ ] **Step 1: Wiring**

I `main.go`, efter challenges-handler-initieringen (rad ~93) lägg till:

```go
	// Telegram bot (optional — enabled when TELEGRAM_BOT_TOKEN is set)
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramBotUsername := os.Getenv("TELEGRAM_BOT_USERNAME")
	var telegramStore *telegram.QueriesStore
	if telegramToken != "" {
		telegramStore = telegram.NewQueriesStore(q)
		api := telegram.NewAPI(telegramToken)
		bot := telegram.NewBot(api, telegramStore, encSvc)
		challengeHandler.SetNotifier(bot)
		go telegram.Run(ctx, api, bot)
		go telegram.RunReminderLoop(ctx, bot)
		log.Info().Msg("telegram bot started")
	}
```

Och bland routes (efter `/api/sessions`-blocket, rad ~233):

```go
	// Telegram link route (child only) — only when the bot is enabled
	if telegramToken != "" {
		linkHandler := telegram.NewLinkHandler(telegramStore, telegramBotUsername)
		r.Route("/api/telegram", func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))
			r.Use(auth.ChildOnly)
			r.Post("/link-code", linkHandler.CreateLinkCode)
		})
	}
```

Import: `"github.com/trollstaven/nioplugget/backend/internal/telegram"`.

- [ ] **Step 2: README**

Lägg till under setup-sektionen i `README.md`:

```markdown
### Telegram-bot (valfritt)

Barnet kan plugga via Telegram. Skapa en bot hos @BotFather och sätt i `backend/.env`:

```
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
TELEGRAM_BOT_USERNAME=NiopluggetBot
```

Utan token startar boten inte. Barnet kopplar sitt konto via knappen
"Koppla Telegram" på studysidan.
```

- [ ] **Step 3: Verifiera** — `go build ./... && go test ./...` grönt. Starta backend utan `TELEGRAM_BOT_TOKEN` och kontrollera att loggen INTE innehåller "telegram bot started" och att `/health` svarar `ok`:

```bash
set -a && source .env && set +a && go run ./cmd/server/main.go &
sleep 3 && curl -s localhost:8080/health && kill %1
```

- [ ] **Step 4: Commit**

```bash
git add cmd/server/main.go ../README.md
git commit -m "feat(telegram): wire bot, reminder loop and link route into server"
```

---

### Task 10: Frontend — "Koppla Telegram" på studysidan

**Files:**
- Modify: `frontend/src/lib/api.ts`
- Modify: `frontend/src/routes/study/+page.svelte`

**Interfaces:**
- Consumes: `POST /api/telegram/link-code` → `{"code":"ABCD2345","link":"https://t.me/<bot>?start=ABCD2345"}`.

- [ ] **Step 1: Läs `frontend/src/lib/api.ts` och följ dess exportmönster.** Lägg till:

```ts
export const telegram = {
	createLinkCode: () =>
		apiFetch('/api/telegram/link-code', { method: 'POST' }) as Promise<{ code: string; link: string }>
};
```

- [ ] **Step 2: Läs `frontend/src/routes/study/+page.svelte` och lägg till en sektion längst ner** (följ sidans befintliga shadcn-svelte-komponenter och styling):

```svelte
<script lang="ts">
	// ...befintliga imports; lägg till:
	import { telegram } from '$lib/api';

	let telegramLink: string | null = null;
	let telegramError = '';

	async function connectTelegram() {
		telegramError = '';
		try {
			const res = await telegram.createLinkCode();
			telegramLink = res.link;
		} catch {
			telegramError = 'Kunde inte skapa kopplingslänk. Försök igen.';
		}
	}
</script>

<!-- längst ner på sidan -->
<section class="mt-8">
	{#if telegramLink}
		<p>
			<a href={telegramLink} target="_blank" rel="noopener" class="underline">
				📲 Öppna Telegram och tryck Start →
			</a>
		</p>
		<p class="text-sm text-muted-foreground">Länken gäller i 15 minuter.</p>
	{:else}
		<button type="button" onclick={connectTelegram}>Koppla Telegram</button>
		{#if telegramError}<p class="text-sm text-destructive">{telegramError}</p>{/if}
	{/if}
</section>
```

(Byt `<button>` mot sidans Button-komponent om en sådan används.)

- [ ] **Step 3: Verifiera** — `cd ../frontend && npx svelte-check --threshold error 2>&1 | tail -5` utan nya fel; `npm run build` grönt.

- [ ] **Step 4: Commit**

```bash
git add src/lib/api.ts src/routes/study/+page.svelte
git commit -m "feat(telegram): connect-Telegram button on study page"
```

---

### Task 11: E2E-verifiering mot riktiga Telegram

**Files:** inga nya — manuell verifiering + ev. buggfixar.

- [ ] **Step 1:** Skapa boten hos @BotFather (användaren gör detta; be om token + botnamn). Lägg in `TELEGRAM_BOT_TOKEN` och `TELEGRAM_BOT_USERNAME` i `backend/.env`.
- [ ] **Step 2:** Starta stacken (`./start.sh` eller motsvarande). Verifiera i loggen: `telegram bot started`.
- [ ] **Step 3:** Logga in som barn i webben → "Koppla Telegram" → öppna länken → boten svarar "Kopplat! 🎉" + meny.
- [ ] **Step 4:** Kör en hel övning: 📚 Plugga → ämne → område → övning → 2-3 dialogmeddelanden → "✅ Klar" → stjärnor/XP visas; kontrollera i webben att sessionen syns under progress.
- [ ] **Step 5:** Publicera en läxutmaning som förälder i webben → notis dyker upp i Telegram → kör en utmaningsövning.
- [ ] **Step 6:** Skicka ett meddelande från ett okopplat Telegram-konto → artigt avvisande.
- [ ] **Step 7:** Committa eventuella fixar; uppdatera specen om beteendet justerats.

---

## Self-review (utförd vid planskrivning)

- Spec-täckning: koppling (Task 1/5/6/10), konversationsflöde (6/7), matte-suffix (7), pushar (8), arkitektur/felhantering (4/6/9), utanför scope respekterat.
- Kända osäkerheter markerade som kontrollsteg (inte placeholders): fältnamn i genererade sqlc-typer (Task 5 Step 1), `apikey.NewEncryptionService`-nyckelformat (Task 7 Step 1), variabelnamn i `Publish`-handlern (Task 8 Step 4), exportmönster i `api.ts` (Task 10).
- Typkonsistens: `Sender`-interfacet matchar `*API`-metoderna; `CompleteFunc`/`ScoreFunc` matchar `chat.CompleteChatResponse`/`chat.ScoreSession`; callbackdata-schemat är detsamma i router, menyer och notiser.
