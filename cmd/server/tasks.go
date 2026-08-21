package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Gui97p/lia-server/internal/tasks"
)

func recoverStaleTasks(ctx context.Context, tasksStore tasks.Store, logger *slog.Logger) {
	recovered, err := tasksStore.RecoverStaleTasks(ctx)
	if err != nil {
		logger.Error("failed to recover stale tasks", "error", err)
		os.Exit(1)
	}
	if recovered > 0 {
		logger.Warn("recovered stale tasks on startup", "count", recovered)
	}
}
