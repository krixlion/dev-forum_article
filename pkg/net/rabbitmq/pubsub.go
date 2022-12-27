package rabbitmq

import (
	"context"
	"encoding/json"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/tracing"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

func (mq *RabbitMQ) Publish(ctx context.Context, e event.Event) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "rabbitmq.Publish")
	defer span.End()

	msg := makeMessageFromEvent(e)

	err := mq.prepareExchange(ctx, msg)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	err = mq.publish(ctx, msg)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	return nil
}

// prepareExchange validates a message and declares a RabbitMQ exchange derived from the message.
func (mq *RabbitMQ) prepareExchange(ctx context.Context, msg Message) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "rabbitmq.prepareExchange")
	defer span.End()

	ch := mq.channel()
	defer ch.Close()

	if err := ctx.Err(); err != nil {
		return err
	}

	succeded, err := mq.breaker.Allow()
	if err != nil {
		tracing.SetSpanErr(span, err)
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
		tracing.SetSpanErr(span, err)

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
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "rabbitmq.publish")
	defer span.End()

	ch := mq.channel()
	defer ch.Close()

	if err := ctx.Err(); err != nil {
		return err
	}

	succeded, err := mq.breaker.Allow()
	if err != nil {
		tracing.SetSpanErr(span, err)
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

		tracing.SetSpanErr(span, err)
		return err
	}
	succeded(true)
	return nil
}

func (mq *RabbitMQ) Consume(ctx context.Context, command string, entity entity.EntityName, eType event.EventType) (<-chan event.Event, error) {
	events := make(chan event.Event)
	ch := mq.channel()

	if err := ctx.Err(); err != nil {
		return events, err
	}

	route := makeRouteFromEvent(event.Event{
		Entity: entity,
		Type:   eType,
	})

	callSucceded, err := mq.breaker.Allow()
	if err != nil {
		return nil, err
	}

	queue, err := ch.QueueDeclare(
		command, // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		if isConnectionError(err.(*amqp.Error)) {
			callSucceded(false)
		}
		// Error did not render broker unavailable.
		callSucceded(true)

		return nil, err
	}
	callSucceded(true)

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	callSucceded, err = mq.breaker.Allow()
	if err != nil {
		return nil, err
	}

	err = ch.QueueBind(
		queue.Name,         // queue name
		route.RoutingKey,   // routing key
		route.ExchangeName, // exchange
		false,              // Immidiate
		nil,                // Additional args
	)
	if err != nil {
		if isConnectionError(err.(*amqp.Error)) {
			callSucceded(false)
		}
		// Error did not render broker unavailable.
		callSucceded(true)

		return nil, err
	}
	callSucceded(true)

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	callSucceded, err = mq.breaker.Allow()
	if err != nil {
		return nil, err
	}

	messages, err := ch.Consume(
		queue.Name,      // queue
		mq.ConsumerName, // consumer
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

		return nil, err
	}
	callSucceded(true)

	go func() {
		for {
			select {
			case message := <-messages:
				go func() {
					ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "rabbitmq.Consume")
					defer span.End()

					event := event.Event{}
					if message.ContentType == string(ContentTypeJson) {
						if err := json.Unmarshal(message.Body, &event); err != nil {
							tracing.SetSpanErr(span, err)
							mq.logger.Log(ctx, "Failed to process message", "err", err)
							return
						}
					}

					message.Ack(false)
					events <- event
				}()
			case <-ctx.Done():
				close(events)
				return
			}
		}
	}()

	return events, nil
}
