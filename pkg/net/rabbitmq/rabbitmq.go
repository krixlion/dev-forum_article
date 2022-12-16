package rabbitmq

import (
	"context"
	"sync"
	"time"

	"github.com/krixlion/dev-forum_article/pkg/log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/sony/gobreaker"
	"golang.org/x/sync/errgroup"
)

const (
// ArticleServiceExchangeName = "article"
// ArticleServiceExchangeKind = "topic"
)

type RabbitMQ struct {
	MaxFailures       uint32
	ReconnectInterval time.Duration

	connection      *amqp.Connection
	mutex           sync.RWMutex            // Mutex protecting connection during reconnecting.
	url             string                  // Connection string to RabbitMQ broker.
	notifyConnClose chan *amqp.Error        // Channel to watch for errors from the broker in order to renew the connection.
	publishQueue    chan Message            // Queue for messages waiting to be republished.
	readChannel     chan chan *amqp.Channel // Access channel for accessing the RabbitMQ Channel in a thread-safe way.
	breaker         *gobreaker.TwoStepCircuitBreaker
	breakerSettings gobreaker.Settings
	logger          log.Logger
}

func NewRabbitMQ(url string, queueSize int, maxFailures uint32, reconnectInterval time.Duration, st gobreaker.Settings) *RabbitMQ {
	logger, _ := log.NewLogger()
	return &RabbitMQ{
		publishQueue:      make(chan Message, queueSize),
		readChannel:       make(chan chan *amqp.Channel),
		notifyConnClose:   make(chan *amqp.Error, 16),
		logger:            logger,
		breaker:           gobreaker.NewTwoStepCircuitBreaker(st),
		breakerSettings:   st,
		url:               url,
		MaxFailures:       maxFailures,
		ReconnectInterval: reconnectInterval,
	}
}

func (mq *RabbitMQ) Run(ctx context.Context) error {
	if err := mq.dial(); err != nil {
		return err
	}

	errg, ctx := errgroup.WithContext(ctx)
	errg.Go(func() error {
		return mq.runPublishQueue(ctx)
	})

	go mq.handleConnectionErrors()
	go mq.handleChannelReads()

	return errg.Wait()
}

func (mq *RabbitMQ) runPublishQueue(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		default:
			preparedMessages := mq.prepareExchangePipelined(ctx, mq.publishQueue)
			mq.publishPipelined(ctx, preparedMessages)
		}
	}
}

// handleChannelReads is meant to be run in a seperate goroutine.
func (mq *RabbitMQ) handleChannelReads() {
	for req := range mq.readChannel {
		channel, err := mq.setUpChannel()
		if err != nil {
			mq.logger.Log("Failed opening a new channel", "err", err)
			close(req)
			continue
		}
		req <- channel
	}
}

// handleConnectionErrors is meant to be run in a seperate goroutine.
func (mq *RabbitMQ) handleConnectionErrors() {
	for e := range mq.notifyConnClose {
		if e == nil {
			continue
		}

		for {
			err := mq.dial()
			if err == nil {
				break
			}
			mq.logger.Log("Reconnecting to RabbitMQ")
			time.Sleep(mq.ReconnectInterval)
		}
	}
}

// dial renews current TCP connection.
func (mq *RabbitMQ) dial() error {
	callSucceded, err := mq.breaker.Allow()
	if err != nil {
		return err
	}
	conn, err := amqp.Dial(mq.url)
	if err != nil {
		if isConnectionError(err.(*amqp.Error)) {
			callSucceded(false)
		}
		// Error did not render broker unavailable.
		callSucceded(true)

		return err
	}
	callSucceded(true)

	mq.mutex.Lock()
	defer mq.mutex.Unlock()
	mq.connection = conn
	return nil
}

func (mq *RabbitMQ) channel() *amqp.Channel {
	for {
		ask := make(chan *amqp.Channel)
		mq.readChannel <- ask

		// For range over channel terminates automatically when the channel closes.
		for channel := range ask {
			return channel
		}

		time.Sleep(mq.ReconnectInterval)
	}
}

// Close active connection gracefully.
func (mq *RabbitMQ) Close() {
	if mq.connection != nil && !mq.connection.IsClosed() {
		mq.logger.Log("Closing active connections")
		if err := mq.connection.Close(); err != nil {
			mq.logger.Log("Failed to close active connections", "err", err.Error())
		}
	}
}

func (mq *RabbitMQ) setUpChannel() (*amqp.Channel, error) {
	ch, err := mq.connection.Channel()
	mq.mutex.Lock()
	defer mq.mutex.Unlock()
	mq.connection.NotifyClose(mq.notifyConnClose)
	return ch, err
}
