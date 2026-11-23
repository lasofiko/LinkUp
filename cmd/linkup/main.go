package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/lasofiko/LinkUp/internal/config"
	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/handler"
	"github.com/lasofiko/LinkUp/internal/matching"
	"github.com/lasofiko/LinkUp/internal/profileclient"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		slog.Error("LinkUp exited with error", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	users := memory.NewUserRepository()
	for _, user := range demoUsers {
		if err := users.Create(user); err != nil {
			return fmt.Errorf("seed demo user %d: %w", user.ID, err)
		}
	}
	profiles := memory.NewProfiles(demoProfiles)
	pairStore := memory.NewPairStore()

	profileClient, err := profileclient.NewClient(
		cfg.ProfileBaseURL,
		&http.Client{Timeout: cfg.ProfileTimeout},
		cfg.ProfileMaxBody,
	)
	if err != nil {
		return fmt.Errorf("create profile API client: %w", err)
	}

	reactService := usecase.NewReactService(users, profileClient, pairStore)
	discoveryService := usecase.NewPairDiscoveryService(
		matching.NewDiscoveryService(profiles),
		pairStore,
	)

	mux := http.NewServeMux()
	mux.Handle("GET /v1/me/discovery", handler.DiscoveryHandler(discoveryService))
	mux.Handle("PUT /v1/me/likes/{targetID}", handler.LikeHandler(reactService))
	mux.Handle("GET /v1/me/pairs/{targetID}", handler.PairHandler(pairStore))

	var root http.Handler = mux
	root = handler.IdentityMiddleware(root)
	root = handler.TimeoutMiddleware(cfg.CommandTimeout)(root)
	root = handler.LoggingMiddleware(logger)(root)
	root = handler.RequestIDMiddleware(root)

	server := &http.Server{
		Addr:              cfg.LinkUp.Address,
		Handler:           root,
		ReadHeaderTimeout: cfg.LinkUp.ReadHeaderTimeout,
		ReadTimeout:       cfg.LinkUp.ReadTimeout,
		WriteTimeout:      cfg.LinkUp.WriteTimeout,
		IdleTimeout:       cfg.LinkUp.IdleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("LinkUp listening",
		slog.String("address", cfg.LinkUp.Address),
		slog.String("profile_base_url", cfg.ProfileBaseURL),
		slog.Duration("command_timeout", cfg.CommandTimeout),
	)

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("linkup server: %w", err)

	case <-ctx.Done():
		stop()
		logger.Info("shutdown signal received, draining active requests",
			slog.Duration("grace_period", cfg.LinkUp.ShutdownTimeout),
		)

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.LinkUp.ShutdownTimeout)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Warn("grace period exceeded, forcing server close", slog.Any("error", err))
			_ = server.Close()
			return fmt.Errorf("linkup graceful shutdown: %w", err)
		}

		if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("linkup server: %w", err)
		}

		logger.Info("LinkUp stopped")
		return nil
	}
}

var demoUsers = []domain.User{
	{ID: 1, Name: "Соня"},
	{ID: 2, Name: "Арина"},
	{ID: 3, Name: "Ваня"},
	{ID: 4, Name: "Оля"},
	{ID: 5, Name: "Demo"},
}

var demoProfiles = []domain.Profile{
	{UserID: 1, City: "Moscow", Interests: []domain.Interest{"Go", "Music"}, Available: true},
	{UserID: 2, City: "Moscow", Interests: []domain.Interest{"Go"}, Available: true},
	{UserID: 3, City: "Moscow", Available: false},
	{UserID: 4, City: "Moscow", Available: true},
	{UserID: 5, City: "Moscow", Available: true},
}
