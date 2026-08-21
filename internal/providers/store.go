package providers

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrNotFound = errors.New("provider not found")

type Provider struct {
	UserID                uuid.UUID
	GroqApiKeyEncrypted   *string
	GeminiApiKeyEncrypted *string
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type Store interface {
	FindByUser(ctx context.Context, userID uuid.UUID) (*Provider, error)
	SetKey(ctx context.Context, userID uuid.UUID, keyProvider ProviderName, encryptedkey string) error
	ResetKey(ctx context.Context, userID uuid.UUID, key ProviderName) error
	Delete(ctx context.Context, userID uuid.UUID) error
}
