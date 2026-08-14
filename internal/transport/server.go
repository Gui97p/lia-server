package transport

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Gui97p/lia-server/internal/config"
)

type Server struct {
	logger *slog.Logger
	router *router
}

func New(cfg *config.Config, logger *slog.Logger) *http.Server {
	s := &Server{logger: logger, router: newRouter()}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /ws", s.handleWS)

	return &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
}
