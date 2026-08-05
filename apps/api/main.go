package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/FacileStudio/Casier/apps/api/internal/audit"
	"github.com/FacileStudio/Casier/apps/api/internal/crypto"
	"github.com/FacileStudio/Casier/apps/api/internal/database"
	documentation "github.com/FacileStudio/Casier/apps/api/internal/documentation"
	"github.com/FacileStudio/Casier/apps/api/internal/env"
	"github.com/FacileStudio/Casier/apps/api/internal/webhook"
	"github.com/FacileStudio/Casier/apps/api/modules/auth"
	"github.com/FacileStudio/Casier/apps/api/modules/environments"
	"github.com/FacileStudio/Casier/apps/api/modules/projects"
	"github.com/FacileStudio/Casier/apps/api/modules/secrets"
	"github.com/FacileStudio/Casier/apps/api/schemas"

	"github.com/FacileStudio/Journal/sdk/journal"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/healthcheck"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/FacileStudio/tronc/logger"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/spa"
	"github.com/go-chi/chi/v5"
)

func redirectToProjects(w http.ResponseWriter, request *http.Request) {
	target := "/projects" + strings.TrimPrefix(request.URL.Path, "/spaces")
	http.Redirect(w, request, target, http.StatusMovedPermanently)
}

func main() {
	if healthcheck.Handle(os.Args) {
		return
	}

	os.Exit(run())
}

func run() int {
	appEnv, err := env.Load()
	appLogger := logger.New(logger.Config{})
	if err != nil {
		appLogger.Error("failed to load config", slog.Any("error", err))
		return 1
	}
	var journalClient *journal.Client
	appLogger = logger.New(logger.Config{
		Level: appEnv.LogLevel,
		Wrap: func(handler slog.Handler) slog.Handler {
			if appEnv.JournalURL == "" || appEnv.JournalToken == "" {
				return handler
			}
			journalClient = journal.New(journal.Config{URL: appEnv.JournalURL, Token: appEnv.JournalToken})
			return journal.NewHandler(journalClient, handler)
		},
	})
	if journalClient != nil {
		defer journalClient.Close()
	}

	encKey, err := crypto.KeyFromBase64(appEnv.EncryptionKey)
	if err != nil {
		appLogger.Error("invalid encryption key", slog.Any("error", err))
		return 1
	}
	keyring := crypto.NewKeyring(encKey)

	db, err := database.Open(appEnv.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to open database", slog.Any("error", err))
		return 1
	}

	if err := schemas.Migrate(db); err != nil {
		appLogger.Error("failed to run migrations", slog.Any("error", err))
		return 1
	}
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("failed to access database handle", slog.Any("error", err))
		return 1
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			appLogger.Error("failed to close database", slog.Any("error", err))
		}
	}()

	shutdownCtx, cancelCleanup := context.WithCancel(context.Background())
	defer cancelCleanup()

	authService := auth.NewService(db, keyring, appLogger)
	authService.StartSessionCleanup(shutdownCtx)
	webhookDispatcher := webhook.NewDispatcher(db, appLogger)
	auditLogger := audit.NewLogger(db, appLogger)
	projectsService := projects.NewService(db, keyring, auditLogger)
	environmentsService := environments.NewService(db, projectsService)
	secretsService := secrets.NewService(db, appLogger, projectsService, environmentsService, projectsService, webhookDispatcher, auditLogger)

	docs := documentation.Response{
		Modules: []documentation.Module{
			auth.Documentation,
			projects.Documentation,
			environments.Documentation,
			secrets.Documentation,
		},
	}

	router := httpx.NewRouter(httpx.Config{
		Logger: appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins:   appEnv.CORSAllowedOrigins,
			AllowCredentials: true,
		},
	})

	health.Mount(router, health.DB(sqlDB))
	router.Get("/docs", func(w http.ResponseWriter, request *http.Request) {
		httpjson.WriteJSON(w, http.StatusOK, docs)
	})

	router.Route("/api", func(r chi.Router) {
		auth.RegisterRoutes(r, authService, appEnv)
		projects.RegisterRoutes(r, projectsService, authService, appEnv)
		environments.RegisterRoutes(r, environmentsService, authService)
		secrets.RegisterRoutes(r, secretsService, authService)
	})

	router.Get("/spaces", redirectToProjects)
	router.Get("/spaces/*", redirectToProjects)

	clientDir := spa.DirFromEnv()
	if spa.Available(clientDir) {
		router.Handle("/*", spa.Handler(spa.Config{Dir: clientDir}))
		appLogger.Info("serving client", slog.String("dir", clientDir))
	}

	addr := ":" + strconv.Itoa(appEnv.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	appLogger.Info("server starting", slog.String("addr", addr))
	select {
	case err := <-serverErrCh:
		if !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("server stopped", slog.Any("error", err))
			return 1
		}
	case <-shutdownSignal.Done():
		appLogger.Info("server shutting down")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownContext); err != nil {
			appLogger.Error("server shutdown failed", slog.Any("error", err))
			return 1
		}
		appLogger.Info("server stopped")
	}

	return 0
}
