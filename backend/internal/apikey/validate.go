package apikey

import (
	"fmt"
	"net/http"
	"time"
)

const anthropicModelsURL = "https://api.anthropic.com/v1/models"

// ValidateAPIKey validates an API key against the Claude API.
// Returns nil if the key is valid.
func ValidateAPIKey(apiKey string) error {
	return ValidateAPIKeyWithURL(apiKey, anthropicModelsURL)
}

// ValidateAPIKeyWithURL validates an API key using the provided URL.
// This allows tests to override the Anthropic API endpoint.
func ValidateAPIKeyWithURL(apiKey, url string) error {
	if apiKey == "" {
		return fmt.Errorf("API-nyckel saknas")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("Kunde inte ansluta till Anthropic")
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Kunde inte ansluta till Anthropic")
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("Ogiltig API-nyckel")
	default:
		return fmt.Errorf("Anthropic svarade med statuskod %d", resp.StatusCode)
	}
}

// MaskAPIKey returns a masked version of the API key.
// If the key is shorter than 8 characters, returns "****".
// Otherwise, returns the first 6 characters followed by "...****".
func MaskAPIKey(key string) string {
	if len(key) < 8 {
		return "****"
	}
	return key[:6] + "...****"
}
