// cmd/worker/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-worker/client"
	"go-worker/internals"
	"go-worker/worker"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading .env file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal")
		cancel() // Cancels the context, signaling workers to stop
	}()

	rabbitmqConfig := internals.RabbitMqConfigGet()

	url := fmt.Sprintf(
		"amqp://%s:%s@%s:5672/",
		rabbitmqConfig.RABBITMQ_USER,
		rabbitmqConfig.RABBITMQ_PASS,
		rabbitmqConfig.RABBITMQ_HOST,
	)

	// Initialize RabbitMQ client
	rabbitMQClient, err := client.NewRabbitMQClient(ctx, url)
	if err != nil {
		log.Fatalf("Failed to create RabbitMQ client: %v", err)
	}
	defer rabbitMQClient.Close()

	// Initialize and start consumer workers
	workers, err := worker.InitializeWorkers(rabbitMQClient)
	if err != nil {
		log.Fatalf("Failed to initialize workers: %v", err)
	}

	for _, w := range workers {
		if err := w.Start(ctx); err != nil {
			log.Fatalf("Failed to start worker  %v", err)
		}
	}

	log.Println("All workers started. Waiting for messages...")

	<-ctx.Done() // Wait until context is canceled (CTRL+C)

	log.Println("Shutting down workers")

	// Stop all workers and wait for cleanup
	for _, w := range workers {
		log.Printf("Stopping worker")
		if err := w.Stop(ctx); err != nil {
			log.Printf("Error stopping worker  %v", err)
		}
	}

	log.Println("All workers stopped")
}
