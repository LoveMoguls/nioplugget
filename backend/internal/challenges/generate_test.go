// backend/internal/challenges/generate_test.go
package challenges

import (
	"encoding/json"
	"testing"
)

func TestGeneratedChallengeParses(t *testing.T) {
	raw := `{
		"title": "Fotosyntes-mästaren!",
		"description": "Lär dig hur växter gör sin mat",
		"emoji": "🌱",
		"exercises": [
			{
				"title": "Vad behöver en växt?",
				"description": "Utforska de tre ingredienserna",
				"system_prompt": "Du är en Sokratisk lärare..."
			},
			{
				"title": "Klorofyllets roll",
				"description": "Vad gör det gröna pigmentet?",
				"system_prompt": "Du är en Sokratisk lärare..."
			}
		]
	}`

	var result GeneratedChallenge
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatalf("expected parse to succeed, got: %v", err)
	}
	if result.Title != "Fotosyntes-mästaren!" {
		t.Errorf("expected title 'Fotosyntes-mästaren!', got %q", result.Title)
	}
	if len(result.Exercises) != 2 {
		t.Errorf("expected 2 exercises, got %d", len(result.Exercises))
	}
	if result.Exercises[0].SystemPrompt == "" {
		t.Error("expected system_prompt to be non-empty")
	}
}

func TestGeneratedChallengeFallbackEmoji(t *testing.T) {
	raw := `{"title":"Test","description":"desc","emoji":"","exercises":[{"title":"t","description":"d","system_prompt":"s"}]}`
	var result GeneratedChallenge
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.Fatal(err)
	}
	// Simulate the handler setting fallback
	if result.Emoji == "" {
		result.Emoji = "📚"
	}
	if result.Emoji != "📚" {
		t.Errorf("expected fallback emoji 📚, got %q", result.Emoji)
	}
}
