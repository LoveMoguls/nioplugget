package child

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/trollstaven/nioplugget/backend/internal/auth"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// UpdateStudentInviteParams holds the data for refreshing an invite code on an existing student.
type UpdateStudentInviteParams struct {
	ID              pgtype.UUID        `json:"id"`
	InviteCode      pgtype.Text        `json:"invite_code"`
	InviteExpiresAt pgtype.Timestamptz `json:"invite_expires_at"`
}

// ChildQuerier is the interface ChildHandler uses for all DB operations.
// Using an interface allows test mocking without a live database.
type ChildQuerier interface {
	CreateStudent(ctx context.Context, arg queries.CreateStudentParams) (queries.Student, error)
	GetStudentsByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.Student, error)
	ActivateStudent(ctx context.Context, arg queries.ActivateStudentParams) (queries.Student, error)
	GetStudentByNameAndParent(ctx context.Context, arg queries.GetStudentByNameAndParentParams) (queries.Student, error)
	GetParentByEmail(ctx context.Context, email string) (queries.Parent, error)
	ListStudentNamesByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.ListStudentNamesByParentIDRow, error)
	UpdateStudentInvite(ctx context.Context, arg UpdateStudentInviteParams) (queries.Student, error)
	GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error)
}

// ChildHandler holds the dependencies for child management HTTP handlers.
type ChildHandler struct {
	queries     ChildQuerier
	rateLimiter *PINRateLimiter
}

// NewChildHandler creates a ChildHandler with the given querier and PIN rate limiter.
func NewChildHandler(q ChildQuerier, rl *PINRateLimiter) *ChildHandler {
	return &ChildHandler{queries: q, rateLimiter: rl}
}

// pinRegex validates a 4-digit PIN.
var pinRegex = regexp.MustCompile(`^[0-9]{4}$`)

// frontendURL returns the base URL for invite links from env or a safe default.
func frontendURL() string {
	u := os.Getenv("FRONTEND_URL")
	if u == "" {
		return "http://localhost:5173"
	}
	return u
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// parseUUIDParam parses a UUID string and returns a pgtype.UUID.
func parseUUIDParam(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		return pgtype.UUID{}, err
	}
	return u, nil
}

// uuidToString converts pgtype.UUID to a hyphenated UUID string.
func uuidToString(id pgtype.UUID) string {
	if !id.Valid {
		return ""
	}
	b := id.Bytes
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// --- Create child ---

type createChildRequest struct {
	Name string `json:"name"`
}

type createChildResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	InviteURL string    `json:"inviteURL"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Create handles POST /api/children (parent-protected).
func (h *ChildHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createChildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt format"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Namn krävs"})
		return
	}

	parentID := auth.GetUserIDFromContext(r.Context())
	parentUUID, err := parseUUIDParam(parentID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Ogiltig session"})
		return
	}

	token, err := GenerateInviteToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	expiresAt := time.Now().Add(72 * time.Hour)
	student, err := h.queries.CreateStudent(r.Context(), queries.CreateStudentParams{
		ParentID: parentUUID,
		Name:     req.Name,
		InviteCode: pgtype.Text{
			String: token,
			Valid:  true,
		},
		InviteExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	writeJSON(w, http.StatusCreated, createChildResponse{
		ID:        uuidToString(student.ID),
		Name:      student.Name,
		InviteURL: frontendURL() + "/invite/" + token,
		ExpiresAt: expiresAt,
	})
}

// --- List children ---

type childListItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Activated bool      `json:"activated"`
	CreatedAt time.Time `json:"createdAt"`
}

// List handles GET /api/children (parent-protected).
func (h *ChildHandler) List(w http.ResponseWriter, r *http.Request) {
	parentID := auth.GetUserIDFromContext(r.Context())
	parentUUID, err := parseUUIDParam(parentID)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Ogiltig session"})
		return
	}

	students, err := h.queries.GetStudentsByParentID(r.Context(), parentUUID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	items := make([]childListItem, 0, len(students))
	for _, s := range students {
		items = append(items, childListItem{
			ID:        uuidToString(s.ID),
			Name:      s.Name,
			Activated: s.ActivatedAt.Valid,
			CreatedAt: s.CreatedAt.Time,
		})
	}
	writeJSON(w, http.StatusOK, items)
}

// --- Generate invite ---

type generateInviteResponse struct {
	InviteURL string    `json:"inviteURL"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// GenerateInvite handles POST /api/children/{id}/invite (parent-protected).
func (h *ChildHandler) GenerateInvite(w http.ResponseWriter, r *http.Request) {
	studentIDStr := chi.URLParam(r, "id")
	studentUUID, err := parseUUIDParam(studentIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt ID"})
		return
	}

	// Fetch student to verify ownership
	student, err := h.queries.GetStudentByID(r.Context(), studentUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Barnet hittades inte"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	// Verify ownership
	parentID := auth.GetUserIDFromContext(r.Context())
	if uuidToString(student.ParentID) != parentID {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Åtkomst nekad"})
		return
	}

	token, err := GenerateInviteToken()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	expiresAt := time.Now().Add(72 * time.Hour)
	_, err = h.queries.UpdateStudentInvite(r.Context(), UpdateStudentInviteParams{
		ID:         studentUUID,
		InviteCode: pgtype.Text{String: token, Valid: true},
		InviteExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	writeJSON(w, http.StatusOK, generateInviteResponse{
		InviteURL: frontendURL() + "/invite/" + token,
		ExpiresAt: expiresAt,
	})
}

// --- Activate invite ---

type activateRequest struct {
	Name string `json:"name"`
	PIN  string `json:"pin"`
}

// Activate handles POST /api/invite/{token}/activate (public).
func (h *ChildHandler) Activate(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	var req activateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt format"})
		return
	}

	// Validate PIN: exactly 4 digits
	if !pinRegex.MatchString(req.PIN) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "PIN måste vara exakt 4 siffror"})
		return
	}

	// Hash the PIN with Argon2id
	pinHash, err := auth.HashPassword(req.PIN)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	// Atomically activate the student
	student, err := h.queries.ActivateStudent(r.Context(), queries.ActivateStudentParams{
		InviteCode: pgtype.Text{String: token, Valid: true},
		PinHash:    pgtype.Text{String: pinHash, Valid: true},
	})
	if err != nil {
		// Token invalid, expired, or already used — return 410 Gone
		writeJSON(w, http.StatusGone, map[string]string{
			"error": "Länken har gått ut. Be din förälder skapa en ny.",
		})
		return
	}

	// Generate child JWT
	tokenStr, expiry, err := auth.GenerateToken(uuidToString(student.ID), "child")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	setAuthCookie(w, tokenStr, expiry)
	writeJSON(w, http.StatusOK, map[string]any{
		"id":      uuidToString(student.ID),
		"name":    student.Name,
		"message": "Kontot är aktiverat!",
	})
}

