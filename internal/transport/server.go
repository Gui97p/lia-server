package transport

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Gui97p/lia-server/internal/config"
	"github.com/Gui97p/lia-server/internal/messages"
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/Gui97p/lia-server/internal/users"
)

type Deps struct {
	UsersStore    users.Store
	MessagesStore messages.Store
	Hub           *session.Hub
}

type Server struct {
	logger *slog.Logger
	router *router
	hub    *session.Hub

	usersStore    users.Store
	messagesStore messages.Store

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

		jwtSecret:     cfg.JWTSecret,
		encryptionKey: cfg.EncryptionKey,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)

	setupMessageHandlers(s)

	mux.HandleFunc("GET /ws", s.handleWS)

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
