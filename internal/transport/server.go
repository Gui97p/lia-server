package transport

import (
	"log/slog"
	"net/http"
	"time"

	_ "github.com/Gui97p/lia-server/docs/swagger"
	"github.com/Gui97p/lia-server/internal/agent"
	"github.com/Gui97p/lia-server/internal/audio"
	behaviorrules "github.com/Gui97p/lia-server/internal/behavior_rules"
	"github.com/Gui97p/lia-server/internal/config"
	"github.com/Gui97p/lia-server/internal/memories"
	"github.com/Gui97p/lia-server/internal/messages"
	"github.com/Gui97p/lia-server/internal/providers"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/Gui97p/lia-server/internal/tasks"
	"github.com/Gui97p/lia-server/internal/users"
)

type Deps struct {
	UsersStore         users.Store
	ProvidersStore     providers.Store
	MessagesStore      messages.Store
	TasksStore         tasks.Store
	MemoriesStore      memories.Store
	BehaviorRulesStore behaviorrules.Store

	TranscriberClient audio.TranscriberClient
	TTSClient         audio.TTSClient

	Hub           *session.Hub
	PlanningQueue *agent.PlanningQueueManager
	Executor      *agent.Executor
}

type Server struct {
	logger *slog.Logger
	router *router
	hub    *session.Hub

	usersStore         users.Store
	providersStore     providers.Store
	messagesStore      messages.Store
	tasksStore         tasks.Store
	memoriesStore      memories.Store
	behaviorRulesStore behaviorrules.Store

	transcriberClient audio.TranscriberClient
	ttsClient         audio.TTSClient

	planningQueue *agent.PlanningQueueManager
	executor      *agent.Executor

	jwtSecret     []byte
	encryptionKey []byte
}

func New(cfg *config.Config, logger *slog.Logger, deps Deps) *http.Server {
	s := &Server{
		logger: logger,
		router: newRouter(),
		hub:    deps.Hub,

		usersStore:         deps.UsersStore,
		providersStore:     deps.ProvidersStore,
		messagesStore:      deps.MessagesStore,
		tasksStore:         deps.TasksStore,
		memoriesStore:      deps.MemoriesStore,
		behaviorRulesStore: deps.BehaviorRulesStore,

		transcriberClient: deps.TranscriberClient,
		ttsClient:         deps.TTSClient,

		planningQueue: deps.PlanningQueue,
		executor:      deps.Executor,

		jwtSecret:     cfg.JWTSecret,
		encryptionKey: cfg.EncryptionKey,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/docs/", http.StatusFound)
	})

	mux.Handle("GET /docs/swagger/", http.StripPrefix("/docs/swagger/", http.FileServer(http.Dir("docs/swagger"))))
	mux.Handle("GET /docs/asyncapi/", http.StripPrefix("/docs/asyncapi/", http.FileServer(http.Dir("docs/asyncapi"))))
	mux.Handle("GET /docs/", http.StripPrefix("/docs/", http.FileServer(http.Dir("docs/site"))))

	mux.HandleFunc("GET /health", s.handleHealth)

	mux.HandleFunc("POST /audio/speak", s.withAuth(s.handleAudioTTS))
	mux.HandleFunc("POST /audio/transcribe", s.withAuth(s.handleAudioTranscribe))

	setupMessageHandlers(s)
	setupToolHandlers(s)

	mux.HandleFunc("GET /ws", s.handleWS)

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
