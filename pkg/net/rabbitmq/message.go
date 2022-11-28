package rabbitmq

import (
	"encoding/json"
	"time"

	"github.com/krixlion/dev-forum_article/pkg/event"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ContentType string

const (
	ContentTypeJson ContentType = "application/json"
	ContentTypeText ContentType = "text/plain"
)

type Message struct {
	Body        []byte
	ContentType ContentType
	Timestamp   time.Time
	// OnQueue     bool // Whether a Message should be published directly to a queue.
	// Enqueue     bool // Whether a Message should be enqueued on failure for later retry.
	Route
}

type Route struct {
	ExchangeName string
	ExchangeType string
	RoutingKey   string
	QueueName    string
}

// MessageFromEvent returns a message suitable for pub/sub methods and
// a non-nil error if the event could not be marshaled into JSON.
func messageFromEvent(e event.Event) (Message, error) {
	body, err := json.Marshal(e.Body)
	if err != nil {
		return Message{}, err
	}

	rKey := routingKeyFromEvent(e)
	return Message{
		Body:        body,
		ContentType: ContentTypeJson,
		Route: Route{
			ExchangeName: e.Entity,
			ExchangeType: amqp.ExchangeTopic,
			RoutingKey:   rKey,
		},
	}, nil
}

func routingKeyFromEvent(e event.Event) (rKey string) {
	switch e.Type {
	case event.Created:
		rKey = e.Entity + createdEvent
	case event.Deleted:
		rKey = e.Entity + deletedEvent
	}
	return
}
