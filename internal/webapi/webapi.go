package webapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/georgisomnoev/feature-flag-api/internal/validator"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	gracefulShutdownTimeout = 5 * time.Second
	contextTimeout          = 5 * time.Second
)

func NewWebAPI() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Logger.SetLevel(log.INFO)
	e.Use(middleware.ContextTimeout(contextTimeout))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Validator = validator.GetValidator()

	return e
}

func Start(ctx context.Context, e *echo.Echo, apiPort string) {
	go func() {
		e.Logger.Infof("starting the WebAPI server on port: %s", apiPort)
		addr := fmt.Sprintf(":%s", apiPort)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("failed to start WebAPI server: %v", err)
		}
	}()

	<-ctx.Done()
	e.Logger.Info("context canceled, shutting down WebAPI server")

	ctxGrace, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	if err := e.Shutdown(ctxGrace); err != nil {
		e.Logger.Errorf("failed to shutdown WebAPI server: %v", err)
	}
}
