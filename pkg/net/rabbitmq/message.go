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

type Message struct {
	Body        []byte
	ContentType ContentType
	Timestamp   time.Time
	Route
}

type Route struct {
	ExchangeName string
	ExchangeType string
	RoutingKey   string
}

// makeMessageFromEvent returns a message suitable for pub/sub methods and
// a non-nil error if the event could not be marshaled into JSON.
func makeMessageFromEvent(e event.Event) Message {
	// Marshal never returns a non-nil error when input is []byte.
	body, _ := json.Marshal(e)
	return Message{
		Body:        body,
		ContentType: ContentTypeJson,
		Route:       makeRouteFromEvent(e),
		Timestamp:   e.Timestamp,
	}
}

func routingKeyFromEvent(e event.Event) (rKey string) {
	switch e.Type {
	case event.Created:
		rKey = string(e.Entity) + ".event.created"
	case event.Deleted:
		rKey = string(e.Entity) + ".event.deleted"
	case event.Updated:
		rKey = string(e.Entity) + ".event.updated"
	}
	return
}

func makeRouteFromEvent(e event.Event) Route {
	rKey := routingKeyFromEvent(e)

	return Route{
		ExchangeName: string(e.Entity),
		ExchangeType: amqp.ExchangeTopic,
		RoutingKey:   rKey,
	}
}
