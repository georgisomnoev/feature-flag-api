package webapi

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

const (
	gracefulShutdownTimeout = 5
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

func Start(ctx context.Context, wg *sync.WaitGroup, e *echo.Echo, apiPort string, tlsConfig *TLSConfig) {
	startServer(e, wg, apiPort, tlsConfig)
	stopServer(e, wg, ctx)
}

func startServer(e *echo.Echo, wg *sync.WaitGroup, apiPort string, tlsConfig *TLSConfig) {
	wg.Add(1)
	go func() {
		e.Logger.Infof("starting the WebAPI server on port: %s", apiPort)
		defer wg.Done()

		addr := fmt.Sprintf(":%s", apiPort)
		if err := e.StartTLS(addr, tlsConfig.CertFile, tlsConfig.KeyFile); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatalf("failed to start WebAPI server: %v", err)
		}
	}()
}

func stopServer(e *echo.Echo, wg *sync.WaitGroup, ctx context.Context) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		ctxGrace, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
		defer cancel()

		select {
		case <-sigChan:
			e.Logger.Info("received shutdown signal, stopping WebAPI server")
			if err := e.Shutdown(ctxGrace); err != nil {
				e.Logger.Errorf("failed to shutdown WebAPI server: %v", err)
			}
		case <-ctx.Done():
			e.Logger.Info("context cancelled, stopping WebAPI server")
			if err := e.Shutdown(ctxGrace); err != nil {
				e.Logger.Errorf("failed to shutdown WebAPI server: %v", err)
			}
		}
	}()
}
