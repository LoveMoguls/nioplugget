package chat

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// DefaultWindowSize is the number of messages to keep in the sliding window.
const DefaultWindowSize = 20

// BuildMessageHistory converts database messages to Claude API params,
// applying a sliding window to control token costs.
func BuildMessageHistory(messages []queries.Message, windowSize int) []anthropic.MessageParam {
	if windowSize <= 0 {
		windowSize = DefaultWindowSize
	}

	// Apply sliding window
	if len(messages) > windowSize {
		messages = messages[len(messages)-windowSize:]
	}

	// Ensure first message is from "user" (Claude API requirement)
	for len(messages) > 0 && messages[0].Role == "assistant" {
		messages = messages[1:]
	}

	// Convert to SDK params
	params := make([]anthropic.MessageParam, 0, len(messages))
	for _, m := range messages {
		if m.Role == "user" {
			params = append(params, anthropic.NewUserMessage(anthropic.NewTextBlock(m.Content)))
		} else {
			params = append(params, anthropic.NewAssistantMessage(anthropic.NewTextBlock(m.Content)))
		}
	}
	return params
}
