package chat

import (
	"context"
	"net/http"
	"net/http/httptest"
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
