package chat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/apikey"
	"github.com/trollstaven/nioplugget/backend/internal/auth"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
	"github.com/trollstaven/nioplugget/backend/internal/srs"
)

// ChatHandler provides HTTP handlers for chat operations.
type ChatHandler struct {
	store  ChatStore
	encSvc *apikey.EncryptionService
}

// NewChatHandler creates a new ChatHandler.
func NewChatHandler(store ChatStore, encSvc *apikey.EncryptionService) *ChatHandler {
	return &ChatHandler{store: store, encSvc: encSvc}
}

type createSessionRequest struct {
	ExerciseID string `json:"exerciseId"`
}

type createSessionResponse struct {
	ID         string `json:"id"`
	ExerciseID string `json:"exerciseId"`
	StartedAt  string `json:"startedAt"`
}

type sendMessageRequest struct {
	Content string `json:"content"`
}

type messageResponse struct {
	ID        string `json:"id"`
	SessionID string `json:"sessionId"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"createdAt"`
}

type endSessionResponse struct {
	Score    int    `json:"score"`
	Summary  string `json:"summary"`
	Feedback string `json:"feedback"`
	EndedAt  string `json:"endedAt"`
}

// CreateSession handles POST /api/sessions
func (h *ChatHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	studentID := auth.GetUserIDFromContext(r.Context())
	if studentID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Ogiltig begäran"}`, http.StatusBadRequest)
		return
	}

	studentUUID := parseUUID(studentID)
	exerciseUUID := parseUUID(req.ExerciseID)

	session, err := h.store.CreateSession(r.Context(), queries.CreateSessionParams{
		StudentID:  studentUUID,
		ExerciseID: exerciseUUID,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create session")
		http.Error(w, `{"error":"Kunde inte starta session"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, createSessionResponse{
		ID:         uuidToString(session.ID),
		ExerciseID: uuidToString(session.ExerciseID),
		StartedAt:  session.StartedAt.Time.Format("2006-01-02T15:04:05Z"),
	})
}

// SendMessage handles POST /api/sessions/{id}/messages
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	studentID := auth.GetUserIDFromContext(r.Context())
	if studentID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	sessionUUID := parseUUID(sessionID)

	// Verify session belongs to student and is not ended
	session, err := h.store.GetSessionByID(r.Context(), sessionUUID)
	if err != nil {
		http.Error(w, `{"error":"Session hittades inte"}`, http.StatusNotFound)
		return
	}

	if uuidToString(session.StudentID) != studentID {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}

	if session.EndedAt.Valid {
		http.Error(w, `{"error":"Sessionen är redan avslutad"}`, http.StatusBadRequest)
		return
	}

	var req sendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
		http.Error(w, `{"error":"Meddelande saknas"}`, http.StatusBadRequest)
		return
	}

	// Save user message
	_, err = h.store.CreateMessage(r.Context(), queries.CreateMessageParams{
		SessionID: sessionUUID,
		Role:      "user",
		Content:   req.Content,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to save user message")
		http.Error(w, `{"error":"Kunde inte spara meddelande"}`, http.StatusInternalServerError)
		return
	}

	// Get exercise for system prompt
	exercise, err := h.store.GetExerciseByID(r.Context(), session.ExerciseID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get exercise")
		http.Error(w, `{"error":"Övning hittades inte"}`, http.StatusInternalServerError)
		return
	}

	// Get API key
	decryptedKey, err := h.getDecryptedAPIKey(r, session.StudentID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get API key")
		http.Error(w, `{"error":"Förälderns API-nyckel saknas eller är ogiltig"}`, http.StatusInternalServerError)
		return
	}

	// Load all messages for context
	allMessages, err := h.store.ListMessagesBySessionID(r.Context(), sessionUUID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list messages")
		http.Error(w, `{"error":"Kunde inte ladda konversation"}`, http.StatusInternalServerError)
		return
	}

	// Build message history with sliding window
	messageHistory := BuildMessageHistory(allMessages, DefaultWindowSize)

	// Stream response
	fullResponse, err := StreamChatResponse(w, r.Context(), decryptedKey, exercise.SystemPrompt, messageHistory)
	if err != nil {
		log.Error().Err(err).Msg("stream error (response may be partial)")
	}

	// Save AI response (even if partial)
	if fullResponse != "" {
		_, err = h.store.CreateMessage(r.Context(), queries.CreateMessageParams{
			SessionID: sessionUUID,
			Role:      "assistant",
			Content:   fullResponse,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to save AI response")
		}
	}
}

// ListMessages handles GET /api/sessions/{id}/messages
func (h *ChatHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	studentID := auth.GetUserIDFromContext(r.Context())
	if studentID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	sessionUUID := parseUUID(sessionID)

	// Verify session belongs to student
	session, err := h.store.GetSessionByID(r.Context(), sessionUUID)
	if err != nil {
		http.Error(w, `{"error":"Session hittades inte"}`, http.StatusNotFound)
		return
	}

	if uuidToString(session.StudentID) != studentID {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}

	messages, err := h.store.ListMessagesBySessionID(r.Context(), sessionUUID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list messages")
		http.Error(w, `{"error":"Kunde inte hämta meddelanden"}`, http.StatusInternalServerError)
		return
	}

	result := make([]messageResponse, len(messages))
	for i, m := range messages {
		result[i] = messageResponse{
			ID:        uuidToString(m.ID),
			SessionID: uuidToString(m.SessionID),
			Role:      m.Role,
			Content:   m.Content,
			CreatedAt: m.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// GetSession handles GET /api/sessions/{id}
func (h *ChatHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	studentID := auth.GetUserIDFromContext(r.Context())
	if studentID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	sessionUUID := parseUUID(sessionID)
	session, err := h.store.GetSessionByID(r.Context(), sessionUUID)
	if err != nil {
		http.Error(w, `{"error":"Session hittades inte"}`, http.StatusNotFound)
		return
	}

	if uuidToString(session.StudentID) != studentID {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}

	resp := map[string]interface{}{
		"id":         uuidToString(session.ID),
		"exerciseId": uuidToString(session.ExerciseID),
		"startedAt":  session.StartedAt.Time.Format("2006-01-02T15:04:05Z"),
	}

	if session.EndedAt.Valid {
		resp["endedAt"] = session.EndedAt.Time.Format("2006-01-02T15:04:05Z")
	}
	if session.Score.Valid {
		resp["score"] = session.Score.Int32
	}
	if session.Summary.Valid {
		resp["summary"] = session.Summary.String
	}

	writeJSON(w, http.StatusOK, resp)
}

// EndSession handles POST /api/sessions/{id}/end
func (h *ChatHandler) EndSession(w http.ResponseWriter, r *http.Request) {
	sessionID := chi.URLParam(r, "id")
	studentID := auth.GetUserIDFromContext(r.Context())
	if studentID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	sessionUUID := parseUUID(sessionID)

	// Verify session
	session, err := h.store.GetSessionByID(r.Context(), sessionUUID)
	if err != nil {
		http.Error(w, `{"error":"Session hittades inte"}`, http.StatusNotFound)
		return
	}

	if uuidToString(session.StudentID) != studentID {
		http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
		return
	}

	if session.EndedAt.Valid {
		http.Error(w, `{"error":"Sessionen är redan avslutad"}`, http.StatusBadRequest)
		return
	}

	// Load messages and exercise
	allMessages, err := h.store.ListMessagesBySessionID(r.Context(), sessionUUID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list messages for scoring")
		http.Error(w, `{"error":"Kunde inte ladda konversation"}`, http.StatusInternalServerError)
		return
	}

	exercise, err := h.store.GetExerciseByID(r.Context(), session.ExerciseID)
	if err != nil {
		log.Error().Err(err).Msg("failed to get exercise for scoring")
		http.Error(w, `{"error":"Övning hittades inte"}`, http.StatusInternalServerError)
		return
	}

	// Get API key for scoring
	decryptedKey, err := h.getDecryptedAPIKey(r, session.StudentID)

	var scoreResult *ScoreResult
	if err != nil {
		log.Warn().Err(err).Msg("API key not available for scoring, using default")
		scoreResult = defaultScore()
	} else {
		scoreResult, err = ScoreSession(r.Context(), decryptedKey, exercise, allMessages)
		if err != nil {
			log.Error().Err(err).Msg("scoring failed, using default")
			scoreResult = defaultScore()
		}
	}

	// Save summary as final message
	_, err = h.store.CreateMessage(r.Context(), queries.CreateMessageParams{
		SessionID: sessionUUID,
		Role:      "assistant",
		Content:   scoreResult.Summary,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to save summary message")
	}

	// End session with score
	updatedSession, err := h.store.EndSession(r.Context(), queries.EndSessionParams{
		ID:      sessionUUID,
		Score:   pgtype.Int4{Int32: int32(scoreResult.Score), Valid: true},
		Summary: pgtype.Text{String: scoreResult.Summary, Valid: true},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to end session")
		http.Error(w, `{"error":"Kunde inte avsluta sessionen"}`, http.StatusInternalServerError)
		return
	}

	// Update spaced repetition schedule via SM-2 (non-blocking — failure does not affect response)
	srsCtx := r.Context()
	srsStudentUUID := parseUUID(studentID)
	srsExerciseID := session.ExerciseID
	srsScore := scoreResult.Score
	go func() {
		// Get existing schedule (if any) for current EF and interval
		sm2Input := srs.SM2Input{
			Score:           srsScore,
			EaseFactor:      2.5, // SM-2 default
			IntervalDays:    1,
			RepetitionCount: 0,
		}

		existing, err := h.store.GetReviewSchedule(srsCtx, queries.GetReviewScheduleParams{
			StudentID:  srsStudentUUID,
			ExerciseID: srsExerciseID,
		})
		if err == nil {
			sm2Input.EaseFactor = float64(existing.EaseFactor)
			sm2Input.IntervalDays = int(existing.IntervalDays)
			sm2Input.RepetitionCount = int(existing.RepetitionCount)
		}

		output := srs.Calculate(sm2Input, time.Now())

		_, err = h.store.UpsertReviewSchedule(srsCtx, queries.UpsertReviewScheduleParams{
			StudentID:       srsStudentUUID,
			ExerciseID:      srsExerciseID,
			EaseFactor:      float32(output.EaseFactor),
			IntervalDays:    int32(output.IntervalDays),
			RepetitionCount: int32(output.RepetitionCount),
			NextReview:      pgtype.Timestamptz{Time: output.NextReview, Valid: true},
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to update review schedule")
		}
	}()

	writeJSON(w, http.StatusOK, endSessionResponse{
		Score:    scoreResult.Score,
		Summary:  scoreResult.Summary,
		Feedback: scoreResult.Feedback,
		EndedAt:  updatedSession.EndedAt.Time.Format("2006-01-02T15:04:05Z"),
	})
}

// getDecryptedAPIKey looks up the student's parent and decrypts their API key.
func (h *ChatHandler) getDecryptedAPIKey(r *http.Request, studentID pgtype.UUID) (string, error) {
	student, err := h.store.GetStudentByID(r.Context(), studentID)
	if err != nil {
		return "", fmt.Errorf("student not found: %w", err)
	}

	apiKeyRecord, err := h.store.GetAPIKeyByParentID(r.Context(), student.ParentID)
	if err != nil {
		return "", fmt.Errorf("API key not found: %w", err)
	}

	plaintext, err := h.encSvc.Decrypt(apiKeyRecord.EncryptedKey)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

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
