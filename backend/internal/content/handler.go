package content

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// ContentHandler provides HTTP handlers for content browsing.
type ContentHandler struct {
	store ContentStore
}

// NewContentHandler creates a new ContentHandler.
func NewContentHandler(store ContentStore) *ContentHandler {
	return &ContentHandler{store: store}
}

type subjectResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	DisplayOrder int32  `json:"displayOrder"`
}

type topicResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	DisplayOrder int32  `json:"displayOrder"`
}

type exerciseResponse struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	DifficultyOrder int32  `json:"difficultyOrder"`
}

// ListSubjects handles GET /api/subjects
func (h *ContentHandler) ListSubjects(w http.ResponseWriter, r *http.Request) {
	subjects, err := h.store.ListSubjects(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to list subjects")
		http.Error(w, `{"error":"Kunde inte hämta ämnen"}`, http.StatusInternalServerError)
		return
	}

	result := make([]subjectResponse, len(subjects))
	for i, s := range subjects {
		result[i] = subjectResponse{
			ID:           uuidToString(s.ID),
			Name:         s.Name,
			Slug:         s.Slug,
			DisplayOrder: s.DisplayOrder,
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// ListTopics handles GET /api/subjects/{subjectSlug}/topics
func (h *ContentHandler) ListTopics(w http.ResponseWriter, r *http.Request) {
	subjectSlug := chi.URLParam(r, "subjectSlug")

	subject, err := h.store.GetSubjectBySlug(r.Context(), subjectSlug)
	if err != nil {
		http.Error(w, `{"error":"Ämnet hittades inte"}`, http.StatusNotFound)
		return
	}

	topics, err := h.store.ListTopicsBySubjectID(r.Context(), subject.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list topics")
		http.Error(w, `{"error":"Kunde inte hämta områden"}`, http.StatusInternalServerError)
		return
	}

	topicList := make([]topicResponse, len(topics))
	for i, t := range topics {
		topicList[i] = topicResponse{
			ID:           uuidToString(t.ID),
			Name:         t.Name,
			Slug:         t.Slug,
			DisplayOrder: t.DisplayOrder,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"subject": subjectResponse{
			ID:   uuidToString(subject.ID),
			Name: subject.Name,
			Slug: subject.Slug,
		},
		"topics": topicList,
	})
}

// ListExercises handles GET /api/topics/{subjectSlug}/{topicSlug}/exercises
func (h *ContentHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
	subjectSlug := chi.URLParam(r, "subjectSlug")
	topicSlug := chi.URLParam(r, "topicSlug")

	topic, err := h.store.GetTopicBySlug(r.Context(), queries.GetTopicBySlugParams{
		SubjectSlug: subjectSlug,
		TopicSlug:   topicSlug,
	})
	if err != nil {
		http.Error(w, `{"error":"Området hittades inte"}`, http.StatusNotFound)
		return
	}

	subject, err := h.store.GetSubjectBySlug(r.Context(), subjectSlug)
	if err != nil {
		http.Error(w, `{"error":"Ämnet hittades inte"}`, http.StatusNotFound)
		return
	}

	exercises, err := h.store.ListExercisesByTopicID(r.Context(), topic.ID)
	if err != nil {
		log.Error().Err(err).Msg("failed to list exercises")
		http.Error(w, `{"error":"Kunde inte hämta övningar"}`, http.StatusInternalServerError)
		return
	}

	exerciseList := make([]exerciseResponse, len(exercises))
	for i, e := range exercises {
		exerciseList[i] = exerciseResponse{
			ID:              uuidToString(e.ID),
			Title:           e.Title,
			Description:     e.Description,
			DifficultyOrder: e.DifficultyOrder,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"subject": subjectResponse{
			ID:   uuidToString(subject.ID),
			Name: subject.Name,
			Slug: subject.Slug,
		},
		"topic": topicResponse{
			ID:   uuidToString(topic.ID),
			Name: topic.Name,
			Slug: topic.Slug,
		},
		"exercises": exerciseList,
	})
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
