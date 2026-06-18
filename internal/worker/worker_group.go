package worker

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/sync/errgroup"
)

// Worker is the interface every child worker must implement.
type Worker interface {
	Start(ctx context.Context) error
}

// WorkerGroup manages a collection of Workers and runs them concurrently.
// It is the worker-side counterpart to the HTTP router — same concept, different transport.
type WorkerGroup struct {
	workers []Worker
}

func NewWorkerGroup() *WorkerGroup {
	return &WorkerGroup{}
}

// Register adds one or more workers to the group.
// Must be called before StartGroup.
func (g *WorkerGroup) Register(workers ...Worker) {
	g.workers = append(g.workers, workers...)
}

// StartGroup starts all registered workers concurrently and blocks until they all stop.
// If any worker returns a non-nil error the context passed to the others is cancelled,
// all workers stop, and StartGroup returns the first error wrapped with context.
func (g *WorkerGroup) StartGroup(ctx context.Context) error {
	if len(g.workers) == 0 {
		slog.Warn("worker group: no workers registered")
		return nil
	}

	eg, groupCtx := errgroup.WithContext(ctx)
	for _, w := range g.workers {
		w := w
		eg.Go(func() error {
			return w.Start(groupCtx)
		})
	}

	slog.Info("worker group started", slog.Int("count", len(g.workers)))
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("worker group: %w", err)
	}
	return nil
}
