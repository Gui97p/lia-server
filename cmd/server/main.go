package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Gui97p/lia-server/internal/config"
	"github.com/Gui97p/lia-server/internal/db"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	pool, err := db.Connect(ctx, *cfg)
	if err != nil {
		logger.Error("failed to connect to db", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("Server starting", "port", cfg.Port)
}
