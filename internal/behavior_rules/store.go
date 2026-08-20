package behaviorrules

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("behavior rule not found")

type BehaviorRule struct {
	ID        uuid.UUID
	UserID    *uuid.UUID
	Rule      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Store interface {
	Create(ctx context.Context, userID *uuid.UUID, rule string) (*BehaviorRule, error)
	ListActive(ctx context.Context, userID uuid.UUID) ([]BehaviorRule, error)
	Update(ctx context.Context, ruleID uuid.UUID, rule string) error
	Delete(ctx context.Context, ruleID uuid.UUID) error
}
