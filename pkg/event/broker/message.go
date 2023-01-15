package broker

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/krixlion/dev-forum_article/pkg/event"
	rabbitmq "github.com/krixlion/dev-forum_rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// messageFromEvent returns a message suitable for pub/sub methods and
// a non-nil error if the event could not be marshaled into JSON.
func messageFromEvent(e event.Event) rabbitmq.Message {
	body, err := json.Marshal(e)
	if err != nil {
		panic(fmt.Sprintf("Invalid JSON tags on event.Event type, err: %v", err))
	}

	return rabbitmq.Message{
		Body:        body,
		ContentType: rabbitmq.ContentTypeJson,
		Route:       routeFromEvent(e.Type),
		Timestamp:   e.Timestamp,
	}
}

func routeFromEvent(eType event.EventType) rabbitmq.Route {
	v := strings.Split(string(eType), "-")

	return rabbitmq.Route{
		ExchangeName: v[0],
		ExchangeType: amqp.ExchangeTopic,
		RoutingKey:   fmt.Sprintf("%s.event.%s", v[0], v[1]),
	}
}
