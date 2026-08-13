package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) Create(ctx context.Context, username string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (username)
               VALUES ($1)
               RETURNING id, username, groq_api_key_encrypted, token_version, created_at, updated_at`,
		username,
	).Scan(&u.ID, &u.Username, &u.GroqAPIKeyEncrypted, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *PostgresStore) GetByUsername(ctx context.Context, username string) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, groq_api_key_encrypted, token_version, created_at, updated_at FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &u.GroqAPIKeyEncrypted, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *PostgresStore) GetByID(ctx context.Context, userId uuid.UUID) (*User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, groq_api_key_encrypted, token_version, created_at, updated_at FROM users WHERE id = $1`,
		userId,
	).Scan(&u.ID, &u.Username, &u.GroqAPIKeyEncrypted, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &u, nil
}

func (s *PostgresStore) SetGroqAPIKey(ctx context.Context, userID uuid.UUID, encryptedKey string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE users SET groq_api_key_encrypted = $1 WHERE id = $2`,
		encryptedKey, userID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostgresStore) BumpTokenVersion(ctx context.Context, userID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE users SET token_version = token_version + 1 WHERE id = $1`,
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

func (s *PostgresStore) Delete(ctx context.Context, userID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM users WHERE id = $1`,
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
