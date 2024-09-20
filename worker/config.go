package worker

import (
	"context"
	"go-worker/client"
	"go-worker/consumer"
	ocrhandlers "go-worker/handlers"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Worker interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type WorkerConfig struct {
	Name           string
	ConsumerConfig *consumer.ConsumerConfig
	Handler        func(amqp.Delivery)
}

var (
	Exchanges = []client.ExchangeConfig{
		{
			Name:       "reconnection-exchange",
			Kind:       "direct",
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			NoWait:     false,
			Args:       nil,
		},
	}

	Queues = []client.QueueConfig{
		{
			Name:       "go-queue",
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			NoWait:     false,
			Args:       nil,
		},
	}

	Bindings = []client.BindingConfig{
		{
			QueueName:    "go-queue",
			ExchangeName: "reconnection-exchange",
			RoutingKey:   "routing-key",
			NoWait:       false,
			Args:         nil,
		},
	}

	workerConfigs = []WorkerConfig{
		{
			Name: "my-consumer",
			ConsumerConfig: &consumer.ConsumerConfig{
				QueueName:     "go-queue",
				ConsumerTag:   "my-consumer",
				AutoAck:       false,
				PrefetchCount: 10,
			},
			Handler: ocrhandlers.Handler,
		},
	}
)

func InitializeWorkers(rabbitMQClient *client.RabbitMQClient) ([]Worker, error) {
	if err := rabbitMQClient.DeclareExchanges(Exchanges); err != nil {
		return nil, err
	}

	if err := rabbitMQClient.DeclareQueuesAndBindings(Queues, Bindings); err != nil {
		return nil, err
	}

	var workers []Worker

	for _, wc := range workerConfigs {
		if wc.ConsumerConfig == nil || wc.Handler == nil {
			log.Fatalf("ConsumerConfig or Handler is nil for worker %s", wc.Name)
		}
		c := consumer.NewConsumer(rabbitMQClient, *wc.ConsumerConfig, wc.Handler)
		w := &ConsumerWorker{
			Consumer: c,
			Name:     wc.Name,
		}
		workers = append(workers, w)
	}

	return workers, nil
}

type ConsumerWorker struct {
	Consumer *consumer.Consumer
	Name     string
}

func (cw *ConsumerWorker) Start(ctx context.Context) error {
	return cw.Consumer.Start(ctx)
}

func (cw *ConsumerWorker) Stop(ctx context.Context) error {
	return cw.Consumer.Stop(ctx)
}
