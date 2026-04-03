package srs

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/auth"
)

// SRSHandler provides HTTP handlers for spaced repetition operations.
type SRSHandler struct {
	store SRSStore
}

// NewSRSHandler creates a new SRSHandler.
func NewSRSHandler(store SRSStore) *SRSHandler {
	return &SRSHandler{store: store}
}

type dueReviewResponse struct {
	ID            string `json:"id"`
	ExerciseID    string `json:"exerciseId"`
	ExerciseTitle string `json:"exerciseTitle"`
	TopicName     string `json:"topicName"`
	TopicSlug     string `json:"topicSlug"`
	SubjectName   string `json:"subjectName"`
	SubjectSlug   string `json:"subjectSlug"`
	NextReview    string `json:"nextReview"`
	DaysOverdue   int    `json:"daysOverdue"`
}

// ListDueReviews handles GET /api/reviews/due
func (h *SRSHandler) ListDueReviews(w http.ResponseWriter, r *http.Request) {
	studentID := auth.GetUserIDFromContext(r.Context())
	if studentID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	studentUUID := parseUUID(studentID)

	rows, err := h.store.ListDueReviews(r.Context(), studentUUID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list due reviews")
		http.Error(w, `{"error":"Kunde inte hämta repetitioner"}`, http.StatusInternalServerError)
		return
	}

	result := make([]dueReviewResponse, len(rows))
	for i, row := range rows {
		daysOverdue := 0
		if row.NextReview.Valid {
			hours := time.Since(row.NextReview.Time).Hours()
			if hours > 0 {
				daysOverdue = int(hours / 24)
			}
		}
		result[i] = dueReviewResponse{
			ID:            uuidToString(row.ID),
			ExerciseID:    uuidToString(row.ExerciseID),
			ExerciseTitle: row.ExerciseTitle,
			TopicName:     row.TopicName,
			TopicSlug:     row.TopicSlug,
			SubjectName:   row.SubjectName,
			SubjectSlug:   row.SubjectSlug,
			NextReview:    row.NextReview.Time.Format("2006-01-02T15:04:05Z"),
			DaysOverdue:   daysOverdue,
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// uuidToString converts a pgtype.UUID to its string representation.
func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	b := u.Bytes
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// parseUUID parses a UUID string into a pgtype.UUID.
func parseUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}
	}
	return u
}
