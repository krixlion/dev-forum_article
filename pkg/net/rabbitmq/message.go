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
	QueueName    string
}

// makeMessageFromEvent returns a message suitable for pub/sub methods and
// a non-nil error if the event could not be marshaled into JSON.
func makeMessageFromEvent(e event.Event) (Message, error) {
	body, err := json.Marshal(e)
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
	case event.ArticleCreated:
		rKey = entity + ".event.created"
	case event.ArticleDeleted:
		rKey = entity + ".event.deleted"
	case event.ArticleUpdated:
		rKey = entity + ".event.updated"
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
