package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bool64/ctxd"
)

// GracefulShutdown function close all opened connection using shutdownCh channel signal.
//
// Here is the place where you should close connection such as database, or any other instance defined in the locator
// that require to be closed before the application finishes.
func GracefulShutdown(ctx context.Context, l *Locator, errCh chan error) (map[string]chan struct{}, chan struct{}, chan struct{}) {
	toShutdown := make(map[string]chan struct{})
	shutdownCh := make(chan struct{})
	shutdownDoneCh := make(chan struct{})

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGTERM, os.Interrupt)

	go func() {
		<-sigint

		close(shutdownCh)

		deadline := time.After(10 * time.Second)

		for srv, done := range toShutdown {
			select {
			case <-done:
				continue
			case <-deadline:
				errCh <- ctxd.NewError(ctx, fmt.Sprintf("shutdown deadline exceeded while waiting for service %s to shutdown", srv))
			}
		}

		gracefulDBShutdown(ctx, l)

		close(shutdownDoneCh)
	}()

	return toShutdown, shutdownCh, shutdownDoneCh
}

// gracefulDBShutdown close all db opened connection.
//
func gracefulDBShutdown(ctx context.Context, l *Locator) {
	if err := l.DBx.Close(); err != nil {
		l.LoggerProvider.CtxdLogger().Error(
			ctx,
			"Failed to close connection to Postgres",
			"error",
			err,
		)
	}
}
