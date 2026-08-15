package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Gui97p/lia-server/internal/auth"
	"github.com/Gui97p/lia-server/internal/crypto"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

func (s *Server) handshake(ctx context.Context, conn *websocket.Conn) (*Session, error) {
	session := Session{}
	var writeMu sync.Mutex
	session.Writer = func(ctx context.Context, event string, payload any) error {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		data, err := json.Marshal(Envelope{Event: event, Payload: jsonPayload})
		if err != nil {
			return err
		}

		writeMu.Lock()
		defer writeMu.Unlock()
		return conn.Write(ctx, websocket.MessageText, data)
	}

	_, data, err := conn.Read(ctx)
	if err != nil {
		return nil, err
	}

	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		sendError(ctx, &session, "invalid request")
		return nil, err
	}

	if env.Event != "auth" {
		sendError(ctx, &session, "first message must be auth")
		return nil, fmt.Errorf("expected auth event")
	}

	var authPayload struct {
		Token        string   `json:"token"`
		Capabilities []string `json:"capabilities"`
	}
	if err := json.Unmarshal(env.Payload, &authPayload); err != nil {
		sendError(ctx, &session, "invalid payload")
		return nil, err
	}

	claims, err := auth.ParseToken(s.jwtSecret, authPayload.Token)
	if err != nil {
		sendError(ctx, &session, "invalid token")
		return nil, err
	}

	if authPayload.Capabilities == nil {
		sendError(ctx, &session, "invalid capabilities")
		return nil, errors.New("invalid capabilities")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		sendError(ctx, &session, "invalid uuid")
		return nil, err
	}

	user, err := s.usersStore.GetByID(ctx, userID)
	if err != nil {
		sendError(ctx, &session, "user not found")
		return nil, err
	}

	if user.TokenVersion != claims.TokenVersion {
		sendError(ctx, &session, "token revoked")
		return nil, errors.New("token version mismatch")
	}

	if user.GroqAPIKeyEncrypted == nil {
		sendError(ctx, &session, "invalid api key")
		return nil, errors.New("api key not set")
	}

	groqAPIKey, err := crypto.Decrypt(*user.GroqAPIKeyEncrypted, s.encryptionKey)
	if err != nil {
		sendError(ctx, &session, "invalid api key")
		return nil, err
	}

	session.ConnID = uuid.New()
	session.UserID = userID
	session.Username = user.Username
	session.TrustLevel = claims.TrustLevel
	session.GroqAPIKey = groqAPIKey
	session.Capabilities = authPayload.Capabilities

	return &session, nil
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		s.logger.Error("websocket accept failed", "error", err)
		return
	}

	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	session, err := s.handshake(timeoutCtx, conn)
	cancel()
	if err != nil {
		conn.CloseNow()
		return
	}
	defer conn.CloseNow()

	s.hub.Register(session.ConnID, session)
	defer s.hub.Unregister(session.ConnID)

	type AuthResponse struct {
		ConnID string `json:"conn_id"`
	}
	err = session.Writer(ctx, "auth.ok", AuthResponse{ConnID: session.ConnID.String()})
	if err != nil {
		return
	}

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			break
		}
		if err := s.router.dispatch(ctx, session, data); err != nil {
			break
		}
	}
}
