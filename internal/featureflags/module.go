package featureflags

import (
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/handler"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func Process(
	pool *pgxpool.Pool,
	srv *echo.Echo,
	authStore handler.AuthStore,
	jwtHelper handler.JWTHelper,
) {
	//featureFlagStore := store.NewStore(pool)
	//featureFlagService := service.NewService(featureFlagStore, authStore, jwtHelper)
	featureFlagHandler := handler.NewHandler(nil, authStore, jwtHelper)
	featureFlagHandler.RegisterHandlers(srv)
}
