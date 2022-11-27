package event

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Handler interface {
	Close()
	Consumer
	Publisher
	Subscriber
}

type Consumer interface {
	Consume(ctx context.Context) (<-chan interface{}, error)
	ConsumeOnQueue(ctx context.Context) (<-chan Event, error)
}

type Publisher interface {
	Publish(ctx context.Context, msg []byte, exchangeName string, exchangeType string) error
	PublishOnQueue(ctx context.Context, msg []byte, queueName string) error
}

type Subscriber interface {
	Subscribe(ctx context.Context, exchangeName string, exchangeType string, consumerName string, handlerFunc func(amqp.Delivery)) error
	SubscribeToQueue(ctx context.Context, queueName string, consumerName string, handlerFunc func(amqp.Delivery)) error
}
