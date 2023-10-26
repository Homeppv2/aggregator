package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"homepp/aggregator/internal"
	"homepp/aggregator/internal/app"
	"homepp/aggregator/internal/controller"
	"homepp/aggregator/internal/gateway"
)

func main() {
	app.Run()
}

func run() {
	cfg := app.GetConfig()
	eventPublisher, err := controller.NewEventPublisher(cfg.Publisher.URL())
	if err != nil {
		log.Fatal(err)
	}
	socketGateway := gateway.NewSocketGateway(
		cfg.MemoryStorage.KeyPrefix,
		cfg.MemoryStorage.URL(),
	)
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr: cfg.Server.URL(),
		Handler: internal.Server{
			Logf:           log.Printf,
			EventPublisher: eventPublisher,
			SocketGateway:  socketGateway,
		},
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	select {
	case err := <-errCh:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigCh:
		log.Printf("terminating: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	server.Shutdown(ctx)
}
