package featureflags

import (
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/handler"
	handlerWrappers "github.com/georgisomnoev/feature-flag-api/internal/featureflags/handler/wrapped/trace"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/service"
	serviceWrappers "github.com/georgisomnoev/feature-flag-api/internal/featureflags/service/wrapped/trace"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func Process(
	pool *pgxpool.Pool,
	srv *echo.Echo,
	authStore handler.AuthStore,
	jwtHelper handler.JWTHelper,
) {
	featureFlagStore := store.NewStore(pool)
	wrappedFFStore := serviceWrappers.NewStoreWithTracing(featureFlagStore)
	featureFlagService := service.NewService(wrappedFFStore)
	wrappedFFService := handlerWrappers.NewServiceWithTracing(featureFlagService)
	wrappedAuthStore := handlerWrappers.NewAuthStoreWithTracing(authStore)
	wrappedJWTHelper := handlerWrappers.NewJWTHelperWithTracing(jwtHelper)
	featureFlagHandler := handler.NewHandler(wrappedFFService, wrappedAuthStore, wrappedJWTHelper)
	featureFlagHandler.RegisterHandlers(srv)
}
