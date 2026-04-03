package child

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// --- Mock querier ---

type mockQuerier struct {
	createStudentFn           func(ctx context.Context, arg queries.CreateStudentParams) (queries.Student, error)
	getStudentsByParentIDFn   func(ctx context.Context, parentID pgtype.UUID) ([]queries.Student, error)
	activateStudentFn         func(ctx context.Context, arg queries.ActivateStudentParams) (queries.Student, error)
	getStudentByNameAndParentFn func(ctx context.Context, arg queries.GetStudentByNameAndParentParams) (queries.Student, error)
	getParentByEmailFn        func(ctx context.Context, email string) (queries.Parent, error)
	listStudentNamesByParentIDFn func(ctx context.Context, parentID pgtype.UUID) ([]queries.ListStudentNamesByParentIDRow, error)
	updateStudentInviteFn     func(ctx context.Context, arg UpdateStudentInviteParams) (queries.Student, error)
	getStudentByIDFn          func(ctx context.Context, id pgtype.UUID) (queries.Student, error)
}

func (m *mockQuerier) CreateStudent(ctx context.Context, arg queries.CreateStudentParams) (queries.Student, error) {
	return m.createStudentFn(ctx, arg)
}
func (m *mockQuerier) GetStudentsByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.Student, error) {
	return m.getStudentsByParentIDFn(ctx, parentID)
}
func (m *mockQuerier) ActivateStudent(ctx context.Context, arg queries.ActivateStudentParams) (queries.Student, error) {
	return m.activateStudentFn(ctx, arg)
}
func (m *mockQuerier) GetStudentByNameAndParent(ctx context.Context, arg queries.GetStudentByNameAndParentParams) (queries.Student, error) {
	return m.getStudentByNameAndParentFn(ctx, arg)
}
func (m *mockQuerier) GetParentByEmail(ctx context.Context, email string) (queries.Parent, error) {
	return m.getParentByEmailFn(ctx, email)
}
func (m *mockQuerier) ListStudentNamesByParentID(ctx context.Context, parentID pgtype.UUID) ([]queries.ListStudentNamesByParentIDRow, error) {
	return m.listStudentNamesByParentIDFn(ctx, parentID)
}
func (m *mockQuerier) UpdateStudentInvite(ctx context.Context, arg UpdateStudentInviteParams) (queries.Student, error) {
	return m.updateStudentInviteFn(ctx, arg)
}
func (m *mockQuerier) GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error) {
	return m.getStudentByIDFn(ctx, id)
}

// --- Test helpers ---

const testParentUUID = "550e8400-e29b-41d4-a716-446655440000"
const testStudentUUID = "660e8400-e29b-41d4-a716-446655440000"
const testOtherParentUUID = "770e8400-e29b-41d4-a716-446655440000"

func parseUUID(s string) pgtype.UUID {
	var u pgtype.UUID
	_ = u.Scan(s)
	return u
}

func injectParentJWT(r *http.Request, parentID string) *http.Request {
	ta := jwtauth.New("HS256", []byte("test-secret"), nil)
	_, tokenStr, _ := ta.Encode(map[string]any{"sub": parentID, "role": "parent"})
	token, _ := ta.Decode(tokenStr)
	ctx := jwtauth.NewContext(r.Context(), token, nil)
	return r.WithContext(ctx)
}

func makeStudent(name, parentIDStr, studentIDStr string, activated bool) queries.Student {
	s := queries.Student{
		ID:       parseUUID(studentIDStr),
		ParentID: parseUUID(parentIDStr),
		Name:     name,
		InviteCode: pgtype.Text{String: "test-invite-code", Valid: true},
		InviteExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(72 * time.Hour), Valid: true},
	}
	if activated {
		s.PinHash = pgtype.Text{String: "$argon2id$hashed", Valid: true}
		s.ActivatedAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		s.InviteCode = pgtype.Text{Valid: false}
		s.InviteExpiresAt = pgtype.Timestamptz{Valid: false}
	}
	return s
}

func newTestHandler(q ChildQuerier) *ChildHandler {
	rl := NewPINRateLimiter(5, 15*time.Minute)
	return NewChildHandler(q, rl)
}

// --- Tests: Create child ---

func TestHandlerCreateChildSuccess(t *testing.T) {
	q := &mockQuerier{
		createStudentFn: func(ctx context.Context, arg queries.CreateStudentParams) (queries.Student, error) {
			return makeStudent("Erik", testParentUUID, testStudentUUID, false), nil
		},
	}
	h := newTestHandler(q)

	body := `{"name": "Erik"}`
	r := httptest.NewRequest(http.MethodPost, "/api/children", bytes.NewBufferString(body))
	r = injectParentJWT(r, testParentUUID)
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["id"] == "" || resp["id"] == nil {
		t.Error("response missing id")
	}
	if resp["name"] == "" || resp["name"] == nil {
		t.Error("response missing name")
	}
	if resp["inviteURL"] == "" || resp["inviteURL"] == nil {
		t.Error("response missing inviteURL")
	}
}

