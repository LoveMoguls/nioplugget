package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/cors"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing.
// The allowed origin is read from FRONTEND_URL (default: http://localhost:5173).
func CORS() func(http.Handler) http.Handler {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	origins := []string{frontendURL}
	if extra := os.Getenv("EXTRA_ORIGINS"); extra != "" {
		for _, o := range strings.Split(extra, ",") {
			origins = append(origins, strings.TrimSpace(o))
		}
	}

	return cors.Handler(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
