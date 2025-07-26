package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/model"
	"github.com/google/uuid"
)

type Service struct {
	store Store
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//go:generate gowrap gen -g -p ./ -i Store -t ../../observability/templates/otel_trace.tmpl -o ./wrapped/trace/store.go
//counterfeiter:generate . Store
type Store interface {
	ListFlags(ctx context.Context) ([]model.FeatureFlag, error)
	GetFlagByID(ctx context.Context, id uuid.UUID) (model.FeatureFlag, error)

	CreateFlag(ctx context.Context, flag model.FeatureFlag) error
	UpdateFlag(ctx context.Context, flag model.FeatureFlag) error
	DeleteFlag(ctx context.Context, id uuid.UUID) error
}

func NewService(store Store) *Service {
	return &Service{store: store}
}

func (s *Service) ListFlags(ctx context.Context) ([]model.FeatureFlag, error) {
	flags, err := s.store.ListFlags(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list flags: %w", err)
	}
	return flags, nil
}

func (s *Service) GetFlagByID(ctx context.Context, id uuid.UUID) (model.FeatureFlag, error) {
	flag, err := s.store.GetFlagByID(ctx, id)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.FeatureFlag{}, model.ErrNotFound
		}
		return model.FeatureFlag{}, fmt.Errorf("failed to fetch flag: %w", err)
	}
	return flag, nil
}

func (s *Service) CreateFlag(ctx context.Context, flag model.FeatureFlag) error {
	if err := s.store.CreateFlag(ctx, flag); err != nil {
		return fmt.Errorf("failed to create flag: %w", err)
	}
	return nil
}

func (s *Service) UpdateFlag(ctx context.Context, flag model.FeatureFlag) error {
	if err := s.store.UpdateFlag(ctx, flag); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.ErrNotFound
		}
		return fmt.Errorf("failed to update flag: %w", err)
	}
	return nil
}

func (s *Service) DeleteFlag(ctx context.Context, id uuid.UUID) error {
	if err := s.store.DeleteFlag(ctx, id); err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return model.ErrNotFound
		}
		return fmt.Errorf("failed to delete flag: %w", err)
	}
	return nil
}
