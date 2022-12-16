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

// prepareExchange validates a message and declares a RabbitMQ exchange derived from the message.
func (mq *RabbitMQ) prepareExchange(ctx context.Context, msg Message) error {

	ch := mq.channel()
	defer ch.Close()

	if err := ctx.Err(); err != nil {
		return err
	}

	succeded, err := mq.breaker.Allow()
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare(
		msg.ExchangeName, // name
		msg.ExchangeType, // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		if isConnectionError(err.(*amqp.Error)) {
			succeded(false)
		}

		succeded(true)
		return err
	}
	succeded(true)

	return nil
}

func (mq *RabbitMQ) publish(ctx context.Context, msg Message) error {

	ch := mq.channel()
	defer ch.Close()

	if err := ctx.Err(); err != nil {
		return err
	}

	succeded, err := mq.breaker.Allow()
	if err != nil {
		return err
	}

	err = ch.PublishWithContext(ctx,
		msg.ExchangeName, // exchange
		msg.RoutingKey,   // routing key
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: string(msg.ContentType),
			Body:        msg.Body,
		},
	)
	if err != nil {
		if isConnectionError(err.(*amqp.Error)) {
			succeded(false)
		}
		// Error did not render broker unavailable.
		succeded(true)

		return err
	}
	succeded(true)

	return err
}

func (mq *RabbitMQ) Consume(ctx context.Context, r Route) (<-chan event.Event, <-chan error) {
	events := make(chan event.Event)
	errors := make(chan error, 1)

	ch := mq.channel()

	if err := ctx.Err(); err != nil {
		errors <- err
		close(errors)
		return events, errors
	}

	callSucceded, err := mq.breaker.Allow()
	if err != nil {
		return nil, errors
	}

	queue, err := ch.QueueDeclare(
		r.QueueName, // name
		false,       // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		if isConnectionError(err.(*amqp.Error)) {
			callSucceded(false)
		}
		// Error did not render broker unavailable.
		callSucceded(true)

		errors <- err
		return nil, errors
	}
	callSucceded(true)

	if err := ctx.Err(); err != nil {
		errors <- err
		return events, errors
	}

	callSucceded, err = mq.breaker.Allow()
	if err != nil {
		errors <- err
		return nil, errors
	}

	err = ch.QueueBind(
		queue.Name,     // queue name
		r.RoutingKey,   // routing key
		r.ExchangeName, // exchange
		false,          // Immidiate
		nil,            // Additional args
	)
	if err != nil {
		if isConnectionError(err.(*amqp.Error)) {
			callSucceded(false)
		}
		// Error did not render broker unavailable.
		callSucceded(true)

		errors <- err
		return events, errors
	}
	callSucceded(true)

	if err := ctx.Err(); err != nil {
		errors <- err
		return events, errors
	}

	callSucceded, err = mq.breaker.Allow()
	if err != nil {
		errors <- err
		return events, errors
	}

	messages, err := ch.Consume(
		queue.Name,      // queue
		DefaultConsumer, // consumer
		false,           // auto ack
		false,           // exclusive
		false,           // no local
		false,           // no wait
		nil,             // args
	)
	if err != nil {
		if isConnectionError(err.(*amqp.Error)) {
			callSucceded(false)
		}
		// Error did not render broker unavailable.
		callSucceded(true)

		errors <- err
		return events, errors
	}
	callSucceded(true)

	go func() {
		for message := range messages {
			event := event.Event{}

			if message.ContentType == string(ContentTypeJson) {
				if err := json.Unmarshal(message.Body, &event); err != nil {
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
