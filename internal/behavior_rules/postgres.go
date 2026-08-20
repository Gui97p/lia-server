package behaviorrules

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

func (s *PostgresStore) Create(ctx context.Context, userID *uuid.UUID, rule string) (*BehaviorRule, error) {
	var br BehaviorRule
	err := s.pool.QueryRow(ctx,
		`INSERT INTO behavior_rules (user_id, rule) VALUES ($1, $2) 
			RETURNING id, user_id, rule, created_at, updated_at`,
		userID, rule,
	).Scan(&br.ID, &br.UserID, &br.Rule, &br.CreatedAt, &br.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &br, nil
}

func (s *PostgresStore) ListActive(ctx context.Context, userID uuid.UUID) ([]BehaviorRule, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, user_id, rule, created_at, updated_at FROM behavior_rules WHERE user_id = $1 OR user_id IS NULL ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brs []BehaviorRule
	for rows.Next() {
		var br BehaviorRule
		if err := rows.Scan(&br.ID, &br.UserID, &br.Rule, &br.CreatedAt, &br.UpdatedAt); err != nil {
			return nil, err
		}
		brs = append(brs, br)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return brs, nil
}

func (s *PostgresStore) Update(ctx context.Context, ruleID uuid.UUID, rule string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE behavior_rules SET rule = $1 WHERE id = $2`,
		rule, ruleID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostgresStore) Delete(ctx context.Context, ruleID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM behavior_rules WHERE id = $1`,
		ruleID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
