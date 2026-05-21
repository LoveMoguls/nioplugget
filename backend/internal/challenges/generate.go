// backend/internal/challenges/generate.go
package challenges

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// GeneratedExercise is one Socratic exercise from Claude.
type GeneratedExercise struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	SystemPrompt string `json:"system_prompt"`
}

// GeneratedChallenge is the full course Claude creates from homework images.
type GeneratedChallenge struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Emoji       string              `json:"emoji"`
	Exercises   []GeneratedExercise `json:"exercises"`
}

const generationSystemPrompt = `Du är en pedagog som skapar engagerande studieövningar för en 13-årig elev.
Analysera bilderna och skapa ett JSON-objekt med exakt följande struktur:
{
  "title": "En catchy, motiverande titel på svenska (max 40 tecken, gärna med utropstecken)",
  "description": "En kort mening som beskriver vad utmaningen handlar om",
  "emoji": "En passande emoji som representerar ämnet",
  "exercises": [
    {
      "title": "Kort övningsrubrik (max 30 tecken)",
      "description": "En mening som beskriver vad övningen handlar om",
      "system_prompt": "Fullständig systemPrompt på svenska. Du är en Sokratisk lärare — ställ alltid ledande frågor, ge ALDRIG direkta svar. Anpassa språket för en 13-åring. Var uppmuntrande och energisk."
    }
  ]
}

Skapa 4-6 övningar. Svara ENBART med JSON, ingen annan text.`

// GenerateChallenge calls Claude with the provided images and returns a structured challenge.
// mediaTypes should be MIME types like "image/jpeg", "image/png", "image/webp".
func GenerateChallenge(ctx context.Context, apiKey string, imageDataList [][]byte, mediaTypes []string) (*GeneratedChallenge, error) {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	var contentBlocks []anthropic.ContentBlockParamUnion
	for i, data := range imageDataList {
		encoded := base64.StdEncoding.EncodeToString(data)
		contentBlocks = append(contentBlocks, anthropic.NewImageBlockBase64(mediaTypes[i], encoded))
	}
	contentBlocks = append(contentBlocks, anthropic.NewTextBlock("Analysera bilderna och skapa övningarna som JSON."))

	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 2048,
		System: []anthropic.TextBlockParam{
			{Text: generationSystemPrompt},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(contentBlocks...),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("claude API error: %w", err)
	}

	var responseText string
	for _, block := range resp.Content {
		if tb, ok := block.AsAny().(anthropic.TextBlock); ok {
			responseText = tb.Text
		}
	}

	var result GeneratedChallenge
	if err := json.Unmarshal([]byte(responseText), &result); err != nil {
		return nil, fmt.Errorf("failed to parse challenge JSON from Claude: %w", err)
	}

	if len(result.Exercises) == 0 {
		return nil, fmt.Errorf("Claude generated no exercises")
	}
	if result.Emoji == "" {
		result.Emoji = "📚"
	}

	return &result, nil
}
