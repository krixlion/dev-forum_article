package rabbitmq

import (
	"fmt"
	"sync"
	"time"

	"github.com/krixlion/dev-forum_article/pkg/log"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

const (
	ArticleServiceEntity = "article"
	// ArticleServiceExchangeName = "article"
	// ArticleServiceExchangeKind = "topic"
)

type RabbitMQ struct {
	MaxFailures       uint32
	ReconnectInterval time.Duration

	mu         sync.RWMutex            // Mutex protecting connection when reconnecting.
	url        string                  // Connection string to RabbitMQ broker.
	errC       chan *amqp.Error        // Channel to watch for errors from broker in order to renew the connection.
	retryC     chan Message            // Queue for messages waiting to be republished.
	readC      chan chan *amqp.Channel // Access channel for accessing the RabbitMQ Channel in a thread-safe way.
	connection *amqp.Connection
	breaker    *gobreaker.CircuitBreaker
	logger     *zap.SugaredLogger
}

func NewRabbitMQ(url string, queueSize int, maxFailures uint32, reconnectInterval time.Duration, st gobreaker.Settings) *RabbitMQ {
	logger, _ := log.MakeZapLogger()
	return &RabbitMQ{
		retryC:            make(chan Message, queueSize),
		readC:             make(chan chan *amqp.Channel),
		logger:            logger,
		breaker:           gobreaker.NewCircuitBreaker(st),
		url:               url,
		MaxFailures:       maxFailures,
		ReconnectInterval: reconnectInterval,
	}
}

func (mq *RabbitMQ) Run() error {
	return nil
}

// HandleChannelRequests is meant to be run in a seperate goroutine.
func (mq *RabbitMQ) handleChannelRequests() {
	for req := range mq.readC {
		// mq.mu.RLock()
		channel, err := mq.setUpChannel()
		if err != nil {
			mq.logger.Infow("Failed opening a new channel", "err", err)
		}
		// mq.mu.RUnlock()
		req <- channel
	}
}

// WatchConnection is meant to be run in a seperate goroutine.
func (mq *RabbitMQ) handleConnectionErrors() {
	for e := range mq.errC {
		if e == nil || !isConnectionError(e) {
			continue
		}

		for {
			err := mq.dial()
			if err == nil {
				break
			}
			mq.logger.Infow("Reconnecting to RabbitMQ")
			time.Sleep(mq.ReconnectInterval)
		}
	}
}

// Dial renews current TCP connection.
func (mq *RabbitMQ) dial() error {
	conn, err := mq.breaker.Execute(func() (interface{}, error) {
		return amqp.Dial(mq.url)
	})
	if err != nil {
		return err
	}
	mq.mu.Lock()
	defer mq.mu.Unlock()
	mq.connection = conn.(*amqp.Connection)
	return nil
}

func (mq *RabbitMQ) channel() *amqp.Channel {
	ask := make(chan *amqp.Channel)
	mq.readC <- ask
	s := <-ask
	return s
}

// Close active connection gracefully.
func (mq *RabbitMQ) Close() {
	if mq.connection != nil && !mq.connection.IsClosed() {
		mq.logger.Infow("Closing active connections")
		if err := mq.connection.Close(); err != nil {
			mq.logger.Infow("Failed to close active connections", "err", err.Error())
		}
	}
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

func (mq *RabbitMQ) setUpChannel() (*amqp.Channel, error) {
	ch, err := mq.connection.Channel()
	mq.mu.Lock()
	mq.errC = ch.NotifyClose(make(chan *amqp.Error, 16))
	defer mq.mu.Unlock()
	return ch, err
}
