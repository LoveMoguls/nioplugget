package chat

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// CompleteChatResponse gets a full (non-streaming) Claude reply. Used by
// channels that cannot stream, e.g. the Telegram bot. Extra request options
// (like option.WithBaseURL in tests) may be passed.
func CompleteChatResponse(ctx context.Context, apiKey, systemPrompt string, messages []anthropic.MessageParam, opts ...option.RequestOption) (string, error) {
	clientOpts := append([]option.RequestOption{option.WithAPIKey(apiKey)}, opts...)
	client := anthropic.NewClient(clientOpts...)
	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 1024,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Messages:  messages,
	})
	if err != nil {
		return "", fmt.Errorf("claude completion: %w", err)
	}
	var text string
	for _, block := range resp.Content {
		if b, ok := block.AsAny().(anthropic.TextBlock); ok {
			text += b.Text
		}
	}
	return text, nil
}
