package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/trollstaven/nioplugget/backend/internal/apikey"
	"github.com/trollstaven/nioplugget/backend/internal/auth"
	"github.com/trollstaven/nioplugget/backend/internal/child"
	"github.com/trollstaven/nioplugget/backend/internal/database"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
	appMiddleware "github.com/trollstaven/nioplugget/backend/internal/middleware"
)

func main() {
	// Configure zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Load required env vars
	databaseURL := mustEnv("DATABASE_URL")
	jwtSecret := mustEnv("JWT_SECRET")
	encryptionKey := mustEnv("ENCRYPTION_KEY")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database pool
	ctx := context.Background()
	pool, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()
	log.Info().Msg("database pool established")

	// Initialize JWT auth
	tokenAuth := auth.NewTokenAuth(jwtSecret)

	// Initialize database queries
	q := queries.New(pool)

	// Initialize auth handler
	authHandler := auth.NewAuthHandler(q, tokenAuth)

	// Initialize API key handler
	encSvc, err := apikey.NewEncryptionService(encryptionKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize encryption service")
	}
	apiKeyStore := apikey.NewQueriesStore(q)
	apiKeyHandler := apikey.NewAPIKeyHandler(apiKeyStore, encSvc, "")

	// Initialize child handler
	childStore := child.NewQueriesStore(q, pool)
	childRateLimiter := child.NewPINRateLimiter(5, 15*time.Minute)
	childHandler := child.NewChildHandler(childStore, childRateLimiter)

	// Build router
	r := chi.NewRouter()

	// Middleware stack
	r.Use(appMiddleware.RedactingLogger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(appMiddleware.CORS())

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})

	// Auth routes (public)
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)

		// Protected: logout requires valid parent JWT
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))
			r.Use(auth.ParentOnly)
			r.Post("/logout", authHandler.Logout)
		})
	})

	// API key routes (protected: parent JWT required)
	r.Route("/api/apikey", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Use(auth.ParentOnly)
		r.Post("/", apiKeyHandler.Store)
		r.Get("/", apiKeyHandler.Get)
		r.Put("/", apiKeyHandler.Update)
		r.Delete("/", apiKeyHandler.Delete)
	})
	// Child management routes (protected: parent JWT required)
	r.Route("/api/children", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Use(auth.ParentOnly)
		r.Post("/", childHandler.Create)
		r.Get("/", childHandler.List)
		r.Post("/{id}/invite", childHandler.GenerateInvite)
	})

	// Public child routes
	r.Route("/api/child", func(r chi.Router) {
		r.Post("/login", childHandler.PINLogin)
		r.Get("/names", childHandler.ListNames)
	})

	// Invite activation route (public)
	r.Route("/api/invite", func(r chi.Router) {
		r.Post("/{token}/activate", childHandler.Activate)
	})

	addr := ":" + port
	log.Info().Str("addr", addr).Msg("starting server")

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("server error")
	}
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatal().Str("key", key).Msg("required environment variable is not set")
	}
	return val
}
