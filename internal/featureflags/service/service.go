package service

import "errors"

var (
	ErrNotFound = errors.New("feature flag not found")
)
