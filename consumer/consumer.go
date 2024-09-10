package consumer

import (
 amqp "github.com/rabbitmq/amqp091-go"
  "sync"
  "log"
)

func failOnError(err error, msg string) {
 if err != nil {
  log.Fatalf("%s: %s", msg, err)
 }
}

func Consumer (queueName string, consumerName string, handler func(amqp.Delivery)) {

  var wg sync.WaitGroup
  conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
  failOnError(err, "Failed to connect to RabbitMQ")
  defer conn.Close()

  ch, err := conn.Channel()
  failOnError(err, "Failed to open a channel")
  defer ch.Close()


  q, err := ch.QueueDeclare(
   queueName,
   false,     
   false,      
   false,      
   false,      
   nil,        
  )
  failOnError(err, "Failed to declare a queue")

  msgs, err := ch.Consume(
    q.Name, // name of the queue again
    consumerName,     
    true,   
    false,  
    false,  
    false,  
    nil,    
  )

  failOnError(err, "Failed to register a consumer")

  wg.Add(1)
  go func() {
    for d := range msgs {
    defer wg.Done()
    handler(d)
    }
  }()

  wg.Wait()
}

