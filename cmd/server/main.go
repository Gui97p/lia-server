package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/Gui97p/lia-server/internal/agent"
	"github.com/Gui97p/lia-server/internal/audio"
	"github.com/Gui97p/lia-server/internal/config"
	"github.com/Gui97p/lia-server/internal/db"
	"github.com/Gui97p/lia-server/internal/llm"
	"github.com/Gui97p/lia-server/internal/messages"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/Gui97p/lia-server/internal/tasks"
	"github.com/Gui97p/lia-server/internal/transport"
	"github.com/Gui97p/lia-server/internal/users"
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

	messagesStore := messages.NewPostgresStore(pool)

	app := transport.New(cfg, logger, transport.Deps{
		UsersStore:        users.NewPostgresStore(pool),
		MessagesStore:     messagesStore,
		TasksStore:        tasks.NewPostgresStore(pool),
		TranscriberClient: audio.NewGroqTranscriber("whisper-large-v3-turbo", logger),
		TTSClient:         audio.NewEdgeTTSClient("pt-BR-FranciscaNeural"),
		Hub:               session.NewHub(),
		PlanningQueue:     agent.NewPlanningQueue(agent.NewPlanner(llm.NewGroqClient("qwen/qwen3.6-27b", logger))),
		Executor:          agent.NewExecutor(messagesStore),
	})

	logger.Info("server starting", "port", cfg.Port)
	if err := app.ListenAndServe(); err != nil {
		logger.Error("server failed", "error", err)
		os.Exit(1)
	}
}
