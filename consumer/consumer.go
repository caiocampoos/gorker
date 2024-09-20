package consumer

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go-worker/client"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ConsumerConfig struct {
	QueueName     string
	ConsumerTag   string
	AutoAck       bool
	PrefetchCount int
}

type Consumer struct {
	client         *client.RabbitMQClient
	consumerConfig ConsumerConfig
	handler        func(amqp.Delivery)
	ctx            context.Context
	wg             sync.WaitGroup
}

func NewConsumer(
	client *client.RabbitMQClient,
	consumerConfig ConsumerConfig,
	handler func(amqp.Delivery),
) *Consumer {
	return &Consumer{
		client:         client,
		consumerConfig: consumerConfig,
		handler:        handler,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	ch, err := c.client.Connection.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %w", err)
	}

	go func() {
		<-ctx.Done()
		ch.Close()
	}()

	err = ch.Qos(
		c.consumerConfig.PrefetchCount,
		0,
		false,
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := ch.Consume(
		c.consumerConfig.QueueName,
		c.consumerConfig.ConsumerTag,
		c.consumerConfig.AutoAck,
		false, // Exclusive
		false, // No-local
		false, // No-wait
		nil,   // Args
	)
	if err != nil {
		ch.Close()
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				c.handler(msg)
			}
		}
	}()

	return nil
}

func (c *Consumer) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Minute)
	defer cancel()

	c.wg.Wait()

	return nil
}
