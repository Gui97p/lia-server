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
	TaskID    *uuid.UUID
	CreatedAt time.Time
}

type Store interface {
	Create(ctx context.Context, userID uuid.UUID, role, content string, taskID uuid.UUID) (*Message, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Message, error)
	ListByTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, limit int) ([]Message, error)
}
