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
	"github.com/FacileStudio/Casier/apps/api/internal/mailer"
	"github.com/FacileStudio/Casier/apps/api/internal/webhook"
	"github.com/FacileStudio/Casier/apps/api/migrations"
	"github.com/FacileStudio/Casier/apps/api/modules/auth"
	"github.com/FacileStudio/Casier/apps/api/modules/environments"
	"github.com/FacileStudio/Casier/apps/api/modules/projects"
	"github.com/FacileStudio/Casier/apps/api/modules/secrets"

	"github.com/FacileStudio/Journal/sdk/journal"
	"github.com/FacileStudio/porte/local"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/healthcheck"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/FacileStudio/tronc/logger"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/migrate"
	"github.com/FacileStudio/tronc/spa"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func redirectToProjects(w http.ResponseWriter, request *http.Request) {
	target := "/projects" + strings.TrimPrefix(request.URL.Path, "/spaces")
	http.Redirect(w, request, target, http.StatusMovedPermanently)
}

func registry() documentation.Response {
	return documentation.Response{
		Modules: []documentation.Module{
			auth.Documentation,
			projects.Documentation,
			environments.Documentation,
			secrets.Documentation,
		},
	}
}

func refConfig() apiref.Config {
	return apiref.Config{
		Title:       "Casier API",
		Description: "Self-hosted secrets and environment variable manager for teams.",
		Servers:     []string{"/api"},
		Registry:    registry(),
	}
}

func buildAuth(ctx context.Context, db *gorm.DB, appEnv env.Config, appLogger *slog.Logger) (*session.Manager, *local.Kit, *oidc.Kit, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, nil, err
	}
	store := portepg.New(sqlDB)
	cfg := appEnv.Porte()

	sessions, err := session.New(cfg, session.Deps{Sessions: store.Sessions(), Logger: appLogger})
	if err != nil {
		return nil, nil, nil, err
	}

	var kit *oidc.Kit
	if cfg.Issuer != "" {
		if err := auth.NewUserStore(db).BackfillAvatars(ctx, cfg.Issuer); err != nil {
			return nil, nil, nil, err
		}
		kit, err = oidc.New(ctx, cfg, oidc.Deps{
			Users:      auth.NewUserStore(db),
			Identities: store.Identities(),
			Sessions:   sessions,
			Codes:      store.LoginCodes(),
			Logger:     appLogger,
		})
		if err != nil {
			return nil, nil, nil, err
		}
	}

	var passwords *local.Kit
	passwords, err = local.New(local.Config{AllowRegistration: !appEnv.SSOOnly, MinPasswordLength: 8}, local.Deps{
		Users:      auth.NewUserStore(db),
		Identities: store.Identities(),
		Sessions:   sessions,
		Logger:     appLogger,
		Count:      auth.NewUserStore(db).CountUsers,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return sessions, passwords, kit, nil
}

var authService *auth.Service

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

	migrateConfig := migrate.Config{DB: sqlDB, FS: migrations.FS, Logger: appLogger}
	if handled, err := migrate.Command(context.Background(), os.Args, migrateConfig); handled {
		if err != nil {
			appLogger.Error("migrate", slog.Any("error", err))
			return 1
		}
		return 0
	}
	if err := migrate.Run(context.Background(), migrateConfig); err != nil {
		appLogger.Error("failed to run migrations", slog.Any("error", err))
		return 1
	}

	shutdownCtx, cancelCleanup := context.WithCancel(context.Background())
	defer cancelCleanup()

	appMailer, err := mailer.New(appEnv.SMTP)
	if err != nil {
		appLogger.Error("invalid SMTP configuration", slog.Any("error", err))
		return 1
	}

	sessions, passwords, kit, err := buildAuth(shutdownCtx, db, appEnv, appLogger)
	if err != nil {
		appLogger.Error("failed to initialise porte", slog.Any("error", err))
		return 1
	}

	store := portepg.New(sqlDB)
	authService = auth.NewService(db, keyring, appLogger, appMailer, appEnv.PublicURL, sessions, passwords, store.Identities())
	authService.StartSessionCleanup(shutdownCtx)

	webhookDispatcher := webhook.NewDispatcher(db, appLogger)
	auditLogger := audit.NewLogger(db, appLogger)
	projectsService := projects.NewService(db, keyring, auditLogger)
	environmentsService := environments.NewService(db, projectsService)
	secretsService := secrets.NewService(db, appLogger, projectsService, environmentsService, projectsService, webhookDispatcher, auditLogger)
	secretsService.StartTrashPurge(shutdownCtx)

	router := httpx.NewRouter(httpx.Config{
		TrustedProxies: appEnv.TrustedProxies,
		CDNProxies:     appEnv.CDNProxies,
		CDNHeader:      appEnv.CDNHeader,
		Logger:         appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins:   appEnv.CORSAllowedOrigins,
			AllowCredentials: true,
		},
	})

	health.Mount(router, health.DB(sqlDB))
	apiref.Mount(router, refConfig())

	router.Route("/api", func(r chi.Router) {
		auth.RegisterRoutes(r, sessions, kit, authService, appEnv)
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
