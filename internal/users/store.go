package users

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("user not found")

type User struct {
	ID           uuid.UUID
	Username     string
	TokenVersion int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Store interface {
	Create(ctx context.Context, username string) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByID(ctx context.Context, userID uuid.UUID) (*User, error)
	BumpTokenVersion(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, userID uuid.UUID) error
}
