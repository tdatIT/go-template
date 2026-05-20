package main

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"sync"

	_ "go.uber.org/automaxprocs"

	server "github.com/tdatIT/go-template/internal"
)

func main() {
	fmt.Println("starting app...")
	serv := server.NewServer()
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		fmt.Println("gracefully shutting down...")
		serv.Shutdown()
	}()

	var wg sync.WaitGroup
	wg.Go(func() {
		if err := serv.API().Start(serv.Config().Server.Port); err != nil {
			slog.Error("failed to start server", slog.String("error", err.Error()))
			os.Exit(1)
		}
	})

	wg.Wait()
	log.Println(" server shutdown successfully")
}
