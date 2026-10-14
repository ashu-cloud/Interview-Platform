package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/ai-interview/services/turn-projector/internal/config"
	"github.com/yourusername/ai-interview/services/turn-projector/internal/repository"
	"github.com/yourusername/ai-interview/services/turn-projector/internal/worker"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load(os.Getenv)
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to create database pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	worker.NewTurnProjector(cfg, repository.NewTurnRepository(pool), logger).Run(ctx)
	logger.Info("turn-projector stopped")
}
