package middleware

import (
	"net/http"
	"os"

	"github.com/go-chi/cors"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
// The allowed origin is read from FRONTEND_URL (default: http://localhost:5173).
func CORS() func(http.Handler) http.Handler {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
