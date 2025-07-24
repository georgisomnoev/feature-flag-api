package main

import (
	"context"

	"github.com/georgisomnoev/feature-flag-api/internal/auth"
	"github.com/georgisomnoev/feature-flag-api/internal/config"
	"github.com/georgisomnoev/feature-flag-api/internal/webapi"
)

func main() {
	appCtx := context.Background()
	cfg := config.Load()

	srv := webapi.NewWebAPI()

	tlsCfg := &webapi.TLSConfig{
		CertFile: cfg.WebAPICertFile,
		KeyFile:  cfg.WebAPIKeyFile,
	}

	auth.Process(appCtx, nil, srv)

	srv.Logger.Fatal(webapi.StartServer(srv, cfg.Port, tlsCfg))
}
