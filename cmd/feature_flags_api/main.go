package main

import (
	"fmt"

	"github.com/georgisomnoev/feature-flag-api/internal/auth"
	"github.com/georgisomnoev/feature-flag-api/internal/config"
	"github.com/georgisomnoev/feature-flag-api/internal/featureflags"
	"github.com/georgisomnoev/feature-flag-api/internal/jwthelper"
	"github.com/georgisomnoev/feature-flag-api/internal/lifecycle"
	"github.com/georgisomnoev/feature-flag-api/internal/observability"
	"github.com/georgisomnoev/feature-flag-api/internal/pg"
	"github.com/georgisomnoev/feature-flag-api/internal/webapi"
)

func main() {
	appCtx := lifecycle.CreateAppContext()

	cfg := config.Load()

	if cfg.OtelCollectorHost != "" {
		if err := observability.InitOtel(appCtx, cfg.OtelCollectorHost, "feature-flags-api"); err != nil {
			panic(fmt.Errorf("failed initializing Otel: %w", err))
		}
	}

	srv := webapi.NewWebAPI()

	dbCfg := pg.PoolConfig{
		MinConns:          cfg.DBMinConns,
		MaxConns:          cfg.DBMaxConns,
		MaxConnLifetime:   cfg.DBMaxConnLifetime,
		MaxConnIdleTime:   cfg.DBMaxConnIdleTime,
		HealthCheckPeriod: cfg.DBHealthCheck,
	}
	pool, err := pg.InitPool(appCtx, cfg.DBConnectionURL, dbCfg)
	if err != nil {
		panic(fmt.Errorf("failed initializing DB pool: %w", err))
	}
	defer pool.Close()

	jwtHelper, err := jwthelper.NewJWTHelper(
		cfg.JWTPrivateKeyPath,
		cfg.JWTPublicKeyPath,
	)
	if err != nil {
		panic(fmt.Errorf("failed initializing JWT helper: %w", err))
	}

	authStore := auth.Process(pool, srv, jwtHelper)
	featureflags.Process(pool, srv, authStore, jwtHelper)

	webapi.Start(appCtx, srv, cfg.APIPort)
}
