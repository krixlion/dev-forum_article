package rabbitmq

import (
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	connectionError = 1
	channelError    = 2
)

type Config struct {
	QueueSize         int           // Max number of messages internally queued for publishing.
	ReconnectInterval time.Duration // Time between reconnect attempts.

	// Settings for the internal circuit breaker.
	MaxRequests   uint32        // Number of requests allowed to half-open state.
	ClearInterval time.Duration // Time after which failed calls count is cleared.
	ClosedTimeout time.Duration // Time after which closed state becomes half-open.
}

func DefaultConfig(consumer string) Config {
	return Config{
		QueueSize:         100,              // Max number of messages internally queued for publishing.
		ReconnectInterval: time.Second * 2,  // Time between reconnect attempts.
		MaxRequests:       10,               // Number of requests allowed to half-open state.
		ClearInterval:     time.Second * 10, // Time after which failed calls count is cleared.
		ClosedTimeout:     time.Second * 10, // Time after which closed state becomes half-open.
	}
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
		return channelError

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
		return connectionError
	}
}

func isConnectionError(err *amqp.Error) bool {
	return errorType(err.Code) == connectionError
}

// func isChannelError(err *amqp.Error) bool {
// 	return errorType(err.Code) == ChannelError
// }
