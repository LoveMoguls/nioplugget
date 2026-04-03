package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/jwtauth/v5"

	"github.com/trollstaven/nioplugget/backend/internal/auth"
)

func makeTestServer(middleware func(http.Handler) http.Handler) *httptest.Server {
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	return httptest.NewServer(handler)
}

func makeTokenRequest(ts *httptest.Server, tokenStr string) *http.Response {
	req, _ := http.NewRequest("GET", ts.URL, nil)
	if tokenStr != "" {
		req.Header.Set("Authorization", "Bearer "+tokenStr)
	}
	client := ts.Client()
	resp, _ := client.Do(req)
	return resp
}

func makeTokenRequestFromCookie(ts *httptest.Server, tokenStr string) *http.Response {
	req, _ := http.NewRequest("GET", ts.URL, nil)
	if tokenStr != "" {
		req.AddCookie(&http.Cookie{Name: "token", Value: tokenStr})
	}
	client := ts.Client()
	resp, _ := client.Do(req)
	return resp
}

func TestMiddleware_ParentOnly_NoToken_Returns401(t *testing.T) {
	auth.NewTokenAuth("test-secret")

	// Wrap with jwtauth verifier + authenticator + ParentOnly
	handler := jwtauth.Verifier(auth.TokenAuth)(
		jwtauth.Authenticator(auth.TokenAuth)(
			auth.ParentOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		),
	)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp := makeTokenRequest(ts, "")
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
}

func TestMiddleware_ParentOnly_ChildToken_Returns403(t *testing.T) {
	auth.NewTokenAuth("test-secret")
	childToken, _, _ := auth.GenerateToken("child-1", "child")

	handler := jwtauth.Verifier(auth.TokenAuth)(
		jwtauth.Authenticator(auth.TokenAuth)(
			auth.ParentOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		),
	)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp := makeTokenRequest(ts, childToken)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

func TestMiddleware_ParentOnly_ParentToken_Allows(t *testing.T) {
	auth.NewTokenAuth("test-secret")
	parentToken, _, _ := auth.GenerateToken("parent-1", "parent")

	handler := jwtauth.Verifier(auth.TokenAuth)(
		jwtauth.Authenticator(auth.TokenAuth)(
			auth.ParentOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		),
	)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp := makeTokenRequest(ts, parentToken)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestMiddleware_ChildOnly_ParentToken_Returns403(t *testing.T) {
	auth.NewTokenAuth("test-secret")
	parentToken, _, _ := auth.GenerateToken("parent-1", "parent")

	handler := jwtauth.Verifier(auth.TokenAuth)(
		jwtauth.Authenticator(auth.TokenAuth)(
			auth.ChildOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		),
	)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp := makeTokenRequest(ts, parentToken)
	if resp.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403, got %d", resp.StatusCode)
	}
}

func TestMiddleware_ChildOnly_ChildToken_Allows(t *testing.T) {
	auth.NewTokenAuth("test-secret")
	childToken, _, _ := auth.GenerateToken("child-1", "child")

	handler := jwtauth.Verifier(auth.TokenAuth)(
		jwtauth.Authenticator(auth.TokenAuth)(
			auth.ChildOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})),
		),
	)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	resp := makeTokenRequest(ts, childToken)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}
