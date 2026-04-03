package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/rs/zerolog/log"
)

// StreamChatResponse streams a Claude API response as SSE events to the client.
// Returns the full AI response text for persistence.
func StreamChatResponse(w http.ResponseWriter, ctx context.Context, apiKey string, systemPrompt string, messages []anthropic.MessageParam) (string, error) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // nginx proxy support

	flusher, ok := w.(http.Flusher)
	if !ok {
		return "", fmt.Errorf("streaming not supported")
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	stream := client.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 1024,
		System: []anthropic.TextBlockParam{
			{Text: systemPrompt},
		},
		Messages: messages,
	})

	var fullResponse strings.Builder
	message := anthropic.Message{}
	for stream.Next() {
		event := stream.Current()
		if err := message.Accumulate(event); err != nil {
			log.Error().Err(err).Msg("failed to accumulate stream event")
			continue
		}

		switch ev := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			switch delta := ev.Delta.AsAny().(type) {
			case anthropic.TextDelta:
				fullResponse.WriteString(delta.Text)
				data, _ := json.Marshal(map[string]string{"text": delta.Text})
				fmt.Fprintf(w, "data: %s\n\n", data)
				flusher.Flush()
			}
		}
	}

	if stream.Err() != nil {
		log.Error().Err(stream.Err()).Msg("claude stream error")
		errData, _ := json.Marshal(map[string]string{"error": "AI-tjänsten svarade inte korrekt. Försök igen."})
		fmt.Fprintf(w, "data: %s\n\n", errData)
		flusher.Flush()
		return fullResponse.String(), stream.Err()
	}

	// Send done event
	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()

	return fullResponse.String(), nil
}
