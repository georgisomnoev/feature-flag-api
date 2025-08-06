package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/georgisomnoev/feature-flag-api/internal/auth/model"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// TODO: Move it as an environment variable. Adjust the tests.
const ttl = 24 * time.Hour

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidUserRole    = errors.New("invalid user role")
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//go:generate gowrap gen -g -p ./ -i Store -t ../../observability/templates/otel_trace.tmpl -o ./wrapped/trace/store.go
//go:generate gowrap gen -g -p ./ -i Store -t ../../observability/templates/otel_metric.tmpl -o ./wrapped/metric/store.go
//counterfeiter:generate . Store
type Store interface {
	GetByUsername(context.Context, string) (*model.User, error)
}

//go:generate gowrap gen -g -p ./ -i JWTHelper -t ../../observability/templates/otel_trace.tmpl -o ./wrapped/trace/jwt_helper.go
//go:generate gowrap gen -g -p ./ -i JWTHelper -t ../../observability/templates/otel_metric.tmpl -o ./wrapped/metric/jwt_helper.go
//counterfeiter:generate . JWTHelper
type JWTHelper interface {
	GenerateToken(jwt.Claims) (string, error)
}

type Service struct {
	store     Store
	jwtHelper JWTHelper
}

func NewService(store Store, jwtHelper JWTHelper) *Service {
	return &Service{
		store:     store,
		jwtHelper: jwtHelper,
	}
}

func (a *Service) Authenticate(ctx context.Context, username, password string) (string, error) {
	user, err := a.store.GetByUsername(ctx, username)
	if err != nil {
		return "", fmt.Errorf("failed to get user data: %w", err)
	}
	if user == nil {
		return "", ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(ttl).Unix(),
	}

	switch user.Role {
	case model.RoleEditor:
		claims["scopes"] = []string{"read:flags", "write:flags"}
	case model.RoleViewer:
		claims["scopes"] = []string{"read:flags"}
	default:
		return "", ErrInvalidUserRole
	}

	token, err := a.jwtHelper.GenerateToken(claims)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}