func TestHandlerCreateChildEmptyName(t *testing.T) {
	h := newTestHandler(&mockQuerier{})

	body := `{"name": ""}`
	r := httptest.NewRequest(http.MethodPost, "/api/children", bytes.NewBufferString(body))
	r = injectParentJWT(r, testParentUUID)
	w := httptest.NewRecorder()

	h.Create(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// --- Tests: List children ---

func TestHandlerListChildrenSuccess(t *testing.T) {
	q := &mockQuerier{
		getStudentsByParentIDFn: func(ctx context.Context, parentID pgtype.UUID) ([]queries.Student, error) {
			return []queries.Student{
				makeStudent("Erik", testParentUUID, testStudentUUID, false),
				makeStudent("Anna", testParentUUID, "880e8400-e29b-41d4-a716-446655440000", true),
			}, nil
		},
	}
	h := newTestHandler(q)

	r := httptest.NewRequest(http.MethodGet, "/api/children", nil)
	r = injectParentJWT(r, testParentUUID)
	w := httptest.NewRecorder()

	h.List(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp []map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Errorf("got %d children, want 2", len(resp))
	}
}

// --- Tests: Generate invite ---

func TestHandlerGenerateInviteSuccess(t *testing.T) {
	q := &mockQuerier{
		getStudentByIDFn: func(ctx context.Context, id pgtype.UUID) (queries.Student, error) {
			return makeStudent("Erik", testParentUUID, testStudentUUID, false), nil
		},
		updateStudentInviteFn: func(ctx context.Context, arg UpdateStudentInviteParams) (queries.Student, error) {
			s := makeStudent("Erik", testParentUUID, testStudentUUID, false)
			s.InviteCode = pgtype.Text{String: "new-invite-code", Valid: true}
			return s, nil
		},
	}
	h := newTestHandler(q)

	r := httptest.NewRequest(http.MethodPost, "/api/children/"+testStudentUUID+"/invite", nil)
	r = injectParentJWT(r, testParentUUID)

	// Inject chi URL param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testStudentUUID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.GenerateInvite(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["inviteURL"] == "" || resp["inviteURL"] == nil {
		t.Error("response missing inviteURL")
	}
}

func TestHandlerGenerateInviteForbiddenForOtherParent(t *testing.T) {
	q := &mockQuerier{
		getStudentByIDFn: func(ctx context.Context, id pgtype.UUID) (queries.Student, error) {
			// Belongs to testParentUUID
			return makeStudent("Erik", testParentUUID, testStudentUUID, false), nil
		},
	}
	h := newTestHandler(q)

	r := httptest.NewRequest(http.MethodPost, "/api/children/"+testStudentUUID+"/invite", nil)
	// Authenticated as a different parent
	r = injectParentJWT(r, testOtherParentUUID)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", testStudentUUID)
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.GenerateInvite(w, r)

	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

// --- Tests: Activate invite ---

func TestHandlerActivateSuccess(t *testing.T) {
	q := &mockQuerier{
		activateStudentFn: func(ctx context.Context, arg queries.ActivateStudentParams) (queries.Student, error) {
			return makeStudent("Erik", testParentUUID, testStudentUUID, true), nil
		},
	}
	h := newTestHandler(q)

	body := `{"name": "Erik", "pin": "1234"}`
	r := httptest.NewRequest(http.MethodPost, "/api/invite/abc123/activate", bytes.NewBufferString(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "abc123")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.Activate(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	// Check JWT cookie set
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "token" {
			found = true
		}
	}
	if !found {
		t.Error("expected token cookie to be set")
	}
}

func TestHandlerActivateInvalidPIN(t *testing.T) {
	h := newTestHandler(&mockQuerier{})

	testCases := []struct {
		pin string
	}{
		{"123"},   // too short
		{"12345"}, // too long
		{"abcd"},  // non-digits
		{""},      // empty
	}

	for _, tc := range testCases {
		body, _ := json.Marshal(map[string]string{"name": "Erik", "pin": tc.pin})
		r := httptest.NewRequest(http.MethodPost, "/api/invite/abc123/activate", bytes.NewBuffer(body))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("token", "abc123")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

		w := httptest.NewRecorder()
		h.Activate(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("pin=%q: status = %d, want 400", tc.pin, w.Code)
		}
	}
}

func TestHandlerActivateExpiredToken(t *testing.T) {
	q := &mockQuerier{
		activateStudentFn: func(ctx context.Context, arg queries.ActivateStudentParams) (queries.Student, error) {
			return queries.Student{}, errors.New("no rows")
		},
	}
	h := newTestHandler(q)

	body := `{"name": "Erik", "pin": "1234"}`
	r := httptest.NewRequest(http.MethodPost, "/api/invite/expired-token/activate", bytes.NewBufferString(body))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("token", "expired-token")
	r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	h.Activate(w, r)

	if w.Code != http.StatusGone {
		t.Errorf("status = %d, want 410", w.Code)
	}
}

// --- Tests: PIN Login ---

func TestHandlerPINLoginSuccess(t *testing.T) {
	q := &mockQuerier{
		getParentByEmailFn: func(ctx context.Context, email string) (queries.Parent, error) {
			return queries.Parent{
				ID:    parseUUID(testParentUUID),
				Email: "parent@test.com",
			}, nil
		},
		getStudentByNameAndParentFn: func(ctx context.Context, arg queries.GetStudentByNameAndParentParams) (queries.Student, error) {
			// Return student with argon2id hash of "1234"
			s := makeStudent("Erik", testParentUUID, testStudentUUID, true)
			// Use a real hash for "1234"
			s.PinHash = pgtype.Text{String: mustHashPIN("1234"), Valid: true}
			return s, nil
		},
	}
	h := newTestHandler(q)

	body := `{"parentEmail": "parent@test.com", "studentName": "Erik", "pin": "1234"}`
	r := httptest.NewRequest(http.MethodPost, "/api/child/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.PINLogin(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (body: %s)", w.Code, w.Body.String())
	}
	// Check JWT cookie
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "token" {
			found = true
		}
	}
	if !found {
		t.Error("expected token cookie to be set on successful login")
	}
}

func TestHandlerPINLoginWrongPIN(t *testing.T) {
	q := &mockQuerier{
		getParentByEmailFn: func(ctx context.Context, email string) (queries.Parent, error) {
			return queries.Parent{ID: parseUUID(testParentUUID), Email: email}, nil
		},
		getStudentByNameAndParentFn: func(ctx context.Context, arg queries.GetStudentByNameAndParentParams) (queries.Student, error) {
			s := makeStudent("Erik", testParentUUID, testStudentUUID, true)
			s.PinHash = pgtype.Text{String: mustHashPIN("1234"), Valid: true}
			return s, nil
		},
	}
	h := newTestHandler(q)

	body := `{"parentEmail": "parent@test.com", "studentName": "Erik", "pin": "9999"}`
	r := httptest.NewRequest(http.MethodPost, "/api/child/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.PINLogin(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if _, ok := resp["remaining"]; !ok {
		t.Error("response missing 'remaining' field")
	}
}

func TestHandlerPINLoginLocked(t *testing.T) {
	q := &mockQuerier{
		getParentByEmailFn: func(ctx context.Context, email string) (queries.Parent, error) {
			return queries.Parent{ID: parseUUID(testParentUUID), Email: email}, nil
		},
		getStudentByNameAndParentFn: func(ctx context.Context, arg queries.GetStudentByNameAndParentParams) (queries.Student, error) {
			s := makeStudent("Erik", testParentUUID, testStudentUUID, true)
			s.PinHash = pgtype.Text{String: mustHashPIN("1234"), Valid: true}
			return s, nil
		},
	}
	rl := NewPINRateLimiter(5, 15*time.Minute)
	h := NewChildHandler(q, rl)

	// Trigger lockout by making 5 wrong attempts
	studentUUID := testStudentUUID
	for i := 0; i < 5; i++ {
		rl.Check(studentUUID)
	}

	body := `{"parentEmail": "parent@test.com", "studentName": "Erik", "pin": "9999"}`
	r := httptest.NewRequest(http.MethodPost, "/api/child/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.PINLogin(w, r)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", w.Code)
	}
}

// --- Tests: List names ---

func TestHandlerListNamesSuccess(t *testing.T) {
	q := &mockQuerier{
		getParentByEmailFn: func(ctx context.Context, email string) (queries.Parent, error) {
			return queries.Parent{ID: parseUUID(testParentUUID), Email: email}, nil
		},
		listStudentNamesByParentIDFn: func(ctx context.Context, parentID pgtype.UUID) ([]queries.ListStudentNamesByParentIDRow, error) {
			return []queries.ListStudentNamesByParentIDRow{
				{ID: parseUUID(testStudentUUID), Name: "Erik"},
				{ID: parseUUID("880e8400-e29b-41d4-a716-446655440000"), Name: "Anna"},
			}, nil
		},
	}
	h := newTestHandler(q)

	r := httptest.NewRequest(http.MethodGet, "/api/child/names?parent_email=parent@test.com", nil)
	w := httptest.NewRecorder()

	h.ListNames(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var resp []map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if len(resp) != 2 {
		t.Errorf("got %d names, want 2", len(resp))
	}
}
