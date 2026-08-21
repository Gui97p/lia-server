package session

import (
	"context"
	"sync"

	"github.com/Gui97p/lia-server/internal/auth"
	"github.com/Gui97p/lia-server/internal/providers"
	"github.com/google/uuid"
)

type Writer func(ctx context.Context, event string, payload any) error

type Session struct {
	ConnID       uuid.UUID
	UserID       uuid.UUID
	Username     string
	TrustLevel   auth.TrustLevel
	Providers    providers.Providers
	Capabilities []string
	Writer       Writer

	pendingMu sync.Mutex
	pending   map[string]chan ToolResult
}

type Hub struct {
	mtx      sync.Mutex
	sessions map[uuid.UUID]*Session
}

func NewHub() *Hub {
	return &Hub{
		mtx:      sync.Mutex{},
		sessions: make(map[uuid.UUID]*Session),
	}
}

func (h *Hub) Register(connID uuid.UUID, session *Session) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	h.sessions[connID] = session
}

func (h *Hub) Unregister(connID uuid.UUID) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	delete(h.sessions, connID)
}

func (h *Hub) FindByUser(userID uuid.UUID) []*Session {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	sessions := make([]*Session, 0)

	for _, s := range h.sessions {
		if s.UserID == userID {
			sessions = append(sessions, s)
		}
	}

	return sessions
}

func (h *Hub) FindByID(connID uuid.UUID) (*Session, bool) {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	session, ok := h.sessions[connID]

	return session, ok
}
