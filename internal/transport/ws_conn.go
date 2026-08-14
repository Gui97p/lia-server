package transport

import (
	"net/http"

	"github.com/coder/websocket"
)

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		s.logger.Error("websocket accept failed", "error", err)
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			break
		}
		if err := s.router.dispatch(ctx, conn, data); err != nil {
			break
		}
	}
}
