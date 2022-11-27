package rabbitmq

import (
	"context"
	"sync"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/krixlion/dev-forum_article/pkg/log"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sony/gobreaker"
)

type RabbitMQ struct {
	MaxFailures       uint32
	ReconnectInterval time.Duration
	mu                sync.RWMutex            // Mutex protecting connection when reconnecting.
	url               string                  // Connection string to RabbitMQ broker.
	errC              chan *amqp.Error        // Channel to watch for errors from broker in order to renew the connection.
	retryC            chan Message            // Queue for messages waiting to be republished.
	readC             chan chan *amqp.Channel // Access channel for accessing the RabbitMQ Channel in a thread-safe way.
	connection        *amqp.Connection
	breaker           *gobreaker.CircuitBreaker
	logger            kitlog.Logger
}

func NewRabbitMQ(url string, queueSize int, maxFailures uint32, reconnectInterval time.Duration, st gobreaker.Settings) *RabbitMQ {
	return &RabbitMQ{
		retryC:            make(chan Message, queueSize),
		readC:             make(chan chan *amqp.Channel),
		logger:            log.MakeLogger(),
		breaker:           gobreaker.NewCircuitBreaker(st),
		url:               url,
		MaxFailures:       maxFailures,
		ReconnectInterval: reconnectInterval,
	}
}

func (mq *RabbitMQ) Run() {
	err := mq.reDial()
	if err != nil {
		mq.logger.Log("err", "Failed to connect to RabbitMQ")
	}

	mq.setUpChannel()
	if err != nil {
		mq.logger.Log("err", "")
	}

	go mq.watchForConnectionErrors()
	go mq.watchForChannelRequests()
	go mq.runRetryQueue()
}

func (mq *RabbitMQ) watchForChannelRequests() error {
	for req := range mq.readC {
		// mq.mu.RLock()
		channel, err := mq.connection.Channel()
		if err != nil {
			return err
		}
		req <- channel
		// mq.mu.RUnlock()
	}
	return nil
}

// WatchConnection is meant to be run in a seperate goroutine.
func (mq *RabbitMQ) watchForConnectionErrors() {
	for e := range mq.errC {
		if e == nil || !isConnectionError(e) {
			continue
		}

		for {
			err := mq.reDial()
			if err == nil {
				break
			}
			mq.logger.Log("msg", "Reconnecting to RabbitMQ")
			time.Sleep(mq.ReconnectInterval)
		}
	}
}
func (mq *RabbitMQ) runRetryQueue() {
	for msg := range mq.retryC {
		for {
			_, err := mq.breaker.Execute(func() (interface{}, error) {
				return nil, mq.Publish(context.Background(), 1*time.Second, msg)
			})
			if err == nil {
				break
			}
		}
	}
}

// ReDial renews the TCP connection.
func (mq *RabbitMQ) reDial() error {
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
		mq.logger.Log("msg", "Closing active connections")
		if err := mq.connection.Close(); err != nil {
			mq.logger.Log("msg", "Failed to close active connections", "err", err.Error())
		}
	}
}
