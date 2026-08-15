package transport

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Gui97p/lia-server/internal/config"
	"github.com/Gui97p/lia-server/internal/users"
)

type Deps struct {
	UsersStore users.Store
}

type Server struct {
	logger *slog.Logger
	router *router
	hub    *Hub

	usersStore users.Store

	jwtSecret     []byte
	encryptionKey []byte
}

func New(cfg *config.Config, logger *slog.Logger, deps Deps) *http.Server {
	s := &Server{
		logger:        logger,
		router:        newRouter(),
		hub:           newHub(),
		jwtSecret:     cfg.JWTSecret,
		encryptionKey: cfg.EncryptionKey,
		usersStore:    deps.UsersStore,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /ws", s.handleWS)

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
