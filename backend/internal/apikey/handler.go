package apikey

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/trollstaven/nioplugget/backend/internal/auth"
)

// ErrNotFound is returned when an API key is not found for the given parent.
var ErrNotFound = errors.New("api key not found")

// APIKeyRecord represents a stored API key record.
type APIKeyRecord struct {
	EncryptedKey []byte
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// APIKeyStore is the interface the handler uses for persistence.
// It is implemented by the real queries.Queries (via adapter) and by test mocks.
type APIKeyStore interface {
	UpsertAPIKey(parentID pgtype.UUID, encKey []byte) (*APIKeyRecord, error)
	GetAPIKeyByParentID(parentID pgtype.UUID) (*APIKeyRecord, error)
	DeleteAPIKeyByParentID(parentID pgtype.UUID) error
}

// APIKeyHandler handles HTTP requests for API key management.
type APIKeyHandler struct {
	store       APIKeyStore
	encSvc      *EncryptionService
	validateURL string // overrides Anthropic URL in tests
}

// NewAPIKeyHandler creates a new APIKeyHandler.
// validateURL is the Anthropic models URL; pass "" to use the default.
func NewAPIKeyHandler(store APIKeyStore, encSvc *EncryptionService, validateURL string) *APIKeyHandler {
	return &APIKeyHandler{store: store, encSvc: encSvc, validateURL: validateURL}
}

// Store handles POST /api/apikey — validates, encrypts, and stores the API key.
func (h *APIKeyHandler) Store(w http.ResponseWriter, r *http.Request) {
	parentID, ok := h.parentUUID(w, r)
	if !ok {
		return
	}

	var req struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt JSON-format"})
		return
	}

	if err := h.validate(req.APIKey); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "API-nyckeln är inte giltig. Kontrollera att du kopierat rätt nyckel från console.anthropic.com",
		})
		return
	}

	encrypted, err := h.encSvc.Encrypt([]byte(req.APIKey))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Kunde inte kryptera API-nyckel"})
		return
	}

	if _, err := h.store.UpsertAPIKey(parentID, encrypted); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Kunde inte spara API-nyckel"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"maskedKey": MaskAPIKey(req.APIKey),
		"message":   "API-nyckel sparad",
	})
}

// Get handles GET /api/apikey — returns masked API key and timestamps.
func (h *APIKeyHandler) Get(w http.ResponseWriter, r *http.Request) {
	parentID, ok := h.parentUUID(w, r)
	if !ok {
		return
	}

	record, err := h.store.GetAPIKeyByParentID(parentID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Ingen API-nyckel sparad"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Kunde inte hämta API-nyckel"})
		return
	}

	rawKey, err := h.encSvc.Decrypt(record.EncryptedKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Kunde inte läsa API-nyckel"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"maskedKey": MaskAPIKey(string(rawKey)),
		"createdAt": record.CreatedAt,
		"updatedAt": record.UpdatedAt,
	})
}

// Update handles PUT /api/apikey — re-validates and re-encrypts the API key.
func (h *APIKeyHandler) Update(w http.ResponseWriter, r *http.Request) {
	parentID, ok := h.parentUUID(w, r)
	if !ok {
		return
	}

	var req struct {
		APIKey string `json:"apiKey"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt JSON-format"})
		return
	}

	if err := h.validate(req.APIKey); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "API-nyckeln är inte giltig. Kontrollera att du kopierat rätt nyckel från console.anthropic.com",
		})
		return
	}

	encrypted, err := h.encSvc.Encrypt([]byte(req.APIKey))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Kunde inte kryptera API-nyckel"})
		return
	}

	if _, err := h.store.UpsertAPIKey(parentID, encrypted); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Kunde inte uppdatera API-nyckel"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"maskedKey": MaskAPIKey(req.APIKey),
		"message":   "API-nyckel uppdaterad",
	})
}

// Delete handles DELETE /api/apikey — removes the stored API key.
func (h *APIKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	parentID, ok := h.parentUUID(w, r)
	if !ok {
		return
	}

	if err := h.store.DeleteAPIKeyByParentID(parentID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Kunde inte ta bort API-nyckel"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "API-nyckel borttagen"})
}

// parentUUID extracts and parses the parent's UUID from the JWT context.
func (h *APIKeyHandler) parentUUID(w http.ResponseWriter, r *http.Request) (pgtype.UUID, bool) {
	idStr := auth.GetUserIDFromContext(r.Context())
	if idStr == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Ej autentiserad"})
		return pgtype.UUID{}, false
	}

	var uid pgtype.UUID
	if err := uid.Scan(idStr); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt användar-ID"})
		return pgtype.UUID{}, false
	}
	return uid, true
}

// validate calls ValidateAPIKey or ValidateAPIKeyWithURL depending on configuration.
func (h *APIKeyHandler) validate(apiKey string) error {
	if h.validateURL != "" {
		return ValidateAPIKeyWithURL(apiKey, h.validateURL)
	}
	return ValidateAPIKey(apiKey)
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
