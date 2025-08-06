package auth

import (
	"github.com/georgisomnoev/feature-flag-api/internal/auth/handler"
	handlerWrappers "github.com/georgisomnoev/feature-flag-api/internal/auth/handler/wrapped/trace"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/service"
	metricWrappers "github.com/georgisomnoev/feature-flag-api/internal/auth/service/wrapped/metric"
	traceWrappers "github.com/georgisomnoev/feature-flag-api/internal/auth/service/wrapped/trace"
	"github.com/georgisomnoev/feature-flag-api/internal/auth/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

func Process(
	pool *pgxpool.Pool,
	srv *echo.Echo,
	jwtHelper service.JWTHelper,
) *store.Store {
	authStore := store.NewStore(pool)
	metricWrappedAuthStore := metricWrappers.NewStoreWithMetrics(authStore)
	wrappedAuthStore := traceWrappers.NewStoreWithTracing(metricWrappedAuthStore)
	metricWrappedJWTHelper := metricWrappers.NewJWTHelperWithMetrics(jwtHelper)
	wrappedJWTHelper := traceWrappers.NewJWTHelperWithTracing(metricWrappedJWTHelper)
	authService := service.NewService(wrappedAuthStore, wrappedJWTHelper)
	wrappedAuthService := handlerWrappers.NewServiceWithTracing(authService)
	authHandler := handler.NewHandler(wrappedAuthService)
	authHandler.RegisterHandlers(srv)

	return authStore
}
