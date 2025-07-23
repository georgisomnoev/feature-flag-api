package main

import (
	"github.com/georgisomnoev/feature-flag-api/internal/config"
	"github.com/georgisomnoev/feature-flag-api/internal/webapi"
)

func main() {
	cfg := config.Load()

	srv := webapi.NewWebAPI()

	tlsCfg := &webapi.TLSConfig{
		CertFile: cfg.WebAPICertFile,
		KeyFile:  cfg.WebAPIKeyFile,
	}

	srv.Logger.Fatal(webapi.StartServer(srv, cfg.Port, tlsCfg))
}
