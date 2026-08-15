package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Gui97p/lia-server/internal/auth"
	"github.com/Gui97p/lia-server/internal/crypto"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type ConnectionContext struct {
	UserID       string
	Username     string
	TrustLevel   auth.TrustLevel
	GroqAPIKey   string
	Capabilities []string
}

func (s *Server) handshake(ctx context.Context, conn *websocket.Conn) (*ConnectionContext, error) {
	_, data, err := conn.Read(ctx)
	if err != nil {
		return nil, err
	}

	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		sendError(ctx, conn, "invalid request")
		return nil, err
	}

	if env.Event != "auth" {
		sendError(ctx, conn, "first message must be auth")
		return nil, fmt.Errorf("expected auth event")
	}

	var authPayload struct {
		Token        string   `json:"token"`
		Capabilities []string `json:"capabilities"`
	}
	if err := json.Unmarshal(env.Payload, &authPayload); err != nil {
		sendError(ctx, conn, "invalid payload")
		return nil, err
	}

	claims, err := auth.ParseToken(s.jwtSecret, authPayload.Token)
	if err != nil {
		sendError(ctx, conn, "invalid token")
		return nil, err
	}

	if authPayload.Capabilities == nil {
		sendError(ctx, conn, "invalid capabilities")
		return nil, errors.New("invalid capabilities")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		sendError(ctx, conn, "invalid uuid")
		return nil, err
	}

	user, err := s.usersStore.GetByID(ctx, userID)
	if err != nil {
		sendError(ctx, conn, "user not found")
		return nil, err
	}

	if user.TokenVersion != claims.TokenVersion {
		sendError(ctx, conn, "token revoked")
		return nil, errors.New("token version mismatch")
	}

	if user.GroqAPIKeyEncrypted == nil {
		sendError(ctx, conn, "invalid api key")
		return nil, errors.New("api key not set")
	}

	groqAPIKey, err := crypto.Decrypt(*user.GroqAPIKeyEncrypted, s.encryptionKey)
	if err != nil {
		sendError(ctx, conn, "invalid api key")
		return nil, err
	}

	return &ConnectionContext{
		UserID:       claims.UserID,
		Username:     user.Username,
		TrustLevel:   claims.TrustLevel,
		GroqAPIKey:   groqAPIKey,
		Capabilities: authPayload.Capabilities,
	}, nil
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		s.logger.Error("websocket accept failed", "error", err)
		return
	}

	ctx := r.Context()
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	connCtx, err := s.handshake(timeoutCtx, conn)
	if err != nil {
		conn.CloseNow()
		return
	}

	defer conn.CloseNow()
	err = sendEvent(ctx, conn, "auth.ok", struct{}{})
	if err != nil {
		return
	}

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			break
		}
		if err := s.router.dispatch(ctx, conn, connCtx, data); err != nil {
			break
		}
	}
}
