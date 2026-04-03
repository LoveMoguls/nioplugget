package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// ScoreResult holds the AI-generated session evaluation.
type ScoreResult struct {
	Score    int    `json:"score"`    // 1-5
	Summary  string `json:"summary"`  // 2-3 sentence summary
	Feedback string `json:"feedback"` // Encouraging feedback
}

// ScoreSession evaluates student performance using a separate Claude call.
func ScoreSession(ctx context.Context, apiKey string, exercise queries.GetExerciseByIDRow, messages []queries.Message) (*ScoreResult, error) {
	scoringPrompt := fmt.Sprintf(`Du är en bedömare av elevers prestation i en pedagogisk dialog.

ÖVNING: %s

UPPGIFT: Bedöm elevens prestation på en skala 1-5 baserat på samtalet nedan.

BEDÖMNINGSKRITERIER:
1 = Eleven visade inte förståelse för ämnet
2 = Eleven visade viss förståelse men hade stora luckor
3 = Eleven visade grundläggande förståelse
4 = Eleven visade god förståelse och kunde resonera
5 = Eleven visade djup förståelse och kunde koppla ihop begrepp

Svara med EXAKT detta JSON-format, inget annat:
{"score": <1-5>, "summary": "<2-3 meningar om vad eleven lärde sig>", "feedback": "<1 mening uppmuntrande men ärlig feedback>"}`, exercise.Title)

	// Build conversation context for scoring
	var convoBuilder strings.Builder
	for _, m := range messages {
		role := "Elev"
		if m.Role == "assistant" {
			role = "AI-lärare"
		}
		fmt.Fprintf(&convoBuilder, "%s: %s\n\n", role, m.Content)
	}

	client := anthropic.NewClient(option.WithAPIKey(apiKey))
	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 256,
		System: []anthropic.TextBlockParam{
			{Text: scoringPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(
				"Här är samtalet att bedöma:\n\n" + convoBuilder.String(),
			)),
		},
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to score session")
		return defaultScore(), nil
	}

	// Extract text from response
	var responseText string
	for _, block := range resp.Content {
		switch b := block.AsAny().(type) {
		case anthropic.TextBlock:
			responseText = b.Text
		}
	}

	// Parse JSON response
	var result ScoreResult
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		log.Error().Err(err).Str("response", responseText).Msg("failed to parse score JSON")
		return defaultScore(), nil
	}

	// Validate score range
	if result.Score < 1 || result.Score > 5 {
		result.Score = 3
	}

	return &result, nil
}

func defaultScore() *ScoreResult {
	return &ScoreResult{
		Score:    3,
		Summary:  "Passet avslutades utan bedömning.",
		Feedback: "Bra jobbat att du övade!",
	}
}
