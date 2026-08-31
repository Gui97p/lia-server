package tasks

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/Gui97p/lia-server/internal/auth"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

var _ Store = (*PostgresStore)(nil)

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(ctx context.Context, userID uuid.UUID, triggerType TriggerType, trustLevel auth.TrustLevel, conversationID uuid.UUID) (*Task, error) {
	var t Task
	err := s.pool.QueryRow(ctx,
		`INSERT INTO tasks (user_id, state, trigger_type, authorized_trust_level, conversation_id) VALUES ($1, $2, $3, $4, $5)
			RETURNING id, user_id, state, trigger_type, authorized_trust_level, conversation_id, created_at, updated_at`,
		userID, Created, triggerType, trustLevel, conversationID,
	).Scan(&t.ID, &t.UserID, &t.State, &t.TriggerType, &t.AuthorizedTrustLevel, &t.ConversationID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	t.Workflow = json.RawMessage{}
	return &t, nil
}

func (s *PostgresStore) LatestUserTask(ctx context.Context, userID uuid.UUID) (*Task, error) {
	var t Task
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, state, workflow, trigger_type, authorized_trust_level, conversation_id, created_at, updated_at
		FROM tasks WHERE user_id = $1 AND trigger_type = $2 ORDER BY created_at DESC LIMIT 1`,
		userID, User,
	).Scan(&t.ID, &t.UserID, &t.State, &t.Workflow, &t.TriggerType, &t.AuthorizedTrustLevel, &t.ConversationID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (s *PostgresStore) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Task, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, state, workflow, trigger_type, authorized_trust_level, conversation_id, created_at, updated_at FROM tasks WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ts []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.UserID, &t.State, &t.Workflow, &t.TriggerType, &t.AuthorizedTrustLevel, &t.ConversationID, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ts, nil
}

func (s *PostgresStore) GetByID(ctx context.Context, taskID uuid.UUID) (*Task, error) {
	var t Task
	err := s.pool.QueryRow(ctx,
		"SELECT id, user_id, state, workflow, trigger_type, authorized_trust_level, conversation_id, created_at, updated_at FROM tasks WHERE id = $1",
		taskID,
	).Scan(&t.ID, &t.UserID, &t.State, &t.Workflow, &t.TriggerType, &t.AuthorizedTrustLevel, &t.ConversationID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *PostgresStore) SetState(ctx context.Context, taskID uuid.UUID, state TaskState) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE tasks SET state = $1 WHERE id = $2`,
		state, taskID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostgresStore) SetWorkflow(ctx context.Context, taskID uuid.UUID, workflow json.RawMessage) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE tasks SET workflow = $1 WHERE id = $2`,
		workflow, taskID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostgresStore) RecoverStaleTasks(ctx context.Context) (int64, error) {
	tag, err := s.pool.Exec(ctx,
		`UPDATE tasks SET state = 'failed' WHERE state NOT IN ('completed', 'failed', 'cancelled')`,
	)

	return tag.RowsAffected(), err
}
