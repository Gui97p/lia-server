package providers

import (
	"context"
	"errors"
	"fmt"

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

func (s *PostgresStore) FindByUser(ctx context.Context, userID uuid.UUID) (*Provider, error) {
	var p Provider
	err := s.pool.QueryRow(ctx,
		`SELECT user_id, groq_api_key_encrypted, gemini_api_key_encrypted, created_at, updated_at FROM providers WHERE user_id = $1`,
		userID,
	).Scan(&p.UserID, &p.GroqApiKeyEncrypted, &p.GeminiApiKeyEncrypted, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (s *PostgresStore) SetKey(ctx context.Context, userID uuid.UUID, keyProvider ProviderName, encryptedKey string) error {
	if !keyProvider.Valid() {
		return fmt.Errorf("unknown provider: %s", keyProvider)
	}

	column := string(keyProvider) + "_api_key_encrypted"
	query := fmt.Sprintf("INSERT INTO providers (user_id, %s) VALUES ($1, $2) ON CONFLICT (user_id) DO UPDATE SET %s = $2", column, column)

	_, err := s.pool.Exec(ctx, query, userID, encryptedKey)
	return err
}

func (s *PostgresStore) ResetKey(ctx context.Context, userID uuid.UUID, keyProvider ProviderName) error {
	if !keyProvider.Valid() {
		return fmt.Errorf("unknown provider: %s", keyProvider)
	}

	column := string(keyProvider) + "_api_key_encrypted"
	query := fmt.Sprintf("UPDATE providers SET %s = NULL WHERE user_id = $1", column)

	_, err := s.pool.Exec(ctx, query, userID)
	return err
}

func (s *PostgresStore) Delete(ctx context.Context, userID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM providers WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
