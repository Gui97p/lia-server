package transport

import (
	"context"
	"encoding/json"
)

type Envelope struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(ctx context.Context, session *Session, payload json.RawMessage) error

type router struct {
	handlers map[string]EventHandler
}

func newRouter() *router {
	return &router{handlers: make(map[string]EventHandler)}
}

func (r *router) register(event string, h EventHandler) {
	r.handlers[event] = h
}

func (r *router) dispatch(ctx context.Context, session *Session, raw []byte) error {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return sendError(ctx, session, "invalid envelope: "+err.Error())
	}

	handler, ok := r.handlers[env.Event]
	if !ok {
		return sendError(ctx, session, "unknown event: "+env.Event)
	}

	return handler(ctx, session, env.Payload)
}

func sendError(ctx context.Context, session *Session, message string) error {
	return session.Writer(ctx, "error", map[string]string{"message": message})
}
