package main

import (
	"context"
	"log/slog"
	"os"

	behaviorrules "github.com/Gui97p/lia-server/internal/behavior_rules"
	"github.com/Gui97p/lia-server/internal/capabilities"
	"github.com/Gui97p/lia-server/internal/config"
	"github.com/Gui97p/lia-server/internal/db"
	"github.com/Gui97p/lia-server/internal/memories"
	"github.com/Gui97p/lia-server/internal/messages"
	"github.com/Gui97p/lia-server/internal/providers"
	"github.com/Gui97p/lia-server/internal/users"
)

var (
	usersStore         users.Store
	providersStore     providers.Store
	messagesStore      messages.Store
	behaviorRulesStore behaviorrules.Store
	memoriesStore      memories.Store
	capabilitiesStore  capabilities.Store
	encryptionKey      []byte
	jwtSecret          []byte
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

	usersStore = users.NewPostgresStore(pool)
	providersStore = providers.NewPostgresStore(pool)
	messagesStore = messages.NewPostgresStore(pool)
	behaviorRulesStore = behaviorrules.NewPostgresStore(pool)
	memoriesStore = memories.NewPostgresStore(pool)
	capabilitiesStore = capabilities.NewPostgresStore(pool)
	encryptionKey = cfg.EncryptionKey
	jwtSecret = cfg.JWTSecret

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
