package progress

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/auth"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// ProgressHandler provides HTTP handlers for progress views.
type ProgressHandler struct {
	store ProgressStore
}

// NewProgressHandler creates a new ProgressHandler.
func NewProgressHandler(store ProgressStore) *ProgressHandler {
	return &ProgressHandler{store: store}
}

// --- Response types ---

type topicProgressResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Slug            string  `json:"slug"`
	TotalSessions   int32   `json:"totalSessions"`
	AvgScore        float64 `json:"avgScore"`
	UniqueExercises int32   `json:"uniqueExercises"`
}

type subjectProgressResponse struct {
	ID              string                  `json:"id"`
	Name            string                  `json:"name"`
	Slug            string                  `json:"slug"`
	TotalSessions   int32                   `json:"totalSessions"`
	AvgScore        float64                 `json:"avgScore"`
	UniqueExercises int32                   `json:"uniqueExercises"`
	Topics          []topicProgressResponse `json:"topics"`
}

type progressResponse struct {
	Subjects []subjectProgressResponse `json:"subjects"`
}

type sessionResponse struct {
	ID            string `json:"id"`
	Score         int32  `json:"score"`
	StartedAt     string `json:"startedAt"`
	EndedAt       string `json:"endedAt"`
	ExerciseTitle string `json:"exerciseTitle"`
	TopicName     string `json:"topicName"`
	SubjectName   string `json:"subjectName"`
}

type sessionsListResponse struct {
	Sessions []sessionResponse `json:"sessions"`
}

// GetStudentProgress handles GET /api/progress — student's own progress.
func (h *ProgressHandler) GetStudentProgress(w http.ResponseWriter, r *http.Request) {
	studentID := auth.GetUserIDFromContext(r.Context())
	if studentID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	studentUUID := parseUUID(studentID)
	resp, err := h.buildProgressResponse(r, studentUUID)
	if err != nil {
		log.Error().Err(err).Msg("failed to build progress response")
		http.Error(w, `{"error":"Kunde inte hämta progress"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetChildProgress handles GET /api/children/{studentId}/progress — parent views child.
func (h *ProgressHandler) GetChildProgress(w http.ResponseWriter, r *http.Request) {
	parentID := auth.GetUserIDFromContext(r.Context())
	if parentID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	studentIDStr := chi.URLParam(r, "studentId")
	studentUUID := parseUUID(studentIDStr)

	// Verify parent owns this student
	if !h.verifyParentOwnership(w, r, parentID, studentUUID) {
		return
	}

	resp, err := h.buildProgressResponse(r, studentUUID)
	if err != nil {
		log.Error().Err(err).Msg("failed to build child progress response")
		http.Error(w, `{"error":"Kunde inte hämta progress"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// ListChildSessions handles GET /api/children/{studentId}/progress/sessions — parent views session list.
func (h *ProgressHandler) ListChildSessions(w http.ResponseWriter, r *http.Request) {
	parentID := auth.GetUserIDFromContext(r.Context())
	if parentID == "" {
		http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	studentIDStr := chi.URLParam(r, "studentId")
	studentUUID := parseUUID(studentIDStr)

	// Verify parent owns this student
	if !h.verifyParentOwnership(w, r, parentID, studentUUID) {
		return
	}

	rows, err := h.store.ListCompletedSessions(r.Context(), studentUUID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list completed sessions")
		http.Error(w, `{"error":"Kunde inte hämta sessioner"}`, http.StatusInternalServerError)
		return
	}

	sessions := make([]sessionResponse, len(rows))
	for i, row := range rows {
		sessions[i] = sessionResponse{
			ID:            uuidToString(row.ID),
			Score:         row.Score.Int32,
			StartedAt:     row.StartedAt.Time.Format("2006-01-02T15:04:05Z"),
			EndedAt:       row.EndedAt.Time.Format("2006-01-02T15:04:05Z"),
			ExerciseTitle: row.ExerciseTitle,
			TopicName:     row.TopicName,
			SubjectName:   row.SubjectName,
		}
	}

	writeJSON(w, http.StatusOK, sessionsListResponse{Sessions: sessions})
}

// --- Internal helpers ---

func (h *ProgressHandler) buildProgressResponse(r *http.Request, studentUUID pgtype.UUID) (*progressResponse, error) {
	subjectRows, err := h.store.GetStudentProgressBySubject(r.Context(), studentUUID)
	if err != nil {
		return nil, fmt.Errorf("get subject progress: %w", err)
	}

	subjects := make([]subjectProgressResponse, len(subjectRows))
	for i, subj := range subjectRows {
		topicRows, err := h.store.GetStudentProgressByTopic(r.Context(), queries.GetStudentProgressByTopicParams{
			StudentID: studentUUID,
			SubjectID: subj.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("get topic progress for %s: %w", subj.Name, err)
		}

		topics := make([]topicProgressResponse, len(topicRows))
		for j, topic := range topicRows {
			topics[j] = topicProgressResponse{
				ID:              uuidToString(topic.ID),
				Name:            topic.Name,
				Slug:            topic.Slug,
				TotalSessions:   topic.TotalSessions,
				AvgScore:        topic.AvgScore,
				UniqueExercises: topic.UniqueExercises,
			}
		}

		subjects[i] = subjectProgressResponse{
			ID:              uuidToString(subj.ID),
			Name:            subj.Name,
			Slug:            subj.Slug,
			TotalSessions:   subj.TotalSessions,
			AvgScore:        subj.AvgScore,
			UniqueExercises: subj.UniqueExercises,
			Topics:          topics,
		}
	}

	return &progressResponse{Subjects: subjects}, nil
}

func (h *ProgressHandler) verifyParentOwnership(w http.ResponseWriter, r *http.Request, parentID string, studentUUID pgtype.UUID) bool {
	student, err := h.store.GetStudentByID(r.Context(), studentUUID)
	if err != nil {
		http.Error(w, `{"error":"Eleven hittades inte"}`, http.StatusNotFound)
		return false
	}

	parentUUID := parseUUID(parentID)
	if student.ParentID != parentUUID {
		http.Error(w, `{"error":"Du har inte behörighet att se denna profil"}`, http.StatusForbidden)
		return false
	}

	return true
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
