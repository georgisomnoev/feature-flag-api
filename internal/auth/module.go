package auth

import (
	"context"

	"github.com/georgisomnoev/feature-flag-api/internal/auth/handler"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/service"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func Process(
	ctx context.Context,
	pool *pgxpool.Pool,
	srv *echo.Echo,
	jwtHelper service.JWTHelper,
) *store.Store {
	authStore := store.NewStore(pool)
	authService := service.NewService(authStore, jwtHelper)
	handler.RegisterHandlers(ctx, srv, authService)

	return authStore
}
