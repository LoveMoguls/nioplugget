package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAPISendMessageAndError(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer srv.Close()

	api := &API{baseURL: srv.URL + "/botTEST", client: srv.Client()}
	if err := api.SendMessage(context.Background(), 42, "hej", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotPath != "/botTEST/sendMessage" {
		t.Errorf("path = %q", gotPath)
	}
	if gotBody["chat_id"].(float64) != 42 || gotBody["text"] != "hej" {
		t.Errorf("body = %v", gotBody)
	}

	errSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"ok":false,"description":"Bad Request"}`))
	}))
	defer errSrv.Close()
	badAPI := &API{baseURL: errSrv.URL, client: errSrv.Client()}
	if err := badAPI.SendMessage(context.Background(), 42, "hej", nil); err == nil {
		t.Error("expected error on ok=false")
	}
}
