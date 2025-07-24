package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgisomnoev/feature-flag-api/internal/auth/model"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const UserTable = "users"

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}
func (s *Store) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	query := fmt.Sprintf("SELECT id, username, password, roles FROM %s WHERE username = $1", UserTable)
	err := s.pool.QueryRow(ctx, query, username).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *Store) UserExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE id = $1)", UserTable)
	err := s.pool.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
