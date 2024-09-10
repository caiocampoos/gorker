package main

import (
 amqp "github.com/rabbitmq/amqp091-go"
  "log"
  "gonsumer/consumer"
  
)

func handler(delivery amqp.Delivery) {
  log.Printf("Received a message: %s", delivery.Body)
}

func main() {
  done := make(chan bool)
  go consumer.Consumer("go-queue", "go-consumer", handler)
  log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
  <-done
}
