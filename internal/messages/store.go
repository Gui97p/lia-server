package messages

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("message not found")

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
	ListByConversation(ctx context.Context, userID uuid.UUID, conversationID uuid.UUID, limit int) ([]Message, error)
	GetFirstByTask(ctx context.Context, taskID uuid.UUID) (*Message, error)
	Delete(ctx context.Context, messageID uuid.UUID) error
}
