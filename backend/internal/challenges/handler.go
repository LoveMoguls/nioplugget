// backend/internal/challenges/handler.go
package challenges

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/apikey"
	"github.com/trollstaven/nioplugget/backend/internal/auth"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// ChallengeHandler provides HTTP handlers for challenges.
type ChallengeHandler struct {
	store  ChallengeStore
	encSvc *apikey.EncryptionService
}

func NewChallengeHandler(store ChallengeStore, encSvc *apikey.EncryptionService) *ChallengeHandler {
	return &ChallengeHandler{store: store, encSvc: encSvc}
}

// POST /api/challenges — multipart/form-data with field "images" (1-6 files)
func (h *ChallengeHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	role := auth.GetRoleFromContext(r.Context())

	parentID, createdByRole, err := resolveParentID(r.Context(), h.store, userID, role)
	if err != nil {
		http.Error(w, `{"error":"Kunde inte identifiera användaren"}`, http.StatusUnauthorized)
		return
	}

	apiKeyRecord, err := h.store.GetAPIKeyByParentID(r.Context(), parentID)
	if err != nil {
		http.Error(w, `{"error":"API-nyckel saknas. Föräldern måste lägga till en API-nyckel i inställningarna."}`, http.StatusBadRequest)
		return
	}
	decryptedKey, err := h.encSvc.Decrypt(apiKeyRecord.EncryptedKey)
	if err != nil {
		http.Error(w, `{"error":"Kunde inte läsa API-nyckeln"}`, http.StatusInternalServerError)
		return
	}

	if err := r.ParseMultipartForm(30 << 20); err != nil {
		http.Error(w, `{"error":"Kunde inte läsa filerna"}`, http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["images"]
	if len(files) == 0 || len(files) > 6 {
		http.Error(w, `{"error":"Ladda upp 1-6 bilder"}`, http.StatusBadRequest)
		return
	}

	var imageDataList [][]byte
	var mediaTypes []string
	for _, fh := range files {
		if fh.Size > 5<<20 {
			http.Error(w, `{"error":"Varje bild får max vara 5 MB"}`, http.StatusBadRequest)
			return
		}
		f, err := fh.Open()
		if err != nil {
			http.Error(w, `{"error":"Kunde inte öppna filen"}`, http.StatusInternalServerError)
			return
		}
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			http.Error(w, `{"error":"Kunde inte läsa filen"}`, http.StatusInternalServerError)
			return
		}
		ct := fh.Header.Get("Content-Type")
		if ct == "" {
			ct = "image/jpeg"
		}
		imageDataList = append(imageDataList, data)
		mediaTypes = append(mediaTypes, ct)
	}

	generated, err := GenerateChallenge(r.Context(), string(decryptedKey), imageDataList, mediaTypes)
	if err != nil {
		log.Error().Err(err).Msg("challenge generation failed")
		http.Error(w, `{"error":"Kunde inte läsa bilderna. Försök med tydligare foton."}`, http.StatusUnprocessableEntity)
		return
	}

	challenge, err := h.store.CreateChallenge(r.Context(), queries.CreateChallengeParams{
		ParentID:      parentID,
		CreatedByRole: createdByRole,
		Title:         generated.Title,
		Description:   generated.Description,
		CoverEmoji:    generated.Emoji,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to save challenge")
		http.Error(w, `{"error":"Kunde inte spara utmaningen"}`, http.StatusInternalServerError)
		return
	}

	var exerciseResponses []map[string]interface{}
	for i, ex := range generated.Exercises {
		saved, err := h.store.CreateChallengeExercise(r.Context(), queries.CreateChallengeExerciseParams{
			ChallengeID:  challenge.ID,
			Title:        ex.Title,
			Description:  ex.Description,
			SystemPrompt: ex.SystemPrompt,
			DisplayOrder: int32(i + 1),
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to save challenge exercise")
			// Don't leave a half-saved challenge behind — remove it and fail the request
			if delErr := h.store.DeleteChallenge(r.Context(), queries.DeleteChallengeParams{
				ID:       challenge.ID,
				ParentID: parentID,
			}); delErr != nil {
				log.Error().Err(delErr).Msg("failed to clean up partial challenge")
			}
			http.Error(w, `{"error":"Kunde inte spara utmaningen"}`, http.StatusInternalServerError)
			return
		}
		exerciseResponses = append(exerciseResponses, map[string]interface{}{
			"id":           uuidToString(saved.ID),
			"title":        saved.Title,
			"description":  saved.Description,
			"displayOrder": saved.DisplayOrder,
		})
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":          uuidToString(challenge.ID),
		"title":       challenge.Title,
		"description": challenge.Description,
		"coverEmoji":  challenge.CoverEmoji,
		"exercises":   exerciseResponses,
		"createdAt":   challenge.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	})
}

// GET /api/challenges
func (h *ChallengeHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserIDFromContext(r.Context())
	role := auth.GetRoleFromContext(r.Context())

	parentID, _, err := resolveParentID(r.Context(), h.store, userID, role)
	if err != nil {
		http.Error(w, `{"error":"Kunde inte identifiera användaren"}`, http.StatusUnauthorized)
		return
	}

	var challengeList []queries.Challenge
	if role == "parent" {
		challengeList, err = h.store.ListChallengesByParentID(r.Context(), parentID)
	} else {
		challengeList, err = h.store.ListPublishedChallengesByParentID(r.Context(), parentID)
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to list challenges")
		http.Error(w, `{"error":"Kunde inte hämta utmaningar"}`, http.StatusInternalServerError)
		return
	}

	result := make([]map[string]interface{}, len(challengeList))
	for i, c := range challengeList {
		result[i] = map[string]interface{}{
			"id":          uuidToString(c.ID),
			"title":       c.Title,
			"description": c.Description,
			"coverEmoji":  c.CoverEmoji,
			"published":   c.Published,
			"createdAt":   c.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}
	}
	writeJSON(w, http.StatusOK, result)
}

// PATCH /api/challenges/{id}/publish — parent only
func (h *ChallengeHandler) Publish(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := auth.GetUserIDFromContext(r.Context())

	var parentID pgtype.UUID
	if err := parentID.Scan(userID); err != nil {
		http.Error(w, `{"error":"Ogiltigt användar-ID"}`, http.StatusBadRequest)
		return
	}

	var body struct {
		Title string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		http.Error(w, `{"error":"Titel krävs"}`, http.StatusBadRequest)
		return
	}

	updated, err := h.store.PublishChallenge(r.Context(), queries.PublishChallengeParams{
		Title:    body.Title,
		ID:       parseUUID(id),
		ParentID: parentID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error":"Utmaningen hittades inte"}`, http.StatusNotFound)
			return
		}
		log.Error().Err(err).Msg("failed to publish challenge")
		http.Error(w, `{"error":"Kunde inte publicera utmaningen"}`, http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":          uuidToString(updated.ID),
		"title":       updated.Title,
		"description": updated.Description,
		"coverEmoji":  updated.CoverEmoji,
		"published":   updated.Published,
		"createdAt":   updated.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
	})
}

// GET /api/challenges/{id}
func (h *ChallengeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := auth.GetUserIDFromContext(r.Context())
	role := auth.GetRoleFromContext(r.Context())

	challengeUUID := parseUUID(id)
	challenge, err := h.store.GetChallengeByID(r.Context(), challengeUUID)
	if err != nil {
		http.Error(w, `{"error":"Utmaningen hittades inte"}`, http.StatusNotFound)
		return
	}

	// Only members of the owning family may see a challenge; children only published ones
	callerParentID, _, err := resolveParentID(r.Context(), h.store, userID, role)
	if err != nil {
		http.Error(w, `{"error":"Kunde inte identifiera användaren"}`, http.StatusUnauthorized)
		return
	}
	if challenge.ParentID != callerParentID || (role == "child" && !challenge.Published) {
		http.Error(w, `{"error":"Utmaningen hittades inte"}`, http.StatusNotFound)
		return
	}

	// If child, get exercises with their progress
	if role == "child" {
		var studentID pgtype.UUID
		if err := studentID.Scan(userID); err != nil {
			http.Error(w, `{"error":"Ogiltigt användar-ID"}`, http.StatusBadRequest)
			return
		}
		rows, err := h.store.ListChallengeExercisesWithProgress(r.Context(), queries.ListChallengeExercisesWithProgressParams{
			ChallengeID: challengeUUID,
			StudentID:   studentID,
		})
		if err != nil {
			log.Error().Err(err).Msg("failed to get challenge progress")
			http.Error(w, `{"error":"Kunde inte hämta övningar"}`, http.StatusInternalServerError)
			return
		}
		exercises := make([]map[string]interface{}, len(rows))
		for i, row := range rows {
			ex := map[string]interface{}{
				"id":           uuidToString(row.ID),
				"title":        row.Title,
				"description":  row.Description,
				"displayOrder": row.DisplayOrder,
				"completed":    row.SessionID.Valid,
				"stars":        0,
				"xp":           0,
			}
			if row.SessionID.Valid {
				ex["sessionId"] = uuidToString(row.SessionID)
				if row.Score.Valid {
					stars := scoreToStars(int(row.Score.Int32))
					ex["stars"] = stars
					ex["xp"] = stars * 10
				}
			}
			exercises[i] = ex
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"id":          uuidToString(challenge.ID),
			"title":       challenge.Title,
			"description": challenge.Description,
			"coverEmoji":  challenge.CoverEmoji,
			"exercises":   exercises,
		})
		return
	}

	// Parent: just list exercises without progress
	exercises, err := h.store.ListChallengeExercisesByChallengeID(r.Context(), challengeUUID)
	if err != nil {
		http.Error(w, `{"error":"Kunde inte hämta övningar"}`, http.StatusInternalServerError)
		return
	}
	exList := make([]map[string]interface{}, len(exercises))
	for i, e := range exercises {
		exList[i] = map[string]interface{}{
			"id":           uuidToString(e.ID),
			"title":        e.Title,
			"description":  e.Description,
			"displayOrder": e.DisplayOrder,
		}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":          uuidToString(challenge.ID),
		"title":       challenge.Title,
		"description": challenge.Description,
		"coverEmoji":  challenge.CoverEmoji,
		"exercises":   exList,
	})
}

// DELETE /api/challenges/{id} — parent only
func (h *ChallengeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := auth.GetUserIDFromContext(r.Context())

	var parentID pgtype.UUID
	if err := parentID.Scan(userID); err != nil {
		http.Error(w, `{"error":"Ogiltigt användar-ID"}`, http.StatusBadRequest)
		return
	}

	challengeUUID := parseUUID(id)
	if err := h.store.DeleteChallenge(r.Context(), queries.DeleteChallengeParams{
		ID:       challengeUUID,
		ParentID: parentID,
	}); err != nil {
		log.Error().Err(err).Msg("failed to delete challenge")
		http.Error(w, `{"error":"Kunde inte ta bort utmaningen"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func scoreToStars(score int) int {
	switch {
	case score >= 5:
		return 3
	case score >= 3:
		return 2
	default:
		return 1
	}
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
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func parseUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	u.Scan(s)
	return u
}
