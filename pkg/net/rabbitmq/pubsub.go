package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/krixlion/dev-forum_article/pkg/event"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (mq *RabbitMQ) Publish(ctx context.Context, timeout time.Duration, msg Message) error {
	ch := mq.channel()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	defer ch.Close()

	_, err := mq.breaker.Execute(func() (interface{}, error) {
		err := ch.PublishWithContext(ctx,
			msg.Exchange,   // exchange
			msg.RoutingKey, // routing key
			false,          // mandatory
			false,          // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(msg.Body),
			},
		)
		if err != nil {
			return nil, err
		}
		return nil, nil
	})

	return err
}

func (mq *RabbitMQ) PublishOnQueue(ctx context.Context, timeout time.Duration, msg Message, queueName string) (err error) {
	ch := mq.channel()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	defer ch.Close()

	q, err := mq.breaker.Execute(func() (interface{}, error) {
		return ch.QueueDeclare(
			queueName, // our queue name
			false,     // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
	})
	if err != nil {
		return err
	}
	queue := q.(amqp.Queue)

	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Publishes a message onto the queue.
	_, err = mq.breaker.Execute(func() (interface{}, error) {
		err := ch.PublishWithContext(ctx,
			msg.Exchange, // exchange
			queue.Name,   // routing key
			false,        // mandatory
			false,        // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(msg.Body),
			},
		)
		return nil, err
	})

	return err
}

// RetryLater appends a message to the RetryQueue and throws an error if the queue is full.
func (mq *RabbitMQ) RetryLater(msg Message) error {
	select {
	case mq.retryC <- msg:
		return nil
	default:
		return fmt.Errorf("retry queue is full")
	}
}

func (mq *RabbitMQ) Subscribe(ctx context.Context, consumer string, r Route) (<-chan event.Event, error) {
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
			queue.Name,   // queue name
			r.RoutingKey, // routing key
			r.Exchange,   // exchange
			false,        // Immidiate
			nil,          // Additional args
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
	events := make(chan event.Event)

	go func() {
		for delivery := range deliveryC {
			delivery.Ack(false)
			events <- event.Event{
				Type:      delivery.Type,
				Body:      string(delivery.Body),
				Timestamp: delivery.Timestamp,
			}
			if ctx.Err() != nil {
				close(events)
				ch.Close()
				return
			}
		}
	}()

	return events, nil
}
