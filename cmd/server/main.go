package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/buan1027/workshop/internal/auth"
	"github.com/buan1027/workshop/internal/config"
	"github.com/buan1027/workshop/internal/database"
	httpapi "github.com/buan1027/workshop/internal/http"
	"github.com/buan1027/workshop/internal/repository"
	"github.com/buan1027/workshop/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("database pool could not be created", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if cfg.ResetDatabaseOnStart {
		seedCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := database.ResetDemoData(seedCtx, pool); err != nil {
			logger.Error("demo data could not be reset", "error", err)
			os.Exit(1)
		}
		logger.Info("demo data reset")
	}

	repo := repository.NewPostgresGebrauchtwagenRepository(pool)
	gebrauchtwagenService := service.NewGebrauchtwagenService(repo)
	authorizer := buildAuthorizer(ctx, cfg, logger)
	handler := httpapi.NewRouter(httpapi.Dependencies{
		DB:         pool,
		Repository: repo,
		Service:    gebrauchtwagenService,
		Authorizer: authorizer,
		Logger:     logger,
	})

	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Info("server listening", "addr", cfg.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server stopped unexpectedly", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}

func buildAuthorizer(ctx context.Context, cfg config.Config, logger *slog.Logger) auth.Authorizer {
	fallback := auth.NewAdminTokenAuthorizer(cfg.AdminToken)
	if cfg.AuthMode != "keycloak" {
		return fallback
	}

	if cfg.KeycloakIssuerURL == "" {
		logger.Warn("keycloak auth mode configured without issuer url; falling back to admin token")
		return fallback
	}

	keycloak, err := auth.NewKeycloakAuthorizer(ctx, auth.KeycloakConfig{
		IssuerURL: cfg.KeycloakIssuerURL,
		ClientID:  cfg.KeycloakClientID,
	})
	if err != nil {
		logger.Warn("keycloak authorizer could not be initialized; falling back to admin token", "error", err)
		return fallback
	}

	logger.Info("keycloak authorizer enabled", "issuer", cfg.KeycloakIssuerURL, "clientId", cfg.KeycloakClientID)
	return keycloak
}
