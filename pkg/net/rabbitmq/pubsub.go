package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/krixlion/dev-forum_article/pkg/event"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (mq *RabbitMQ) Publish(ctx context.Context, e event.Event) error {

	msg, err := makeMessageFromEvent(e)
	if err != nil {
		return err
	}

	err = validateMessage(msg)
	if err != nil {
		return err
	}

	err = mq.prepareExchange(ctx, msg)
	if err != nil {
		return err
	}

	err = mq.publish(ctx, msg)
	if err != nil {
		return err
	}

	return nil
}

// EnRoute validates a message and declares an RabbitMQ exchange derived from the message.
func (mq *RabbitMQ) prepareExchange(ctx context.Context, msg Message) error {

	ch := mq.channel()
	defer ch.Close()

	if err := ctx.Err(); err != nil {
		return err
	}

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

	return nil
}

func validateMessage(msg Message) error {
	return nil
}

func (mq *RabbitMQ) publish(ctx context.Context, msg Message) error {

	ch := mq.channel()
	defer ch.Close()

	if err := ctx.Err(); err != nil {
		return err
	}

	_, err := mq.breaker.Execute(func() (interface{}, error) {
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

	return err
}

func (mq *RabbitMQ) Consume(ctx context.Context, r Route) (<-chan event.Event, <-chan error) {
	events := make(chan event.Event)
	errors := make(chan error, 1)

	ch := mq.channel()

	if err := ctx.Err(); err != nil {
		errors <- err
		return events, errors
	}

	q, err := mq.breaker.Execute(func() (interface{}, error) {
		return ch.QueueDeclare(
			r.QueueName, // name
			false,       // durable
			false,       // delete when unused
			false,       // exclusive
			false,       // no-wait
			nil,         // arguments
		)
	})
	if err != nil {
		errors <- err
		return events, errors
	}
	queue := q.(amqp.Queue)

	if err := ctx.Err(); err != nil {
		errors <- err
		return events, errors
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
		errors <- err
		return events, errors
	}

	if err := ctx.Err(); err != nil {
		errors <- err
		return events, errors
	}

	c, err := mq.breaker.Execute(func() (interface{}, error) {
		return ch.Consume(
			queue.Name,      // queue
			DefaultConsumer, // consumer
			false,           // auto ack
			false,           // exclusive
			false,           // no local
			false,           // no wait
			nil,             // args
		)
	})
	msgs := c.(<-chan amqp.Delivery)

	if err != nil {
		errors <- err
		return events, errors
	}

	go func() {
		for msg := range msgs {
			event := event.Event{}

			if msg.ContentType == string(ContentTypeJson) {
				if err := json.Unmarshal(msg.Body, &event); err != nil {
					errors <- err
					return
				}

				if err := ctx.Err(); err != nil {
					errors <- err
					return
				}
			}

			events <- event
		}
	}()

	return events, errors
}
