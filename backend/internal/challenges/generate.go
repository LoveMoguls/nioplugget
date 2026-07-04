// backend/internal/challenges/generate.go
package challenges

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// GeneratedExercise is one Socratic exercise from Claude.
type GeneratedExercise struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	SystemPrompt string `json:"system_prompt"`
}

// GeneratedChallenge is the full course Claude creates from homework material.
type GeneratedChallenge struct {
	Title       string              `json:"title"`
	Description string              `json:"description"`
	Emoji       string              `json:"emoji"`
	Exercises   []GeneratedExercise `json:"exercises"`
}

// UploadPart is one uploaded file to base the challenge on.
type UploadPart struct {
	Data      []byte
	MediaType string // "image/jpeg", "image/png", "image/gif", "image/webp" or "application/pdf"
}

const generationSystemPrompt = `Du är en pedagog som skapar engagerande studieövningar för en 13-årig elev.
Analysera materialet (bilder, dokument och/eller text från en läxa) och skapa ett JSON-objekt med exakt följande struktur:
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

// GenerateChallenge calls Claude with the provided homework material (images,
// PDFs and/or pasted text) and returns a structured challenge.
func GenerateChallenge(ctx context.Context, apiKey string, parts []UploadPart, pastedText string) (*GeneratedChallenge, error) {
	client := anthropic.NewClient(option.WithAPIKey(apiKey))

	var contentBlocks []anthropic.ContentBlockParamUnion
	for _, part := range parts {
		encoded := base64.StdEncoding.EncodeToString(part.Data)
		if part.MediaType == "application/pdf" {
			contentBlocks = append(contentBlocks, anthropic.NewDocumentBlock(anthropic.Base64PDFSourceParam{Data: encoded}))
		} else {
			contentBlocks = append(contentBlocks, anthropic.NewImageBlockBase64(part.MediaType, encoded))
		}
	}
	if pastedText != "" {
		contentBlocks = append(contentBlocks, anthropic.NewTextBlock("LÄXTEXT:\n"+pastedText))
	}
	if len(contentBlocks) == 0 {
		return nil, fmt.Errorf("no material provided")
	}
	contentBlocks = append(contentBlocks, anthropic.NewTextBlock("Analysera materialet och skapa övningarna som JSON."))

	resp, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeSonnet4_6,
		MaxTokens: 4096,
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

	// The model is told to answer with bare JSON, but tolerate ```json fences
	responseText = strings.TrimSpace(responseText)
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

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
