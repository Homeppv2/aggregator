package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"homepp/aggregator/internal"
	"homepp/aggregator/internal/config"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	config := config.GetConfig()
	eventPublisher, err := internal.NewEventPublisher(config.Publisher.URL())
	if err != nil {
		return err
	}
	socketGateway := internal.NewSocketGateway(
		config.MemoryStorage.KeyPrefix,
		config.MemoryStorage.URL(),
	)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr: config.Server.URL(),
		Handler: internal.Server{
			Logf:           log.Printf,
			EventPublisher: eventPublisher,
			SocketGateway:  socketGateway,
		},
		ReadTimeout:  time.Second * 10,
		WriteTimeout: time.Second * 10,
	}
	errc := make(chan error, 1)
	go func() {
		errc <- server.ListenAndServe()
	}()

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt)
	select {
	case err := <-errc:
		log.Printf("failed to serve: %v", err)
	case sig := <-sigc:
		log.Printf("terminating: %v", sig)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	return server.Shutdown(ctx)
}
