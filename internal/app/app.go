package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"homepp/aggregator/internal"
)

func Run() {
	cfg := GetConfig()
	eventPublisher, err := internal.NewEventPublisher(cfg.Publisher.URL())
	if err != nil {
		log.Fatal(err)
	}
	socketGateway := internal.NewSocketGateway(
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
