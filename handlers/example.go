package handlers

import (
	"go-worker/internals"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func Handler(delivery amqp.Delivery) {
	internals.PrintMemoryUsage("Start of program")

	log.Printf("Received a message: %s", delivery.Body)

	internals.PrintMemoryUsage("After receiving message")

	// Do some work
	time.Sleep(time.Second)

	internals.PrintMemoryUsage("After sleep")

	//do some more work?

	internals.PrintMemoryUsage("After publishing")
}
