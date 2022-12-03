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

const DefaultConsumer = "article-service"

const (
	createdEvent = ".event.created"
	deletedEvent = ".event.deleted"
	updatedEvent = ".event.updated"
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

// makeMessageFromEvent returns a message suitable for pub/sub methods and
// a non-nil error if the event could not be marshaled into JSON.
func makeMessageFromEvent(e event.Event) (Message, error) {
	body, err := json.Marshal(e.Body)
	if err != nil {
		return Message{}, err
	}

	msg := Message{
		Body:        body,
		ContentType: ContentTypeJson,
		Route:       makeRouteFromEvent(e),
	}

	return msg, nil
}

func routingKeyFromEvent(entity string, t event.EventType) (rKey string) {
	switch t {
	case event.Created:
		rKey = entity + createdEvent
	case event.Deleted:
		rKey = entity + deletedEvent
	case event.Updated:
		rKey = entity + updatedEvent
	}
	return
}

func makeRouteFromEvent(e event.Event) Route {
	rKey := routingKeyFromEvent(e.Entity, e.Type)

	r := Route{
		ExchangeName: e.Entity,
		ExchangeType: amqp.ExchangeTopic,
		RoutingKey:   rKey,
	}

	return r
}
