package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// ParentQuerier is the interface AuthHandler uses for DB operations.
// Using an interface allows test mocking without a live database.
type ParentQuerier interface {
	CreateParent(ctx context.Context, arg queries.CreateParentParams) (queries.Parent, error)
	GetParentByEmail(ctx context.Context, email string) (queries.Parent, error)
}

// AuthHandler holds the dependencies for auth HTTP handlers.
type AuthHandler struct {
	queries   ParentQuerier
	tokenAuth *jwtauth.JWTAuth
}

// NewAuthHandler creates an AuthHandler with the given querier and JWT auth.
func NewAuthHandler(q ParentQuerier, ta *jwtauth.JWTAuth) *AuthHandler {
	return &AuthHandler{queries: q, tokenAuth: ta}
}

// registerRequest is the expected JSON body for POST /api/auth/register.
type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	GdprConsent bool   `json:"gdprConsent"`
}

// loginRequest is the expected JSON body for POST /api/auth/login.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// authResponse is the JSON response for successful register/login.
type authResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func setAuthCookie(w http.ResponseWriter, tokenStr string, expiry time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    tokenStr,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   86400,
		Expires:  expiry,
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	})
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt format"})
		return
	}

	// Validate input
	if req.Email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "E-post krävs"})
		return
	}
	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Lösenordet måste vara minst 8 tecken"})
		return
	}
	if !req.GdprConsent {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "GDPR-samtycke krävs"})
		return
	}

	// Hash password
	hash, err := HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	// Create parent in database
	now := time.Now()
	parent, err := h.queries.CreateParent(r.Context(), queries.CreateParentParams{
		Email:        req.Email,
		PasswordHash: hash,
		GdprConsentAt: pgtype.Timestamptz{
			Time:  now,
			Valid: true,
		},
	})
	if err != nil {
		// Check for unique constraint violation (duplicate email)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeJSON(w, http.StatusConflict, map[string]string{"error": "E-postadressen är redan registrerad"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	// Generate JWT
	tokenStr, expiry, err := GenerateToken(uuidToString(parent.ID), "parent")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	setAuthCookie(w, tokenStr, expiry)
	writeJSON(w, http.StatusCreated, authResponse{
		ID:    uuidToString(parent.ID),
		Email: parent.Email,
	})
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt format"})
		return
	}

	// Look up parent by email — return same error for not-found and wrong password
	// to prevent user enumeration
	parent, err := h.queries.GetParentByEmail(r.Context(), req.Email)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Fel e-post eller lösenord"})
		return
	}

	// Verify password
	match, err := ComparePassword(parent.PasswordHash, req.Password)
	if err != nil || !match {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Fel e-post eller lösenord"})
		return
	}

	// Generate JWT
	tokenStr, expiry, err := GenerateToken(uuidToString(parent.ID), "parent")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	setAuthCookie(w, tokenStr, expiry)
	writeJSON(w, http.StatusOK, authResponse{
		ID:    uuidToString(parent.ID),
		Email: parent.Email,
	})
}

// Logout handles POST /api/auth/logout.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"message": "Utloggad"})
}

// uuidToString converts pgtype.UUID to string.
func uuidToString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	b := id.Bytes
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
