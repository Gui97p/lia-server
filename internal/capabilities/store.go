package capabilities

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Gui97p/lia-server/internal/auth"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("capability not found")

type Capability struct {
	ID          uuid.UUID
	Name        string
	Description string
	Parameters  json.RawMessage
	TrustLevel  auth.TrustLevel
	Source      *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Store interface {
	Create(ctx context.Context, name, description string, parameters json.RawMessage, trustLevel auth.TrustLevel, source *string) (*Capability, error)
	GetByName(ctx context.Context, name string) (*Capability, error)
	GetByNames(ctx context.Context, names []string) ([]Capability, error)
	ListAll(ctx context.Context) ([]Capability, error)
	SetDescription(ctx context.Context, capabilityID uuid.UUID, description string) error
	SetParameters(ctx context.Context, capabilityID uuid.UUID, parameters json.RawMessage) error
	SetTrustLevel(ctx context.Context, capabilityID uuid.UUID, trustLevel auth.TrustLevel) error
	Delete(ctx context.Context, capabilityID uuid.UUID) error
}
