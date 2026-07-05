package device

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/auth"
	"github.com/trollstaven/nioplugget/backend/internal/child"
)

// Handler serves device unlock and profile endpoints.
type Handler struct {
	store   Store
	limiter *child.PINRateLimiter
}

func NewHandler(store Store, limiter *child.PINRateLimiter) *Handler {
	return &Handler{store: store, limiter: limiter}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data) //nolint:errcheck
}

// Unlock handles POST /api/device/unlock — trades the family code for a
// 365-day trusted-device cookie.
func (h *Handler) Unlock(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if allowed, _ := h.limiter.Check(ip); !allowed {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "För många försök. Vänta 15 minuter."})
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt format"})
		return
	}

	settings, err := h.store.GetFamilySettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Ingen familjekod är satt ännu"})
		return
	}

	match, err := auth.ComparePassword(settings.CodeHash, req.Code)
	if err != nil || !match {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Fel familjekod"})
		return
	}
	h.limiter.Reset(ip)

	tokenStr, expiry, err := GenerateDeviceToken(settings.DeviceEpoch)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate device token")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     deviceCookieName,
		Value:    tokenStr,
		Path:     "/",
		Expires:  expiry,
		HttpOnly: true,
		Secure:   auth.IsSecureCookie(),
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Enheten är upplåst"})
}
