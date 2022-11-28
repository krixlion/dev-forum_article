package rabbitmq

import (
	"context"
	"fmt"

	"github.com/krixlion/dev-forum_article/pkg/event"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	createdEvent = ".event.created"
	deletedEvent = ".event.deleted"
	updatedEvent = ".event.updated"
)

func (mq *RabbitMQ) Publish(ctx context.Context, e event.Event) error {
	msg, err := messageFromEvent(e)
	if err != nil {
		return err
	}

	return mq.publish(ctx, msg)
}

func (mq *RabbitMQ) publish(ctx context.Context, msg Message) error {
	ch := mq.channel()
	defer ch.Close()

	_, err := mq.breaker.Execute(func() (interface{}, error) {
		return nil, ch.ExchangeDeclare(
			msg.ExchangeName, // name
			msg.ExchangeType, // type
			true,             // durable
			false,            // auto-deleted
			false,            // internal
			false,            // no-wait
			nil,              // arguments
		)
	})
	if err != nil {
		return err
	}

	_, err = mq.breaker.Execute(func() (interface{}, error) {
		return nil, ch.PublishWithContext(ctx,
			msg.ExchangeName, // exchange
			msg.RoutingKey,   // routing key
			false,            // mandatory
			false,            // immediate
			amqp.Publishing{
				ContentType: string(msg.ContentType),
				Body:        msg.Body,
			},
		)
	})
	if err != nil {
		return err
	}
	return nil
}

// RetryEnqueue appends a message to the RetryQueue and throws an error if the queue is full.
func (mq *RabbitMQ) retryEnqueue(msg Message) error {
	select {
	case mq.retryC <- msg:
		return nil
	default:
		return fmt.Errorf("retry queue is full")
	}
}

func (mq *RabbitMQ) subscribe(ctx context.Context, consumer string, r Route) (<-chan Message, error) {
	ch := mq.channel()

	defer ch.Close()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	q, err := mq.breaker.Execute(func() (interface{}, error) {
		return ch.QueueDeclare(
			r.QueueName, // name
			false,       // durable
			false,       // delete when unused
			true,        // exclusive
			false,       // no-wait
			nil,         // arguments
		)
	})
	if err != nil {
		return nil, err
	}
	queue := q.(amqp.Queue)

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	_, err = mq.breaker.Execute(func() (interface{}, error) {
		return nil, ch.QueueBind(
			queue.Name,     // queue name
			r.RoutingKey,   // routing key
			r.ExchangeName, // exchange
			false,          // Immidiate
			nil,            // Additional args
		)
	})
	if err != nil {
		return nil, err
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	c, err := mq.breaker.Execute(func() (interface{}, error) {
		return ch.Consume(
			queue.Name, // queue
			consumer,   // consumer
			false,      // auto ack
			false,      // exclusive
			false,      // no local
			false,      // no wait
			nil,        // args
		)
	})
	if err != nil {
		return nil, err
	}

	deliveryC := c.(<-chan amqp.Delivery)
	msgs := make(chan Message)

	go func() {
		for delivery := range deliveryC {
			delivery.Ack(false)
			msgs <- Message{
				Body:        delivery.Body,
				ContentType: ContentType(delivery.ContentType),
				Timestamp:   delivery.Timestamp,
				Route:       r,
			}
			if ctx.Err() != nil {
				close(msgs)
				ch.Close()
				return
			}
		}
	}()

	return msgs, nil
}
