package bootstrap

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

const defaultShutdownTimeout = 10 * time.Second

// Application is the fully validated control-plane runtime.
type Application struct {
	HTTP      *fiber.App
	Address   string
	lifecycle *Lifecycle

	shutdownOnce sync.Once
	shutdownErr  error
}

// Start starts every registered background worker.
func (a *Application) Start(ctx context.Context) error {
	if a == nil || a.HTTP == nil || a.lifecycle == nil {
		return errors.New("bootstrap: application is incomplete")
	}

	if err := a.lifecycle.Start(ctx); err != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownTimeout)
		defer cancel()

		return errors.Join(err, a.lifecycle.Stop(shutdownCtx))
	}

	return nil
}

// Listen serves the configured HTTP address.
func (a *Application) Listen() error {
	if a == nil || a.HTTP == nil {
		return errors.New("bootstrap: HTTP application is unavailable")
	}

	return a.HTTP.Listen(a.Address)
}

// Shutdown stops HTTP admission and releases workers and resources once.
func (a *Application) Shutdown(ctx context.Context) error {
	if a == nil {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	a.shutdownOnce.Do(func() {
		timeout := defaultShutdownTimeout
		if deadline, ok := ctx.Deadline(); ok {
			timeout = time.Until(deadline)
			if timeout <= 0 {
				timeout = time.Millisecond
			}
		}

		var httpErr error
		if a.HTTP != nil {
			httpErr = a.HTTP.ShutdownWithTimeout(timeout)
		}

		a.shutdownErr = errors.Join(httpErr, a.lifecycle.Stop(ctx))
	})

	return a.shutdownErr
}
