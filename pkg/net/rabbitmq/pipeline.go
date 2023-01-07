package rabbitmq

import (
	"context"
	"errors"

	amqp "github.com/rabbitmq/amqp091-go"
)

// enqueue appends a message to the publishQueue and return a non-nil error if the queue is full.
func (mq *RabbitMQ) Enqueue(msg Message) error {
	select {
	case mq.publishQueue <- msg:
		return nil
	default:
		return errors.New("publish queue is full")
	}
}

func (mq *RabbitMQ) tryToEnqueue(ctx context.Context, message Message, err error, logErrorMessage string) {
	if err := mq.Enqueue(message); err != nil {
		mq.logger.Log(ctx, "Failed to enqueue message", "err", err)
	}

	mq.logger.Log(ctx, logErrorMessage, "err", err)
}

func (mq *RabbitMQ) publishPipelined(ctx context.Context, messages <-chan Message) {
	go func() {
		channel := mq.askForChannel()
		defer channel.Close()

		limiter := make(chan struct{}, mq.config.MaxWorkers)

		for {
			select {
			case message := <-messages:
				limiter <- struct{}{}
				go func() {
					ctx, span := mq.tracer.Start(ctx, "rabbitmq.publishPipelined")
					defer span.End()
					defer func() { <-limiter }()

					callSucceded, err := mq.breaker.Allow()
					if err != nil {
						setSpanErr(span, err)
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
							Timestamp:   message.Timestamp,
						},
					)
					if err != nil {
						setSpanErr(span, err)

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
	preparedMessages := make(chan Message)

	go func() {
		channel := mq.askForChannel()
		defer channel.Close()
		limiter := make(chan struct{}, mq.config.MaxWorkers)

		for {
			select {
			case message := <-msgs:
				limiter <- struct{}{}
				go func() {
					ctx, span := mq.tracer.Start(ctx, "rabbitmq.prepareExchangePipelined")
					defer span.End()
					defer func() { <-limiter }()

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

						setSpanErr(span, err)
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
