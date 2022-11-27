package rabbitmq

import (
	"context"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sony/gobreaker"
)

const (
	ConnectionError = 1
	ChannelError    = 2
)

func DefaultBreakerSettings() gobreaker.Settings {
	return gobreaker.Settings{
		MaxRequests: 20,
		Interval:    time.Second * 30,
		Timeout:     time.Second * 5,
		IsSuccessful: func(err error) bool {
			if err != context.Canceled && err != context.DeadlineExceeded {
				return false
			}
			return true
		},
	}
}

func (mq *RabbitMQ) setUpChannel() (*amqp.Channel, error) {
	ch, err := mq.connection.Channel()
	mq.mu.Lock()
	mq.errC = ch.NotifyClose(make(chan *amqp.Error, 16))
	defer mq.mu.Unlock()
	return ch, err
}

func errorType(code int) int {
	switch code {
	case
		amqp.ContentTooLarge,    // 311
		amqp.NoConsumers,        // 313
		amqp.AccessRefused,      // 403
		amqp.NotFound,           // 404
		amqp.ResourceLocked,     // 405
		amqp.PreconditionFailed: // 406
		return ChannelError

	case
		amqp.ConnectionForced, // 320
		amqp.InvalidPath,      // 402
		amqp.FrameError,       // 501
		amqp.SyntaxError,      // 502
		amqp.CommandInvalid,   // 503
		amqp.ChannelError,     // 504
		amqp.UnexpectedFrame,  // 505
		amqp.ResourceError,    // 506
		amqp.NotAllowed,       // 530
		amqp.NotImplemented,   // 540
		amqp.InternalError:    // 541
		fallthrough

	default:
		return ConnectionError
	}
}

func isConnectionError(err *amqp.Error) bool {
	return errorType(err.Code) == ConnectionError
}

// func isChannelError(err *amqp.Error) bool {
// 	return errorType(err.Code) == ChannelError
// }
