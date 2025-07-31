package store

import (
	"context"
	"fmt"

	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/model"
	"github.com/google/uuid"
)

func (store *Store) AddTestFlag(ctx context.Context, flag model.FeatureFlag) error {
	query := fmt.Sprintf(`
        INSERT INTO %s (id, key, description, enabled, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, FeatureFlagsTable)
	_, err := store.pool.Exec(
		ctx, query,
		flag.ID,
		flag.Key,
		flag.Description,
		flag.Enabled,
		flag.CreatedAt,
		flag.UpdatedAt,
	)
	return err
}

func (store *Store) RemoveTestFlag(ctx context.Context, id uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, FeatureFlagsTable)
	_, err := store.pool.Exec(ctx, query, id)
	return err
}

func (store *Store) FetchTestFlagByID(ctx context.Context, id uuid.UUID) (model.FeatureFlag, error) {
	query := fmt.Sprintf(`SELECT id, key, description, enabled, created_at, updated_at FROM %s WHERE id = $1`, FeatureFlagsTable)
	var flag model.FeatureFlag
	if err := store.pool.QueryRow(ctx, query, id).Scan(
		&flag.ID, &flag.Key, &flag.Description, &flag.Enabled, &flag.CreatedAt, &flag.UpdatedAt,
	); err != nil {
		return model.FeatureFlag{}, fmt.Errorf("failed to get test feature flag: %w", err)
	}

	return flag, nil
}
