package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/trollstaven/nioplugget/backend/internal/apikey"
	"github.com/trollstaven/nioplugget/backend/internal/auth"
	"github.com/trollstaven/nioplugget/backend/internal/challenges"
	"github.com/trollstaven/nioplugget/backend/internal/chat"
	"github.com/trollstaven/nioplugget/backend/internal/child"
	"github.com/trollstaven/nioplugget/backend/internal/content"
	"github.com/trollstaven/nioplugget/backend/internal/database"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
	"github.com/trollstaven/nioplugget/backend/internal/device"
	appMiddleware "github.com/trollstaven/nioplugget/backend/internal/middleware"
	"github.com/trollstaven/nioplugget/backend/internal/progress"
	"github.com/trollstaven/nioplugget/backend/internal/srs"
	"github.com/trollstaven/nioplugget/backend/internal/telegram"
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

	// Initialize device store (family settings, parents, students) — also used by apikey's
	// family-code verifier below and reused for device/profile endpoints (see Task 6).
	deviceStore := device.NewQueriesStore(q)

	// Initialize API key handler
	encSvc, err := apikey.NewEncryptionService(encryptionKey)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize encryption service")
	}
	apiKeyStore := apikey.NewQueriesStore(q)
	apiKeyHandler := apikey.NewAPIKeyHandler(apiKeyStore, encSvc, "")

	// Wire the family-code verifier: no family code set (ErrNoRows) means no requirement;
	// any other DB error fails closed (requirement on, verification failed) rather than
	// silently allowing the request through.
	familyCodeVerifier := func(ctx context.Context, code string) (bool, bool) {
		settings, err := deviceStore.GetFamilySettings(ctx)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return false, false // no code set — no requirement
			}
			return true, false // unknown DB error — fail closed
		}
		match, err := auth.ComparePassword(settings.CodeHash, code)
		return true, err == nil && match
	}
	apiKeyHandler.SetFamilyCodeVerifier(familyCodeVerifier)

	// Initialize child handler
	childStore := child.NewQueriesStore(q, pool)
	childRateLimiter := child.NewPINRateLimiter(5, 15*time.Minute)
	childHandler := child.NewChildHandler(childStore, childRateLimiter)

	// Initialize content handler
	contentStore := content.NewQueriesStore(q)
	contentHandler := content.NewContentHandler(contentStore)

	// Initialize chat handler
	chatStore := chat.NewQueriesStore(q)
	chatHandler := chat.NewChatHandler(chatStore, encSvc)

	// Initialize SRS handler
	srsStore := srs.NewQueriesStore(q)
	srsHandler := srs.NewSRSHandler(srsStore)

	// Initialize progress handler
	progressStore := progress.NewQueriesStore(q)
	progressHandler := progress.NewProgressHandler(progressStore)

	// Initialize challenges handler
	challengeStore := challenges.NewQueriesStore(q)
	challengeHandler := challenges.NewChallengeHandler(challengeStore, encSvc)

	// Telegram bot (optional — enabled when TELEGRAM_BOT_TOKEN is set)
	telegramToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramBotUsername := os.Getenv("TELEGRAM_BOT_USERNAME")
	var telegramStore *telegram.QueriesStore
	if telegramToken != "" {
		telegramStore = telegram.NewQueriesStore(q)
		api := telegram.NewAPI(telegramToken)
		bot := telegram.NewBot(api, telegramStore, encSvc)
		challengeHandler.SetNotifier(bot)
		go telegram.Run(ctx, api, bot)
		go telegram.RunReminderLoop(ctx, bot)
		log.Info().Msg("telegram bot started")
	}

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

		// Protected: requires valid JWT (any role)
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))
			r.Get("/me", authHandler.Me)
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
		r.Post("/{id}/login-as", childHandler.LoginAs)
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

	// Content routes (child only)
	r.Route("/api/subjects", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Use(auth.ChildOnly)
		r.Get("/", contentHandler.ListSubjects)
		r.Get("/{subjectSlug}/topics", contentHandler.ListTopics)
	})
	r.Route("/api/topics", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Use(auth.ChildOnly)
		r.Get("/{subjectSlug}/{topicSlug}/exercises", contentHandler.ListExercises)
	})

	// Reviews (spaced repetition) routes (child only)
	r.Route("/api/reviews", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Use(auth.ChildOnly)
		r.Get("/due", srsHandler.ListDueReviews)
	})

	// Progress routes (child — own progress)
	r.Route("/api/progress", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Use(auth.ChildOnly)
		r.Get("/", progressHandler.GetStudentProgress)
	})

	// Parent progress routes (parent — child's progress)
	r.Route("/api/children/{studentId}/progress", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Use(auth.ParentOnly)
		r.Get("/", progressHandler.GetChildProgress)
		r.Get("/sessions", progressHandler.ListChildSessions)
	})

	// Challenge routes (parent + child — role handled inside handler)
	r.Route("/api/challenges", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Get("/", challengeHandler.List)
		r.Post("/", challengeHandler.Create)
		r.Get("/{id}", challengeHandler.Get)
		r.Patch("/{id}/publish", func(w http.ResponseWriter, req *http.Request) {
			if auth.GetRoleFromContext(req.Context()) != "parent" {
				http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
				return
			}
			challengeHandler.Publish(w, req)
		})
		r.Delete("/{id}", func(w http.ResponseWriter, req *http.Request) {
			if auth.GetRoleFromContext(req.Context()) != "parent" {
				http.Error(w, `{"error":"Forbidden"}`, http.StatusForbidden)
				return
			}
			challengeHandler.Delete(w, req)
		})
	})

	// Session/chat routes (child only)
	r.Route("/api/sessions", func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Use(auth.ChildOnly)
		r.Post("/", chatHandler.CreateSession)
		r.Get("/{id}", chatHandler.GetSession)
		r.Post("/{id}/messages", chatHandler.SendMessage)
		r.Get("/{id}/messages", chatHandler.ListMessages)
		r.Post("/{id}/end", chatHandler.EndSession)
	})

	// Telegram link route (child only) — only when the bot is enabled
	if telegramToken != "" {
		linkHandler := telegram.NewLinkHandler(telegramStore, telegramBotUsername)
		r.Route("/api/telegram", func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))
			r.Use(auth.ChildOnly)
			r.Post("/link-code", linkHandler.CreateLinkCode)
		})
	}

	addr := ":" + port
	log.Info().Str("addr", addr).Msg("starting server")

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 5 * time.Minute, // Extended for SSE streaming
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
