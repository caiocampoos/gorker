package client

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQClient struct {
	Connection *amqp.Connection
	mutex      sync.Mutex
	ctx        context.Context
	cancel     context.CancelFunc
}

type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

type BindingConfig struct {
	QueueName    string
	ExchangeName string
	RoutingKey   string
	NoWait       bool
	Args         amqp.Table
}

type ExchangeConfig struct {
	Name       string
	Kind       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Args       amqp.Table
}

func NewRabbitMQClient(
	ctx context.Context,
	url string,
) (*RabbitMQClient, error) {
	clientCtx, cancel := context.WithCancel(ctx)
	client := &RabbitMQClient{
		ctx:    clientCtx,
		cancel: cancel,
	}

	err := client.connect(url)
	if err != nil {
		return nil, err
	}

	go client.reconnect(url)

	return client, nil
}

func (client *RabbitMQClient) connect(url string) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	conn, err := amqp.Dial(url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	defer func() {
		if err != nil && conn != nil {
			conn.Close()
		}
	}()

	client.Connection = conn
	log.Print("RabbitMQ connected successfully")

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	log.Print("RabbitMQ exchanges declared successfully")
	return nil
}

func (client *RabbitMQClient) reconnect(url string) {
	var notifyCloseChan chan *amqp.Error

	for {
		client.mutex.Lock()
		if client.Connection != nil && !client.Connection.IsClosed() {
			notifyCloseChan = client.Connection.NotifyClose(make(chan *amqp.Error, 1))
		}
		client.mutex.Unlock()

		select {
		case <-client.ctx.Done():
			log.Print("Reconnection goroutine exiting")
			return
		case err, ok := <-notifyCloseChan:
			if !ok {
				log.Print("NotifyClose channel closed")
				return
			}
			log.Printf("Connection closed: %v", err)

			// implements attempt to reconnect
			//avoid leaks
			for {
				select {
				case <-client.ctx.Done():
					log.Print("Reconnection goroutine exiting")
					return
				case <-time.After(5 * time.Second):
					log.Print("Attempting to reconnect...")
					if err := client.connect(url); err != nil {
						log.Printf("Failed to reconnect: %v", err)
					} else {
						log.Print("Reconnected successfully")
						break // Break the reconnection attempt loop
					}
				}
			}
		}
	}
}

func (client *RabbitMQClient) Close() {
	client.cancel()
	client.mutex.Lock()
	defer client.mutex.Unlock()
	if client.Connection != nil && !client.Connection.IsClosed() {
		client.Connection.Close()
		log.Print("RabbitMQ connection closed")
	}
}

func (client *RabbitMQClient) DeclareQueuesAndBindings(
	queues []QueueConfig,
	bindings []BindingConfig,
) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	ch, err := client.Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	for _, q := range queues {
		_, err := ch.QueueDeclare(
			q.Name,
			q.Durable,
			q.AutoDelete,
			q.Exclusive,
			q.NoWait,
			q.Args,
		)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", q.Name, err)
		}
	}

	for _, b := range bindings {
		err := ch.QueueBind(
			b.QueueName,
			b.RoutingKey,
			b.ExchangeName,
			b.NoWait,
			b.Args,
		)
		if err != nil {
			return fmt.Errorf(
				"failed to bind queue %s to exchange %s: %w",
				b.QueueName,
				b.ExchangeName,
				err,
			)
		}
	}

	log.Print("Queues and bindings declared successfully")
	return nil
}

func (client *RabbitMQClient) DeclareExchanges(exchanges []ExchangeConfig) error {
	client.mutex.Lock()
	defer client.mutex.Unlock()

	ch, err := client.Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}
	defer ch.Close()

	for _, ex := range exchanges {
		err = ch.ExchangeDeclare(
			ex.Name,
			ex.Kind,
			ex.Durable,
			ex.AutoDelete,
			ex.Internal,
			ex.NoWait,
			ex.Args,
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", ex.Name, err)
		}
	}

	log.Print("Exchanges declared successfully")
	return nil
}
