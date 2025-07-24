package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/georgisomnoev/feature-flag-api/internal/auth"
	"github.com/georgisomnoev/feature-flag-api/internal/config"
	"github.com/georgisomnoev/feature-flag-api/internal/jwthelper"
	"github.com/georgisomnoev/feature-flag-api/internal/pg"
	"github.com/georgisomnoev/feature-flag-api/internal/webapi"
)

func main() {
	appCtx := context.Background()
	wg := &sync.WaitGroup{}

	cfg := config.Load()

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

	auth.Process(appCtx, pool, srv, jwtHelper)

	tlsCfg := &webapi.TLSConfig{
		CertFile: cfg.WebAPICertPath,
		KeyFile:  cfg.WebAPIKeyPath,
	}
	webapi.Start(appCtx, wg, srv, cfg.APIPort, tlsCfg)

	wg.Wait()
}
