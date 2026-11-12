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

	"github.com/GuiJeff0/minitest/microservices/minitest-api/internal/config"
	"github.com/GuiJeff0/minitest/microservices/minitest-api/internal/database"
	"github.com/GuiJeff0/minitest/microservices/minitest-api/internal/handler"
	"github.com/GuiJeff0/minitest/microservices/minitest-api/internal/repository"
	"github.com/GuiJeff0/minitest/microservices/minitest-api/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.NewPool(ctx, cfg.DSN())
	if err != nil {
		logger.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.HealthCheck(ctx, pool); err != nil {
		logger.Warn("initial database ping failed; readiness probe will report not ready until available", "error", err)
	}

	repo := repository.NewPostgresItemRepository(pool)
	svc := service.NewItemService(repo)
	readyFn := func(checkCtx context.Context) error {
		return database.HealthCheck(checkCtx, pool)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           handler.NewRouter(svc, readyFn, cfg.AppEnv, logger),
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("starting http server", "port", cfg.HTTPPort, "env", cfg.AppEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("http server error", "error", err)
			os.Exit(1)
		}
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	shutdown(shutdownCtx, srv, pool, logger)
}

func shutdown(ctx context.Context, srv *http.Server, pool *pgxpool.Pool, logger *slog.Logger) {
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	} else {
		logger.Info("http server stopped gracefully")
	}
	pool.Close()
}
