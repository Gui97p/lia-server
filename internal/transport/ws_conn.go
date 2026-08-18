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
	"github.com/Gui97p/lia-server/internal/session"
	"github.com/coder/websocket"
	"github.com/google/uuid"
)

func (s *Server) handshake(ctx context.Context, conn *websocket.Conn) (*session.Session, error) {
	sess := session.Session{}
	var writeMu sync.Mutex
	sess.Writer = func(ctx context.Context, event string, payload any) error {
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
		sendError(ctx, &sess, "invalid request")
		return nil, err
	}

	if env.Event != "auth" {
		sendError(ctx, &sess, "first message must be auth")
		return nil, fmt.Errorf("expected auth event")
	}

	var authPayload struct {
		Token        string   `json:"token"`
		Capabilities []string `json:"capabilities"`
	}
	if err := json.Unmarshal(env.Payload, &authPayload); err != nil {
		sendError(ctx, &sess, "invalid payload")
		return nil, err
	}

	claims, err := auth.ParseToken(s.jwtSecret, authPayload.Token)
	if err != nil {
		sendError(ctx, &sess, "invalid token")
		return nil, err
	}

	if authPayload.Capabilities == nil {
		sendError(ctx, &sess, "invalid capabilities")
		return nil, errors.New("invalid capabilities")
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		sendError(ctx, &sess, "invalid uuid")
		return nil, err
	}

	user, err := s.usersStore.GetByID(ctx, userID)
	if err != nil {
		sendError(ctx, &sess, "user not found")
		return nil, err
	}

	if user.TokenVersion != claims.TokenVersion {
		sendError(ctx, &sess, "token revoked")
		return nil, errors.New("token version mismatch")
	}

	if user.GroqAPIKeyEncrypted == nil {
		sendError(ctx, &sess, "invalid api key")
		return nil, errors.New("api key not set")
	}

	groqAPIKey, err := crypto.Decrypt(*user.GroqAPIKeyEncrypted, s.encryptionKey)
	if err != nil {
		sendError(ctx, &sess, "invalid api key")
		return nil, err
	}

	sess.ConnID = uuid.New()
	sess.UserID = userID
	sess.Username = user.Username
	sess.TrustLevel = claims.TrustLevel
	sess.GroqAPIKey = groqAPIKey
	sess.Capabilities = authPayload.Capabilities

	return &sess, nil
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		s.logger.Error("websocket accept failed", "error", err)
		return
	}

	ctx := context.Background()
	timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)

	sess, err := s.handshake(timeoutCtx, conn)
	cancel()
	if err != nil {
		conn.CloseNow()
		return
	}
	defer conn.CloseNow()

	s.hub.Register(sess.ConnID, sess)
	s.planningQueue.EnsureStarted(sess.UserID)
	defer s.planningQueue.StopIfUnused(sess.UserID, s.hub)
	defer s.hub.Unregister(sess.ConnID)

	type AuthResponse struct {
		ConnID string `json:"conn_id"`
	}
	err = sess.Writer(ctx, "auth.ok", AuthResponse{ConnID: sess.ConnID.String()})
	if err != nil {
		return
	}

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			break
		}
		go func(data []byte) {
			if err := s.router.dispatch(ctx, sess, data); err != nil {
				s.logger.Error("dispatch failed", "error", err)
			}
		}(data)
	}
}
