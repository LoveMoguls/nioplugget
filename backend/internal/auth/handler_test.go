package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/trollstaven/nioplugget/backend/internal/auth"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// mockQueries implements auth.ParentQuerier for testing.
type mockQueries struct {
	createParentFn    func(ctx context.Context, arg queries.CreateParentParams) (queries.Parent, error)
	getParentByEmailFn func(ctx context.Context, email string) (queries.Parent, error)
}

func (m *mockQueries) CreateParent(ctx context.Context, arg queries.CreateParentParams) (queries.Parent, error) {
	if m.createParentFn != nil {
		return m.createParentFn(ctx, arg)
	}
	return queries.Parent{}, nil
}

func (m *mockQueries) GetParentByEmail(ctx context.Context, email string) (queries.Parent, error) {
	if m.getParentByEmailFn != nil {
		return m.getParentByEmailFn(ctx, email)
	}
	return queries.Parent{}, nil
}

// fakeParent returns a queries.Parent with hashed password for test email/password.
func fakeParent(email, password string) queries.Parent {
	hash, _ := auth.HashPassword(password)
	return queries.Parent{
		ID:    pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
		Email: email,
		PasswordHash: hash,
		GdprConsentAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

func setupHandler(q auth.ParentQuerier) *auth.AuthHandler {
	auth.NewTokenAuth("test-secret")
	return auth.NewAuthHandler(q, auth.TokenAuth)
}

func doRequest(handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var buf *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		buf = bytes.NewBuffer(b)
	} else {
		buf = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// --- Register tests ---

func TestHandler_Register_ValidInput_Returns201(t *testing.T) {
	mock := &mockQueries{
		createParentFn: func(ctx context.Context, arg queries.CreateParentParams) (queries.Parent, error) {
			return queries.Parent{
				ID:    pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				Email: arg.Email,
			}, nil
		},
	}
	h := setupHandler(mock)

	rr := doRequest(http.HandlerFunc(h.Register), "POST", "/api/auth/register", map[string]any{
		"email":       "test@example.com",
		"password":    "securepassword",
		"gdprConsent": true,
	})

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	// Should set cookie
	cookieHeader := rr.Header().Get("Set-Cookie")
	if !strings.Contains(cookieHeader, "token=") {
		t.Error("expected token cookie to be set")
	}
	if !strings.Contains(cookieHeader, "HttpOnly") {
		t.Error("expected HttpOnly cookie attribute")
	}
}

func TestHandler_Register_DuplicateEmail_Returns409(t *testing.T) {
	mock := &mockQueries{
		createParentFn: func(ctx context.Context, arg queries.CreateParentParams) (queries.Parent, error) {
			// Simulate unique constraint violation
			return queries.Parent{}, &pgconn.PgError{Code: "23505"}
		},
	}
	h := setupHandler(mock)

	rr := doRequest(http.HandlerFunc(h.Register), "POST", "/api/auth/register", map[string]any{
		"email":       "existing@example.com",
		"password":    "securepassword",
		"gdprConsent": true,
	})

	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestHandler_Register_ShortPassword_Returns400(t *testing.T) {
	h := setupHandler(&mockQueries{})

	rr := doRequest(http.HandlerFunc(h.Register), "POST", "/api/auth/register", map[string]any{
		"email":       "test@example.com",
		"password":    "short",
		"gdprConsent": true,
	})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandler_Register_NoGdprConsent_Returns400(t *testing.T) {
	h := setupHandler(&mockQueries{})

	rr := doRequest(http.HandlerFunc(h.Register), "POST", "/api/auth/register", map[string]any{
		"email":       "test@example.com",
		"password":    "securepassword",
		"gdprConsent": false,
	})

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

// --- Login tests ---

func TestHandler_Login_ValidCredentials_Returns200(t *testing.T) {
	parent := fakeParent("test@example.com", "securepassword")
	mock := &mockQueries{
		getParentByEmailFn: func(ctx context.Context, email string) (queries.Parent, error) {
			return parent, nil
		},
	}
	h := setupHandler(mock)

	rr := doRequest(http.HandlerFunc(h.Login), "POST", "/api/auth/login", map[string]any{
		"email":    "test@example.com",
		"password": "securepassword",
	})

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	cookieHeader := rr.Header().Get("Set-Cookie")
	if !strings.Contains(cookieHeader, "token=") {
		t.Error("expected token cookie to be set")
	}
}

func TestHandler_Login_WrongPassword_Returns401(t *testing.T) {
	parent := fakeParent("test@example.com", "correctpassword")
	mock := &mockQueries{
		getParentByEmailFn: func(ctx context.Context, email string) (queries.Parent, error) {
			return parent, nil
		},
	}
	h := setupHandler(mock)

	rr := doRequest(http.HandlerFunc(h.Login), "POST", "/api/auth/login", map[string]any{
		"email":    "test@example.com",
		"password": "wrongpassword",
	})

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestHandler_Login_UnknownEmail_Returns401(t *testing.T) {
	mock := &mockQueries{
		getParentByEmailFn: func(ctx context.Context, email string) (queries.Parent, error) {
			return queries.Parent{}, errors.New("no rows")
		},
	}
	h := setupHandler(mock)

	rr := doRequest(http.HandlerFunc(h.Login), "POST", "/api/auth/login", map[string]any{
		"email":    "unknown@example.com",
		"password": "anypassword",
	})

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

// --- Logout tests ---

func TestHandler_Logout_Returns200(t *testing.T) {
	h := setupHandler(&mockQueries{})

	rr := doRequest(http.HandlerFunc(h.Logout), "POST", "/api/auth/logout", nil)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	// Cookie should be cleared (MaxAge=-1)
	cookieHeader := rr.Header().Get("Set-Cookie")
	if !strings.Contains(cookieHeader, "token=") {
		t.Error("expected token cookie to be cleared")
	}
	if !strings.Contains(cookieHeader, "Max-Age=0") {
		t.Errorf("expected Max-Age=0 in cookie, got: %s", cookieHeader)
	}
}
