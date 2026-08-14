package transport

import (
	"context"
	"encoding/json"

	"github.com/coder/websocket"
)

type Envelope struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(ctx context.Context, conn *websocket.Conn, payload json.RawMessage) error

type router struct {
	handlers map[string]EventHandler
}

func newRouter() *router {
	return &router{handlers: make(map[string]EventHandler)}
}

func (r *router) register(event string, h EventHandler) {
	r.handlers[event] = h
}

func (r *router) dispatch(ctx context.Context, conn *websocket.Conn, raw []byte) error {
	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return sendError(ctx, conn, "invalid envelope: "+err.Error())
	}

	handler, ok := r.handlers[env.Event]
	if !ok {
		return sendError(ctx, conn, "unknown event: "+env.Event)
	}

	return handler(ctx, conn, env.Payload)
}

func sendError(ctx context.Context, conn *websocket.Conn, message string) error {
	payload, err := json.Marshal(map[string]string{"message": message})
	if err != nil {
		return err
	}

	data, err := json.Marshal(Envelope{Event: "error", Payload: payload})
	if err != nil {
		return err
	}

	return conn.Write(ctx, websocket.MessageText, data)
}
