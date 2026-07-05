package device

import (
	"encoding/json"
	"net/http"
	"strings"

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

// requireDevice validates the device cookie: parses the JWT, checks role
// "device" and that the epoch matches the current family settings.
func (h *Handler) requireDevice(w http.ResponseWriter, r *http.Request) bool {
	tokenStr := TokenFromDeviceCookie(r)
	if tokenStr == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Enheten är inte upplåst"})
		return false
	}
	tok, err := auth.TokenAuth.Decode(tokenStr)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Enheten är inte upplåst"})
		return false
	}
	var role string
	if err := tok.Get("role", &role); err != nil || role != "device" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Enheten är inte upplåst"})
		return false
	}
	var epoch float64
	if err := tok.Get("epoch", &epoch); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Enheten är inte upplåst"})
		return false
	}
	settings, err := h.store.GetFamilySettings(r.Context())
	if err != nil || int32(epoch) != settings.DeviceEpoch {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Enheten är inte upplåst"})
		return false
	}
	return true
}

type profileResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// Profiles handles GET /api/profiles (trusted devices only).
func (h *Handler) Profiles(w http.ResponseWriter, r *http.Request) {
	if !h.requireDevice(w, r) {
		return
	}
	parents, err := h.store.ListParents(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to list parents")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}
	students, err := h.store.ListAllStudents(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to list students")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}
	profiles := make([]profileResponse, 0, len(parents)+len(students))
	for _, p := range parents {
		name := p.Email
		if i := strings.Index(name, "@"); i > 0 {
			name = name[:i]
		}
		profiles = append(profiles, profileResponse{ID: uuidToString(p.ID), Name: name, Role: "parent"})
	}
	for _, s := range students {
		profiles = append(profiles, profileResponse{ID: uuidToString(s.ID), Name: s.Name, Role: "child"})
	}
	writeJSON(w, http.StatusOK, profiles)
}

// ProfileLogin handles POST /api/profile/login (trusted devices only).
func (h *Handler) ProfileLogin(w http.ResponseWriter, r *http.Request) {
	if !h.requireDevice(w, r) {
		return
	}
	var req struct {
		ID   string `json:"id"`
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" || (req.Role != "parent" && req.Role != "child") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt format"})
		return
	}
	id := parseUUID(req.ID)
	var name string
	if req.Role == "parent" {
		p, err := h.store.GetParentByID(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Profilen hittades inte"})
			return
		}
		name = p.Email
	} else {
		s, err := h.store.GetStudentByID(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Profilen hittades inte"})
			return
		}
		name = s.Name
	}
	tokenStr, expiry, err := auth.GenerateToken(req.ID, req.Role)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate token for profile login")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}
	auth.SetAuthCookie(w, tokenStr, expiry)
	writeJSON(w, http.StatusOK, map[string]string{"role": req.Role, "name": name})
}

// SetCode handles POST /api/device/set-code (parent JWT required — mounted
// behind ParentOnly in the router). Changing an existing code requires the
// current code and bumps the device epoch (logs out all devices).
func (h *Handler) SetCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NewCode     string `json:"newCode"`
		CurrentCode string `json:"currentCode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt format"})
		return
	}
	if len(req.NewCode) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Koden måste vara minst 6 tecken"})
		return
	}
	if settings, err := h.store.GetFamilySettings(r.Context()); err == nil {
		match, err := auth.ComparePassword(settings.CodeHash, req.CurrentCode)
		if err != nil || !match {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Fel nuvarande familjekod"})
			return
		}
	}
	hash, err := auth.HashPassword(req.NewCode)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}
	if _, err := h.store.UpsertFamilyCode(r.Context(), hash); err != nil {
		log.Error().Err(err).Msg("failed to save family code")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Kunde inte spara koden"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Familjekoden är sparad"})
}
