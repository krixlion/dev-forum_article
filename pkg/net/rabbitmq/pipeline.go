package rabbitmq

import (
	"context"
	"fmt"

	"github.com/krixlion/dev-forum_article/pkg/event"
	amqp "github.com/rabbitmq/amqp091-go"
)

// ResilientPublish returns an error only if the queue is full or if it failed to serialize the event.
func (mq *RabbitMQ) ResilientPublish(ctx context.Context, e event.Event) error {
	msg, err := makeMessageFromEvent(e)
	if err != nil {
		return err
	}

	return mq.enqueue(msg)
}

// enqueue appends a message to the publishQueue and return a non-nil error if the queue is full.
func (mq *RabbitMQ) enqueue(msg Message) error {
	select {
	case mq.publishQueue <- msg:
		return nil
	default:
		return fmt.Errorf("publish queue is full")
	}
}

// tryToEnqueue is a helper function.
func (mq *RabbitMQ) tryToEnqueue(message Message, err error, logErrorMessage string) {
	if err := mq.enqueue(message); err != nil {
		mq.logger.Log("Failed to enqueue message", "err", err)
	}

	mq.logger.Log(logErrorMessage, "err", err)
}

func (mq *RabbitMQ) publishPipelined(ctx context.Context, messages <-chan Message) {
	go func() {
		channel := mq.channel()
		defer channel.Close()

		for {
			select {
			case message := <-messages:
				go func() {
					callSucceded, err := mq.breaker.Allow()
					if err != nil {
						mq.tryToEnqueue(message, err, "Failed to publish msg")
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
						if isConnectionError(err.(*amqp.Error)) {
							callSucceded(false)
						}
						// Error did not render broker unavailable.
						callSucceded(true)

						mq.tryToEnqueue(message, err, "Failed to publish msg")
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
		channel := mq.channel()
		defer channel.Close()

		for {
			select {
			case message := <-msgs:
				go func() {
					callSucceded, err := mq.breaker.Allow()
					if err != nil {
						mq.tryToEnqueue(message, err, "Failed to prepare exchange before publishing")
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

						mq.tryToEnqueue(message, err, "Failed to declare exchange")
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
