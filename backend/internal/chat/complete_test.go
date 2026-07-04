package chat

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

func TestCompleteChatResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"msg_test","type":"message","role":"assistant","model":"claude-sonnet-4-6","content":[{"type":"text","text":"Vad tror du själv?"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`))
	}))
	defer srv.Close()

	got, err := CompleteChatResponse(context.Background(), "test-key", "system",
		[]anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock("hej"))},
		option.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Vad tror du själv?" {
		t.Errorf("got %q, want %q", got, "Vad tror du själv?")
	}
}

func TestCompleteChatResponse_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"type":"error","error":{"type":"invalid_request_error","message":"bad"}}`))
	}))
	defer srv.Close()

	got, err := CompleteChatResponse(context.Background(), "test-key", "system",
		[]anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock("hej"))},
		option.WithBaseURL(srv.URL))
	if err == nil {
		t.Fatalf("expected error, got nil (result %q)", got)
	}
	if !strings.Contains(err.Error(), "claude completion") {
		t.Errorf("error %q does not contain %q", err.Error(), "claude completion")
	}
}

func TestCompleteChatResponse_MultiBlockConcatenation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"msg_test","type":"message","role":"assistant","model":"claude-sonnet-4-6","content":[{"type":"text","text":"Hej! "},{"type":"text","text":"Vad tror du själv?"}],"stop_reason":"end_turn","usage":{"input_tokens":1,"output_tokens":1}}`))
	}))
	defer srv.Close()

	got, err := CompleteChatResponse(context.Background(), "test-key", "system",
		[]anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock("hej"))},
		option.WithBaseURL(srv.URL))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "Hej! Vad tror du själv?"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
