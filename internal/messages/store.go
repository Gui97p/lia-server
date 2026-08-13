package messages

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("message not found")
var ErrUserNotFound = errors.New("user not found")

type Message struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	Role      string
	Content   string
	CreatedAt time.Time
}

type Store interface {
	Save(ctx context.Context, userID *uuid.UUID, role, content string) (*Message, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Message, error)
}
