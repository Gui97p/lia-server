package memories

import (
	"context"
	"errors"

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

func (s *PostgresStore) Create(ctx context.Context, scope MemoryScope, fact string, userID *uuid.UUID) (*Memory, error) {
	if scope == User && userID == nil {
		return nil, errors.New("user memory cannot be created without an user")
	}
	var m Memory
	err := s.pool.QueryRow(ctx,
		`INSERT INTO memories (user_id, scope, fact) VALUES ($1, $2, $3) 
			RETURNING id, user_id, scope, fact, category, created_at, updated_at`,
		userID, scope, fact,
	).Scan(&m.ID, &m.UserID, &m.Scope, &m.Fact, &m.Category, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *PostgresStore) GetByID(ctx context.Context, memoryID uuid.UUID) (*Memory, error) {
	var m Memory
	err := s.pool.QueryRow(ctx,
		`SELECT id, user_id, scope, fact, category, created_at, updated_at FROM memories WHERE id = $1`,
		memoryID,
	).Scan(&m.ID, &m.UserID, &m.Scope, &m.Fact, &m.Category, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (s *PostgresStore) ListByScope(ctx context.Context, scope MemoryScope, limit int) ([]Memory, error) {
	if scope == User {
		return nil, errors.New("use ListByUser for user-scoped memories")
	}

	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, scope, fact, category, created_at, updated_at FROM memories WHERE scope = $1 ORDER BY created_at DESC LIMIT $2`,
		scope, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ms []Memory
	for rows.Next() {
		var m Memory
		if err := rows.Scan(&m.ID, &m.UserID, &m.Scope, &m.Fact, &m.Category, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ms, nil
}

func (s *PostgresStore) ListByUser(ctx context.Context, userID uuid.UUID, limit int) ([]Memory, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, scope, fact, category, created_at, updated_at FROM memories WHERE user_id = $1 AND scope = 'user' ORDER BY created_at DESC LIMIT $2`,
		userID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ms []Memory
	for rows.Next() {
		var m Memory
		if err := rows.Scan(&m.ID, &m.UserID, &m.Scope, &m.Fact, &m.Category, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		ms = append(ms, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ms, nil
}

func (s *PostgresStore) SetCategory(ctx context.Context, memoryID uuid.UUID, category string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE memories SET category = $1 WHERE id = $2`,
		category, memoryID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostgresStore) SetFact(ctx context.Context, memoryID uuid.UUID, fact string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE memories SET fact = $1 WHERE id = $2`,
		fact, memoryID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostgresStore) Delete(ctx context.Context, memoryID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM memories WHERE id = $1`,
		memoryID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
