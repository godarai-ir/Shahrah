// Package bootstrap assembles and owns the control-plane application graph.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// StartFunc starts a lifecycle-managed worker with the application context.
type StartFunc func(context.Context) error

// StopFunc releases a worker or resource during reverse-order shutdown.
type StopFunc func(context.Context) error

type lifecycleEntry struct {
	name   string
	start  StartFunc
	stop   StopFunc
	active bool
}

// Lifecycle owns resource cleanup and worker start/stop ordering.
type Lifecycle struct {
	mu      sync.Mutex
	entries []lifecycleEntry
	started bool
	stopped bool
	cancel  context.CancelFunc
}

// NewLifecycle creates an empty application lifecycle.
func NewLifecycle() *Lifecycle {
	return &Lifecycle{}
}

// AddResource registers an already-acquired resource for cleanup.
func (l *Lifecycle) AddResource(name string, stop StopFunc) error {
	return l.add(lifecycleEntry{name: name, stop: stop, active: true})
}

// AddWorker registers a worker that starts with the application.
func (l *Lifecycle) AddWorker(name string, start StartFunc, stop StopFunc) error {
	if start == nil {
		return fmt.Errorf("lifecycle: worker %q has no start function", name)
	}

	return l.add(lifecycleEntry{name: name, start: start, stop: stop})
}

// AddLoop registers a context-bound loop and waits for it during shutdown.
func (l *Lifecycle) AddLoop(name string, run func(context.Context)) error {
	if run == nil {
		return fmt.Errorf("lifecycle: loop %q has no run function", name)
	}

	done := make(chan struct{})

	return l.AddWorker(name, func(ctx context.Context) error {
		go func() {
			defer close(done)

			run(ctx)
		}()

		return nil
	}, func(ctx context.Context) error {
		select {
		case <-done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
}

func (l *Lifecycle) add(entry lifecycleEntry) error {
	if l == nil {
		return errors.New("lifecycle is nil")
	}

	if entry.name == "" {
		return errors.New("lifecycle entry name is empty")
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.started || l.stopped {
		return fmt.Errorf("lifecycle: cannot register %q after start", entry.name)
	}

	for _, existing := range l.entries {
		if existing.name == entry.name {
			return fmt.Errorf("lifecycle: duplicate entry %q", entry.name)
		}
	}

	l.entries = append(l.entries, entry)

	return nil
}

// Start starts workers in registration order.
func (l *Lifecycle) Start(parent context.Context) error {
	if l == nil {
		return errors.New("lifecycle is nil")
	}

	if parent == nil {
		parent = context.Background()
	}

	l.mu.Lock()
	if l.started || l.stopped {
		l.mu.Unlock()

		return errors.New("lifecycle has already started or stopped")
	}

	workerCtx, cancel := context.WithCancel(parent)
	l.started = true
	l.cancel = cancel

	for index := range l.entries {
		entry := &l.entries[index]
		if entry.start == nil {
			continue
		}

		if err := entry.start(workerCtx); err != nil {
			cancel()
			l.mu.Unlock()

			return fmt.Errorf("lifecycle: start %q: %w", entry.name, err)
		}

		entry.active = true
	}
	l.mu.Unlock()

	return nil
}

// Stop cancels the application context and releases active entries in reverse order.
func (l *Lifecycle) Stop(ctx context.Context) error {
	if l == nil {
		return nil
	}

	if ctx == nil {
		ctx = context.Background()
	}

	l.mu.Lock()
	if l.stopped {
		l.mu.Unlock()

		return nil
	}

	l.stopped = true
	if l.cancel != nil {
		l.cancel()
	}

	entries := append([]lifecycleEntry(nil), l.entries...)
	l.mu.Unlock()

	var stopErrors []error

	for index := len(entries) - 1; index >= 0; index-- {
		entry := entries[index]
		if !entry.active || entry.stop == nil {
			continue
		}

		if ctxErr := ctx.Err(); ctxErr != nil {
			stopErrors = append(stopErrors, ctxErr)

			break
		}

		if err := entry.stop(ctx); err != nil {
			stopErrors = append(stopErrors, fmt.Errorf("stop %q: %w", entry.name, err))
		}
	}

	return errors.Join(stopErrors...)
}
