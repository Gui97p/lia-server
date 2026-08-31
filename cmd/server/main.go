package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Gui97p/lia-server/internal/agent"
	"github.com/Gui97p/lia-server/internal/audio"
	behaviorrules "github.com/Gui97p/lia-server/internal/behavior_rules"
	"github.com/Gui97p/lia-server/internal/capabilities"
	"github.com/Gui97p/lia-server/internal/config"
	"github.com/Gui97p/lia-server/internal/db"
	"github.com/Gui97p/lia-server/internal/memories"
	"github.com/Gui97p/lia-server/internal/messages"
	"github.com/Gui97p/lia-server/internal/providers"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/Gui97p/lia-server/internal/tasks"
	"github.com/Gui97p/lia-server/internal/transport"
	"github.com/Gui97p/lia-server/internal/users"
)

// @title          Lia Server API
// @version        0.5
// @description    Server Backend da Lia
// @host           localhost:3000
// @BasePath       /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
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

	tasksStore := tasks.NewPostgresStore(pool)
	recoverStaleTasks(ctx, tasksStore, logger)

	messagesStore := messages.NewPostgresStore(pool)
	memoriesStore := memories.NewPostgresStore(pool)
	capabilitiesStore := capabilities.NewPostgresStore(pool)

	toolRegistry := newToolRegistry(cfg, memoriesStore)

	app := transport.New(cfg, logger, transport.Deps{
		UsersStore:         users.NewPostgresStore(pool),
		ProvidersStore:     providers.NewPostgresStore(pool),
		MessagesStore:      messagesStore,
		TasksStore:         tasksStore,
		MemoriesStore:      memoriesStore,
		BehaviorRulesStore: behaviorrules.NewPostgresStore(pool),

		TranscriberClient: audio.NewGroqTranscriber("whisper-large-v3-turbo", logger),
		TTSClient:         audio.NewEdgeTTSClient("pt-BR-FranciscaNeural"),

		Hub:           session.NewHub(),
		PlanningQueue: agent.NewPlanningQueue(agent.NewPlanner(newLLMRouter(logger), toolRegistry, capabilitiesStore, logger)),
		Executor:      agent.NewExecutor(messagesStore, toolRegistry, logger),
	})

	logger.Info("server starting", "port", cfg.Port)
	if err := app.ListenAndServe(); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
