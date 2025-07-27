package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type FeatureFlag struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type FeatureFlagRequest struct {
	Key         string `json:"key" validate:"required"`
	Description string `json:"description" validate:"required"`
	Enabled     bool   `json:"enabled"`
}

type FeatureFlagResponse struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Description string    `json:"description"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

var ErrNotFound = errors.New("feature flag not found")
