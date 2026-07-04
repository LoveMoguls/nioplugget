package telegram

import (
	"context"
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

// createLinkCode generates a link code, persists it, and builds the deep
// link. It is extracted from the HTTP handler so it can be unit-tested
// without needing to inject auth's JWT context machinery.
func createLinkCode(ctx context.Context, store Store, botUsername, studentID string) (code, link string, err error) {
	code, err = generateLinkCode()
	if err != nil {
		return "", "", err
	}
	if _, err = store.CreateTelegramLinkCode(ctx, queries.CreateTelegramLinkCodeParams{
		Code:      code,
		StudentID: parseUUID(studentID),
		ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(linkCodeTTL), Valid: true},
	}); err != nil {
		return "", "", err
	}
	link = "https://t.me/" + botUsername + "?start=" + code
	return code, link, nil
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
	code, link, err := createLinkCode(r.Context(), h.store, h.botUsername, studentID)
	if err != nil {
		log.Error().Err(err).Msg("failed to create telegram link code")
		http.Error(w, `{"error":"Kunde inte skapa kod"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"code": code,
		"link": link,
	})
}
