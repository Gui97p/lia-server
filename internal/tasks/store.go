package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Gui97p/lia-server/internal/auth"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("task not found")

type Task struct {
	ID                   uuid.UUID
	UserID               *uuid.UUID
	State                TaskState
	Workflow             json.RawMessage
	TriggerType          TriggerType
	AuthorizedTrustLevel auth.TrustLevel
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type Store interface {
	Create(ctx context.Context, userID uuid.UUID, triggerType TriggerType, trustLevel auth.TrustLevel) (*Task, error)
	GetByID(ctx context.Context, taskID uuid.UUID) (*Task, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Task, error)
	SetState(ctx context.Context, taskID uuid.UUID, state TaskState) error
	SetWorkflow(ctx context.Context, taskID uuid.UUID, workflow json.RawMessage) error
	RecoverStaleTasks(ctx context.Context) (int64, error)
}
