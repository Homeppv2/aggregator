package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Homeppv2/aggregator-go/internal/app"
	"github.com/Homeppv2/aggregator-go/internal/controller"
)

func main() {
	run()
}

func run() {
	uriBroker := fmt.Sprintf("%s://%s:%s@%s:%s",
		os.Getenv("BROKER_PROTOCOL"),
		os.Getenv("BROKER_USERNAME"),
		os.Getenv("BROKER_PASSWORD"),
		os.Getenv("BROKER_HOST"),
		os.Getenv("BROKER_PORT"),
	)
	eventPublisher, err := controller.NewEventPublisher(uriBroker)
	if err != nil {
		log.Fatal(err)
	}
	uriServer := fmt.Sprintf("%s:%s", os.Getenv("AGGREGATOR_HOST"), os.Getenv("AGGREGATOR_PORT"))
	server := &http.Server{
		Addr: uriServer,
		Handler: app.Server{
			Logf:           log.Printf,
			EventPublisher: eventPublisher,
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
