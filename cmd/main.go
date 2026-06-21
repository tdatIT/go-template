package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "go.uber.org/automaxprocs"

	server "github.com/tdatIT/go-template/internal"
)

func main() {
	fmt.Println("starting app...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serv, err := server.NewServer()
	if err != nil {
		slog.Error("error while creating server", slog.String("error", err.Error()))
		os.Exit(1)
	}

	var wg sync.WaitGroup

	// HTTP component
	wg.Go(func() {
		if err := serv.API().Start(serv.Config().Server.Port); err != nil {
			slog.Error("http server stopped", slog.String("error", err.Error()))
		}
	})

	// Worker component
	wg.Go(func() {
		if err := serv.Workers().StartGroup(ctx); err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("worker group stopped with error", slog.String("error", err.Error()))
		}
	})

	// Block until signal arrives, then start shutdown sequence.
	<-ctx.Done()
	stop() // release signal resources early

	fmt.Println("gracefully shutting down...")
	serv.Shutdown()

	wg.Wait()
	fmt.Println("server shutdown successfully")
	os.Exit(0)
}
