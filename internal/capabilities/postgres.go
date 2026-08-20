package capabilities

import (
	"context"
	"encoding/json"

	"github.com/Gui97p/lia-server/internal/auth"
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

func (s *PostgresStore) Create(ctx context.Context, name, description string, parameters json.RawMessage, trustLevel auth.TrustLevel, source *string) (*Capability, error) {
	var c Capability
	err := s.pool.QueryRow(ctx,
		`INSERT INTO capabilities (name, description, parameters, trust_level, source) VALUES ($1, $2, $3, $4, $5) 
			RETURNING id, name, description, parameters, trust_level, source, created_at, updated_at`,
		name, description, parameters, trustLevel, source,
	).Scan(&c.ID, &c.Name, &c.Description, &c.Parameters, &c.TrustLevel, &c.Source, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *PostgresStore) GetByName(ctx context.Context, name string) (*Capability, error) {
	var c Capability
	err := s.pool.QueryRow(ctx,
		`SELECT id, name, description, parameters, trust_level, source, created_at, updated_at FROM capabilities
		WHERE name = $1
		ORDER BY created_at LIMIT 1`,
		name,
	).Scan(&c.ID, &c.Name, &c.Description, &c.Parameters, &c.TrustLevel, &c.Source, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func (s *PostgresStore) GetByNames(ctx context.Context, names []string) ([]Capability, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, description, parameters, trust_level, source, created_at, updated_at FROM capabilities
		WHERE name = ANY($1)`,
		names,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cs []Capability
	for rows.Next() {
		var c Capability
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Parameters, &c.TrustLevel, &c.Source, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cs, nil
}

func (s *PostgresStore) ListAll(ctx context.Context) ([]Capability, error) {
	rows, err := s.pool.Query(ctx,
		`SELECT id, name, description, parameters, trust_level, source, created_at, updated_at FROM capabilities`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cs []Capability
	for rows.Next() {
		var c Capability
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Parameters, &c.TrustLevel, &c.Source, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		cs = append(cs, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cs, nil
}

func (s *PostgresStore) SetDescription(ctx context.Context, capabilityID uuid.UUID, description string) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE capabilities SET description = $1 WHERE id = $2`,
		description, capabilityID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostgresStore) SetParameters(ctx context.Context, capabilityID uuid.UUID, parameters json.RawMessage) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE capabilities SET parameters = $1 WHERE id = $2`,
		parameters, capabilityID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostgresStore) SetTrustLevel(ctx context.Context, capabilityID uuid.UUID, trustLevel auth.TrustLevel) error {
	tag, err := s.pool.Exec(ctx,
		`UPDATE capabilities SET trust_level = $1 WHERE id = $2`,
		trustLevel, capabilityID,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostgresStore) Delete(ctx context.Context, capabilityID uuid.UUID) error {
	tag, err := s.pool.Exec(ctx,
		`DELETE FROM capabilities WHERE id = $1`,
		capabilityID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
