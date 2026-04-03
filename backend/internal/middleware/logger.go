package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// sensitiveHeaders is a set of header names (lowercase) whose values must never be logged.
var sensitiveHeaders = map[string]struct{}{
	"authorization": {},
	"x-api-key":     {},
}

// sensitiveQueryParams is a set of query param name substrings that trigger redaction.
var sensitiveQuerySubstrings = []string{"key", "token", "secret", "password", "pin"}

// isSensitiveParam returns true if the query parameter name looks sensitive.
func isSensitiveParam(name string) bool {
	lower := strings.ToLower(name)
	for _, sub := range sensitiveQuerySubstrings {
		if strings.Contains(lower, sub) {
			return true
		}
	}
	return false
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.status == 0 {
		rw.status = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// RedactingLogger is a Chi-compatible middleware that logs requests while
// never outputting sensitive header values or query parameters.
func RedactingLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		status := wrapped.status

		// Build safe query string representation
		query := r.URL.Query()
		safeParams := make([]string, 0, len(query))
		for k := range query {
			if isSensitiveParam(k) {
				safeParams = append(safeParams, k+"=[REDACTED]")
			} else {
				safeParams = append(safeParams, k+"="+query.Get(k))
			}
		}

		event := log.Info()
		if status >= 500 {
			event = log.Error()
		} else if status >= 400 {
			event = log.Warn()
		}

		event.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", status).
			Dur("duration", duration).
			Str("remote_addr", r.RemoteAddr).
			Strs("query", safeParams).
			Msg("request")
	})
}
