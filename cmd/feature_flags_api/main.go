package main

import (
	"context"
	"sync"

	"github.com/georgisomnoev/feature-flag-api/internal/auth"
	"github.com/georgisomnoev/feature-flag-api/internal/config"
	"github.com/georgisomnoev/feature-flag-api/internal/webapi"
)

func main() {
	appCtx := context.Background()
	wg := &sync.WaitGroup{}

	cfg := config.Load()

	srv := webapi.NewWebAPI()

	tlsCfg := &webapi.TLSConfig{
		CertFile: cfg.WebAPICertFile,
		KeyFile:  cfg.WebAPIKeyFile,
	}

	auth.Process(appCtx, nil, srv)

	webapi.Start(appCtx, wg, srv, cfg.Port, tlsCfg)

	wg.Wait()
}
