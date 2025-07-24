package store

import (
	"context"
	"fmt"

	"github.com/georgisomnoev/feature-flag-api/internal/auth/model"
	"github.com/google/uuid"
)

func (s *Store) AddUser(ctx context.Context, user model.User) error {
	query := fmt.Sprintf(`INSERT INTO %s (id, username, password, role) VALUES ($1, $2, $3, $4)`, UserTable)

	_, err := s.pool.Exec(ctx, query, user.ID, user.Username, user.Password, user.Role)
	if err != nil {
		return fmt.Errorf("failed to insert user into the DB: %w", err)
	}

	return nil
}

func (s *Store) DeleteUserByID(ctx context.Context, userID uuid.UUID) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, UserTable)

	_, err := s.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user from the DB: %w", err)
	}

	return nil
}
