package lifecycle

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func CreateAppContext() context.Context {
	appCtx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer close(sigChan)
		select {
		case <-sigChan:
			cancel()
		case <-appCtx.Done():
		}
	}()

	return appCtx
}
