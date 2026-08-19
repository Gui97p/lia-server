package transport

import (
	"log/slog"
	"net/http"
	"time"

	_ "github.com/Gui97p/lia-server/docs/swagger"
	"github.com/Gui97p/lia-server/internal/agent"
	"github.com/Gui97p/lia-server/internal/audio"
	"github.com/Gui97p/lia-server/internal/config"
	"github.com/Gui97p/lia-server/internal/messages"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/Gui97p/lia-server/internal/tasks"
	"github.com/Gui97p/lia-server/internal/users"
)

type Deps struct {
	UsersStore        users.Store
	MessagesStore     messages.Store
	TasksStore        tasks.Store
	TranscriberClient audio.TranscriberClient
	TTSClient         audio.TTSClient
	Hub               *session.Hub
	PlanningQueue     *agent.PlanningQueueManager
	Executor          *agent.Executor
}

type Server struct {
	logger *slog.Logger
	router *router
	hub    *session.Hub

	usersStore    users.Store
	messagesStore messages.Store
	tasksStore    tasks.Store

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

		usersStore:    deps.UsersStore,
		messagesStore: deps.MessagesStore,
		tasksStore:    deps.TasksStore,

		transcriberClient: deps.TranscriberClient,
		ttsClient:         deps.TTSClient,

		planningQueue: deps.PlanningQueue,
		executor:      deps.Executor,

		jwtSecret:     cfg.JWTSecret,
		encryptionKey: cfg.EncryptionKey,
	}

	mux := http.NewServeMux()
	mux.Handle("GET /docs/swagger/", http.StripPrefix("/docs/swagger/", http.FileServer(http.Dir("docs/swagger"))))
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /docs/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html>
	<html>
	<head><title>Lia API</title></head>
	<body>
	<script id="api-reference" data-url="/docs/swagger/swagger.json"></script>
	<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
	</body>
	</html>`))
	})

	mux.HandleFunc("POST /audio/tts", s.withAuth(s.handleAudioTTS))
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
