package main

import (
	"log/slog"
	"os"

	"github.com/Gui97p/lia-server/internal/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	env, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config", "error", err.Error())
		os.Exit(1)
	}

	logger.Info("Server starting", "port", env.Port)
}
