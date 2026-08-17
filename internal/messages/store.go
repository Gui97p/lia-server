package messages

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	Role      string
	Content   string
	CreatedAt time.Time
}

type Store interface {
	Save(ctx context.Context, userID uuid.UUID, role, content string) (*Message, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Message, error)
}
