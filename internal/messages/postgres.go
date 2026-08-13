package messages

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Save(ctx context.Context, userID *uuid.UUID, role, content string) (*Message, error) {
	var m Message
	err := s.pool.QueryRow(ctx,
		`INSERT INTO messages (user_id, role, content) VALUES ($1, $2, $3) 
			RETURNING id, user_id, role, content, created_at`,
		userID, role, content,
	).Scan(&m.ID, &m.UserID, &m.Role, &m.Content, &m.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *PostgresStore) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Message, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, role, content, created_at FROM messages WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ms []Message
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ms, nil
}
