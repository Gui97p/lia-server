package memories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("memory not found")

type Memory struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	Scope     string
	Fact      string
	Category  *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Store interface {
	Create(ctx context.Context, scope MemoryScope, fact string, userID *uuid.UUID) (*Memory, error)
	GetByID(ctx context.Context, memoryID uuid.UUID) (*Memory, error)
	ListByScope(ctx context.Context, scope MemoryScope, limit int) ([]Memory, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Memory, error)
	SetCategory(ctx context.Context, memoryID uuid.UUID, category string) error
	SetFact(ctx context.Context, memoryID uuid.UUID, fact string) error
	Delete(ctx context.Context, memoryID uuid.UUID) error
}
