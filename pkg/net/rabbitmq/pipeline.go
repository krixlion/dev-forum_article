package rabbitmq

import (
	"context"
	"errors"

	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/tracing"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

// ResilientPublish returns an error only if the queue is full or if it failed to serialize the event.
func (mq *RabbitMQ) ResilientPublish(ctx context.Context, e event.Event) error {
	_, span := otel.Tracer(tracing.ServiceName).Start(ctx, "rabbitmq.ResilientPublish")
	defer span.End()

	msg := makeMessageFromEvent(e)

	if err := mq.enqueue(msg); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	return nil
}

// enqueue appends a message to the publishQueue and return a non-nil error if the queue is full.
func (mq *RabbitMQ) enqueue(msg Message) error {
	select {
	case mq.publishQueue <- msg:
		return nil
	default:
		return errors.New("publish queue is full")
	}
}

// tryToEnqueue is a helper method.
func (mq *RabbitMQ) tryToEnqueue(ctx context.Context, message Message, err error, logErrorMessage string) {
	if err := mq.enqueue(message); err != nil {
		mq.logger.Log(ctx, "Failed to enqueue message", "err", err)
	}

	mq.logger.Log(ctx, logErrorMessage, "err", err)
}

func (mq *RabbitMQ) publishPipelined(ctx context.Context, messages <-chan Message) {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "rabbitmq.publishPipelined")
	defer span.End()

	go func() {
		channel := mq.channel()
		defer channel.Close()

		for {
			select {
			case message := <-messages:
				go func() {
					callSucceded, err := mq.breaker.Allow()
					if err != nil {
						span.RecordError(err)
						span.SetStatus(codes.Error, err.Error())
						mq.tryToEnqueue(ctx, message, err, "Failed to publish msg")
						return
					}

					err = channel.PublishWithContext(ctx,
						message.ExchangeName, // exchange
						message.RoutingKey,   // routing key
						false,                // mandatory
						false,                // immediate
						amqp.Publishing{
							ContentType: string(message.ContentType),
							Body:        message.Body,
						},
					)
					if err != nil {
						span.RecordError(err)
						span.SetStatus(codes.Error, err.Error())

						if isConnectionError(err.(*amqp.Error)) {
							callSucceded(false)
						}
						// Error did not render broker unavailable.
						callSucceded(true)

						mq.tryToEnqueue(ctx, message, err, "Failed to publish msg")
						return
					}
					callSucceded(true)
				}()

			case <-ctx.Done():
				channel.Close()
				return
			}
		}
	}()
}

func (mq *RabbitMQ) prepareExchangePipelined(ctx context.Context, msgs <-chan Message) <-chan Message {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "rabbitmq.prepareExchangePipelined")
	defer span.End()

	preparedMessages := make(chan Message)

	go func() {
		channel := mq.channel()
		defer channel.Close()

		for {
			select {
			case message := <-msgs:
				go func() {
					callSucceded, err := mq.breaker.Allow()
					if err != nil {
						mq.tryToEnqueue(ctx, message, err, "Failed to prepare exchange before publishing")
					}

					err = channel.ExchangeDeclare(
						message.ExchangeName, // name
						message.ExchangeType, // type
						true,                 // durable
						false,                // auto-deleted
						false,                // internal
						false,                // no-wait
						nil,                  // arguments
					)
					if err != nil {
						if isConnectionError(err.(*amqp.Error)) {
							callSucceded(false)
						}
						// Error did not render broker unavailable.
						callSucceded(true)

						mq.tryToEnqueue(ctx, message, err, "Failed to declare exchange")
						return
					}
					callSucceded(true)

					preparedMessages <- message
				}()

			case <-ctx.Done():
				return
			}
		}
	}()
	return preparedMessages
}