// --- PIN login ---

type pinLoginRequest struct {
	ParentEmail string `json:"parentEmail"`
	StudentName string `json:"studentName"`
	PIN         string `json:"pin"`
}

// PINLogin handles POST /api/child/login (public).
func (h *ChildHandler) PINLogin(w http.ResponseWriter, r *http.Request) {
	var req pinLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt format"})
		return
	}

	// Look up parent by email
	parent, err := h.queries.GetParentByEmail(r.Context(), req.ParentEmail)
	if err != nil {
		// Don't reveal which part is wrong
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Fel PIN-kod"})
		return
	}

	// Look up student by name + parent
	student, err := h.queries.GetStudentByNameAndParent(r.Context(), queries.GetStudentByNameAndParentParams{
		Name:     req.StudentName,
		ParentID: parent.ID,
	})
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Fel PIN-kod"})
		return
	}

	studentIDStr := uuidToString(student.ID)

	// Check rate limiter
	allowed, remaining := h.rateLimiter.Check(studentIDStr)
	if !allowed {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{
			"error": "För många försök. Försök igen om 15 minuter.",
		})
		return
	}

	// Compare PIN hash
	if !student.PinHash.Valid {
		// Account not activated
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Fel PIN-kod"})
		return
	}
	match, err := auth.ComparePassword(student.PinHash.String, req.PIN)
	if err != nil || !match {
		writeJSON(w, http.StatusUnauthorized, map[string]any{
			"error":     "Fel PIN-kod",
			"remaining": remaining,
		})
		return
	}

	// Successful login — reset rate limiter
	h.rateLimiter.Reset(studentIDStr)

	// Generate child JWT
	tokenStr, expiry, err := auth.GenerateToken(studentIDStr, "child")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	setAuthCookie(w, tokenStr, expiry)
	writeJSON(w, http.StatusOK, map[string]any{
		"id":   studentIDStr,
		"name": student.Name,
	})
}

// --- List names (public) ---

// ListNames handles GET /api/child/names?parent_email={email} (public).
func (h *ChildHandler) ListNames(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Query().Get("parent_email")
	if email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "parent_email krävs"})
		return
	}

	parent, err := h.queries.GetParentByEmail(r.Context(), email)
	if err != nil {
		// Return empty list — don't reveal whether parent exists
		writeJSON(w, http.StatusOK, []map[string]string{})
		return
	}

	rows, err := h.queries.ListStudentNamesByParentID(r.Context(), parent.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}

	type nameItem struct {
		Name string `json:"name"`
	}
	items := make([]nameItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, nameItem{Name: row.Name})
	}
	writeJSON(w, http.StatusOK, items)
}

// setAuthCookie sets an httpOnly JWT cookie.
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
