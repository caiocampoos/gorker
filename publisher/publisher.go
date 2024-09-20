package publisher

import (
	"context"
	"fmt"
	"sync"

	"go-worker/client"

	amqp "github.com/rabbitmq/amqp091-go"
)

type PublisherConfig struct {
	ExchangeName string
	RoutingKey   string
	Mandatory    bool
	Immediate    bool
	ContentType  string
}

type Publisher struct {
	client          *client.RabbitMQClient
	publisherConfig PublisherConfig
	channel         *amqp.Channel
	mutex           sync.Mutex
}

func NewPublisher(
	client *client.RabbitMQClient,
	publisherConfig PublisherConfig,
) (*Publisher, error) {
	ch, err := client.Connection.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	publisher := &Publisher{
		client:          client,
		publisherConfig: publisherConfig,
		channel:         ch,
	}

	return publisher, nil
}

func (p *Publisher) Publish(ctx context.Context, body []byte) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.channel == nil || p.channel.IsClosed() {
		ch, err := p.client.Connection.Channel()
		if err != nil {
			return fmt.Errorf("failed to reopen channel: %w", err)
		}
		p.channel = ch
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := p.channel.Publish(
		p.publisherConfig.ExchangeName,
		p.publisherConfig.RoutingKey,
		p.publisherConfig.Mandatory,
		p.publisherConfig.Immediate,
		amqp.Publishing{
			ContentType: p.publisherConfig.ContentType,
			Body:        body,
		},
	)
	if err != nil {
		// Close the channel on error
		p.channel.Close()
		p.channel = nil
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
