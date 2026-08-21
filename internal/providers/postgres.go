package providers

import (
	"context"
	"fmt"

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

func (s *PostgresStore) FindByUser(ctx context.Context, userID uuid.UUID) (Providers, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT provider, encrypted_key FROM providers
		WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ps := make(Providers)
	for rows.Next() {
		var p Provider
		if err := rows.Scan(&p.Provider, &p.EncryptedKey); err != nil {
			return nil, err
		}
		ps[p.Provider] = p.EncryptedKey
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ps, nil
}

func (s *PostgresStore) SetKey(ctx context.Context, userID uuid.UUID, keyProvider ProviderName, encryptedKey string) error {
	if !keyProvider.Valid() {
		return fmt.Errorf("unknown provider: %s", keyProvider)
	}

	_, err := s.pool.Exec(ctx,
		`INSERT INTO providers (user_id, provider, encrypted_key) VALUES ($1, $2, $3) ON CONFLICT (user_id, provider) DO UPDATE SET encrypted_key = $3`,
		userID, string(keyProvider), encryptedKey,
	)
	return err
}

func (s *PostgresStore) ResetKey(ctx context.Context, userID uuid.UUID, keyProvider ProviderName) error {
	if !keyProvider.Valid() {
		return fmt.Errorf("unknown provider: %s", keyProvider)
	}

	_, err := s.pool.Exec(ctx,
		`DELETE FROM providers WHERE user_id = $1 AND provider = $2`,
		userID, string(keyProvider),
	)
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
