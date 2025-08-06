package featureflags

import (
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/handler"
	metricHandlerWrappers "github.com/georgisomnoev/feature-flag-api/internal/featureflags/handler/wrapped/metric"
	traceHandlerWrappers "github.com/georgisomnoev/feature-flag-api/internal/featureflags/handler/wrapped/trace"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags/service"
	metricServiceWrappers "github.com/georgisomnoev/feature-flag-api/internal/featureflags/service/wrapped/metric"
	traceServiceWrappers "github.com/georgisomnoev/feature-flag-api/internal/featureflags/service/wrapped/trace"
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
	metricWrappedFFStore := metricServiceWrappers.NewStoreWithMetrics(featureFlagStore)
	wrappedFFStore := traceServiceWrappers.NewStoreWithTracing(metricWrappedFFStore)
	featureFlagService := service.NewService(wrappedFFStore)
	wrappedFFService := traceHandlerWrappers.NewServiceWithTracing(featureFlagService)
	metricWrappedAuthStore := metricHandlerWrappers.NewAuthStoreWithMetrics(authStore)
	wrappedAuthStore := traceHandlerWrappers.NewAuthStoreWithTracing(metricWrappedAuthStore)
	metricWrappedJWTHelper := metricHandlerWrappers.NewJWTHelperWithMetrics(jwtHelper)
	wrappedJWTHelper := traceHandlerWrappers.NewJWTHelperWithTracing(metricWrappedJWTHelper)
	featureFlagHandler := handler.NewHandler(wrappedFFService, wrappedAuthStore, wrappedJWTHelper)
	featureFlagHandler.RegisterHandlers(srv)
}
