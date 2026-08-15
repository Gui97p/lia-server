package transport

import (
	"context"
	"encoding/json"

	"github.com/Gui97p/lia-server/internal/session"
)

type Envelope struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(ctx context.Context, sess *session.Session, payload json.RawMessage) error

type router struct {
	handlers map[string]EventHandler
}

func newRouter() *router {
	return &router{handlers: make(map[string]EventHandler)}
}

func (r *router) register(event string, h EventHandler) {
	r.handlers[event] = h
}

func (r *router) dispatch(ctx context.Context, sess *session.Session, raw []byte) error {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return sendError(ctx, sess, "invalid envelope: "+err.Error())
	}

	handler, ok := r.handlers[env.Event]
	if !ok {
		return sendError(ctx, sess, "unknown event: "+env.Event)
	}

	return handler(ctx, sess, env.Payload)
}

func sendError(ctx context.Context, sess *session.Session, message string) error {
	return sess.Writer(ctx, "error", map[string]string{"message": message})
}
