package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const FeatureFlagsTable = "feature_flags"

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) ListFlags(ctx context.Context) ([]model.FeatureFlag, error) {
	query := fmt.Sprintf(`SELECT id, key, description, enabled, created_at, updated_at FROM %s`, FeatureFlagsTable)
	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []model.FeatureFlag
	for rows.Next() {
		var flag model.FeatureFlag
		if err := rows.Scan(&flag.ID, &flag.Key, &flag.Description, &flag.Enabled, &flag.CreatedAt, &flag.UpdatedAt); err != nil {
			return nil, err
		}
		flags = append(flags, flag)
	}

	return flags, nil
}

func (s *Store) GetFlagByID(ctx context.Context, id uuid.UUID) (model.FeatureFlag, error) {
	var flag model.FeatureFlag
	query := fmt.Sprintf(`SELECT id, key, description, enabled, created_at, updated_at FROM %s WHERE id = $1`, FeatureFlagsTable)
	err := s.pool.QueryRow(ctx, query, id).Scan(&flag.ID, &flag.Key, &flag.Description, &flag.Enabled, &flag.CreatedAt, &flag.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.FeatureFlag{}, model.ErrNotFound
		}
		return model.FeatureFlag{}, err
	}

	return flag, nil
}

func (s *Store) CreateFlag(ctx context.Context, flag model.FeatureFlag) error {
	query := fmt.Sprintf(`INSERT INTO %s (id, key, description, enabled) 
		VALUES ($1, $2, $3, $4)`, FeatureFlagsTable)
	_, err := s.pool.Exec(ctx, query, flag.ID, flag.Key, flag.Description, flag.Enabled)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) UpdateFlag(ctx context.Context, flag model.FeatureFlag) error {
	query := fmt.Sprintf(`UPDATE %s SET key = $1, description = $2, enabled = $3, updated_at = NOW() WHERE id = $4`, FeatureFlagsTable)
	result, err := s.pool.Exec(ctx, query, flag.Key, flag.Description, flag.Enabled, flag.ID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (s *Store) DeleteFlag(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, FeatureFlagsTable)
	result, err := s.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return model.ErrNotFound
	}
	return nil
}
