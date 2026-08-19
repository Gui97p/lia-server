package messages

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

var _ Store = (*PostgresStore)(nil)

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(ctx context.Context, userID uuid.UUID, role, content string, taskID uuid.UUID) (*Message, error) {
	var m Message
	err := s.pool.QueryRow(ctx,
		`INSERT INTO messages (user_id, role, content, task_id) VALUES ($1, $2, $3, $4) 
			RETURNING id, user_id, role, content, task_id, created_at`,
		userID, role, content, taskID,
	).Scan(&m.ID, &m.UserID, &m.Role, &m.Content, &m.TaskID, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *PostgresStore) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Message, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, role, content, task_id, created_at FROM messages WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ms []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.Role, &m.Content, &m.TaskID, &m.CreatedAt); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ms, nil
}

func (s *PostgresStore) ListByTask(ctx context.Context, userID uuid.UUID, taskID uuid.UUID, limit int) ([]Message, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, role, content, task_id, created_at FROM messages 
		WHERE user_id = $2 AND (task_id = $1 OR 
		task_id IN (SELECT id FROM tasks WHERE user_id = $2 AND state IN ('completed', 'failed', 'cancelled')))
		ORDER BY created_at DESC LIMIT $3`,
		taskID, userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ms []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.Role, &m.Content, &m.TaskID, &m.CreatedAt); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ms, nil
}
