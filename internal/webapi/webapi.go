package webapi

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type TLSConfig struct {
	CertFile string
	KeyFile  string
}

func NewWebAPI() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetLevel(log.INFO)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	return e
}

func StartServer(e *echo.Echo, apiPort string, tlsConfig *TLSConfig) error {
	e.Logger.Infof("Starting HTTPS server on port: %s", apiPort)

	addr := fmt.Sprintf(":%s", apiPort)
	if err := e.StartTLS(addr, tlsConfig.CertFile, tlsConfig.KeyFile); err != nil {
		return fmt.Errorf("failed to start HTTPS server: %w", err)
	}
	return nil
}
