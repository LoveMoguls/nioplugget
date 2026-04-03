package apikey_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/trollstaven/nioplugget/backend/internal/apikey"
)

// --- Mock DB store ---

type mockStore struct {
	upsertFn func(parentID pgtype.UUID, encKey []byte) (*apikey.APIKeyRecord, error)
	getFn    func(parentID pgtype.UUID) (*apikey.APIKeyRecord, error)
	deleteFn func(parentID pgtype.UUID) error
}

func (m *mockStore) UpsertAPIKey(parentID pgtype.UUID, encKey []byte) (*apikey.APIKeyRecord, error) {
	return m.upsertFn(parentID, encKey)
}

func (m *mockStore) GetAPIKeyByParentID(parentID pgtype.UUID) (*apikey.APIKeyRecord, error) {
	return m.getFn(parentID)
}

func (m *mockStore) DeleteAPIKeyByParentID(parentID pgtype.UUID) error {
	return m.deleteFn(parentID)
}

// --- Handler helpers ---

func newTestHandler(t *testing.T, store apikey.APIKeyStore, validateURL string) *apikey.APIKeyHandler {
	t.Helper()
	encKey := make([]byte, 32)
	hexKey := hex.EncodeToString(encKey)
	encSvc, err := apikey.NewEncryptionService(hexKey)
	if err != nil {
		t.Fatalf("failed to create encryption service: %v", err)
	}
	return apikey.NewAPIKeyHandler(store, encSvc, validateURL)
}

// injectParentContext sets a JWT with the given parentID as "sub" into the request context.
func injectParentContext(r *http.Request, parentID string) *http.Request {
	ta := jwtauth.New("HS256", []byte("test-secret"), nil)
	_, tokenStr, _ := ta.Encode(map[string]any{"sub": parentID, "role": "parent"})
	token, _ := ta.Decode(tokenStr)
	ctx := jwtauth.NewContext(r.Context(), token, nil)
	return r.WithContext(ctx)
}

// newAnthropicMock returns an httptest server that responds with the given status code.
func newAnthropicMock(status int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
}

// --- Store mock helpers ---

func defaultUpsert(parentID pgtype.UUID, encKey []byte) (*apikey.APIKeyRecord, error) {
	return &apikey.APIKeyRecord{
		EncryptedKey: encKey,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// --- Tests ---

const testParentID = "550e8400-e29b-41d4-a716-446655440000"

func TestHandler_Store_ValidKey(t *testing.T) {
	mock := newAnthropicMock(http.StatusOK)
	defer mock.Close()

	store := &mockStore{
		upsertFn: defaultUpsert,
	}
	h := newTestHandler(t, store, mock.URL+"/v1/models")

	body := `{"apiKey":"sk-ant-api03-testvalidkey123456789012"}`
	req := httptest.NewRequest(http.MethodPost, "/api/apikey", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectParentContext(req, testParentID)

	rr := httptest.NewRecorder()
	h.Store(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["maskedKey"] == "" {
		t.Fatal("expected maskedKey in response")
	}
	if resp["message"] == "" {
		t.Fatal("expected message in response")
	}
}

func TestHandler_Store_InvalidKey(t *testing.T) {
	mock := newAnthropicMock(http.StatusUnauthorized)
	defer mock.Close()

	store := &mockStore{}
	h := newTestHandler(t, store, mock.URL+"/v1/models")

	body := `{"apiKey":"sk-ant-invalid"}`
	req := httptest.NewRequest(http.MethodPost, "/api/apikey", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectParentContext(req, testParentID)

	rr := httptest.NewRecorder()
	h.Store(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid key, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["error"] == "" {
		t.Fatal("expected error message in response")
	}
}

func TestHandler_Get_WithKey(t *testing.T) {
	encKey := make([]byte, 32)
	hexKey := hex.EncodeToString(encKey)
	encSvc, _ := apikey.NewEncryptionService(hexKey)
	rawKey := "sk-ant-api03-abc"
	encrypted, _ := encSvc.Encrypt([]byte(rawKey))

	store := &mockStore{
		getFn: func(parentID pgtype.UUID) (*apikey.APIKeyRecord, error) {
			return &apikey.APIKeyRecord{
				EncryptedKey: encrypted,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}, nil
		},
	}
	h := newTestHandler(t, store, "")

	req := httptest.NewRequest(http.MethodGet, "/api/apikey", nil)
	req = injectParentContext(req, testParentID)
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["maskedKey"] == "" {
		t.Fatal("expected maskedKey in response")
	}
	// Verify raw key is NOT in response
	respStr := rr.Body.String()
	if containsStr(respStr, rawKey) {
		t.Fatalf("raw API key must not appear in response: %s", respStr)
	}
}

func TestHandler_Get_NotFound(t *testing.T) {
	store := &mockStore{
		getFn: func(parentID pgtype.UUID) (*apikey.APIKeyRecord, error) {
			return nil, apikey.ErrNotFound
		},
	}
	h := newTestHandler(t, store, "")

	req := httptest.NewRequest(http.MethodGet, "/api/apikey", nil)
	req = injectParentContext(req, testParentID)
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when no key stored, got %d", rr.Code)
	}
}

func TestHandler_Update_ValidKey(t *testing.T) {
	mock := newAnthropicMock(http.StatusOK)
	defer mock.Close()

	store := &mockStore{
		upsertFn: defaultUpsert,
	}
	h := newTestHandler(t, store, mock.URL+"/v1/models")

	body := `{"apiKey":"sk-ant-api03-updatekey12345678901234"}`
	req := httptest.NewRequest(http.MethodPut, "/api/apikey", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req = injectParentContext(req, testParentID)

	rr := httptest.NewRecorder()
	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for update, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["maskedKey"] == "" {
		t.Fatal("expected maskedKey in update response")
	}
}

func TestHandler_Delete(t *testing.T) {
	store := &mockStore{
		deleteFn: func(parentID pgtype.UUID) error {
			return nil
		},
	}
	h := newTestHandler(t, store, "")

	req := httptest.NewRequest(http.MethodDelete, "/api/apikey", nil)
	req = injectParentContext(req, testParentID)
	rr := httptest.NewRecorder()
	h.Delete(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for delete, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp["message"] == "" {
		t.Fatal("expected message in delete response")
	}
}

// containsStr checks if s contains substr
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
