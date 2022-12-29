package broker

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/net/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

// MessageFromEvent returns a message suitable for pub/sub methods and
// a non-nil error if the event could not be marshaled into JSON.
func MessageFromEvent(e event.Event) rabbitmq.Message {
	body, err := json.Marshal(e)
	if err != nil {
		panic(fmt.Sprintf("Invalid JSON tags on event.Event type, err: %v", err))
	}

	return rabbitmq.Message{
		Body:        body,
		ContentType: rabbitmq.ContentTypeJson,
		Route:       RouteFromEvent(e.Type),
		Timestamp:   e.Timestamp,
	}
}

func RouteFromEvent(eType event.EventType) rabbitmq.Route {
	v := strings.Split(string(eType), "-")

	return rabbitmq.Route{
		ExchangeName: v[0],
		ExchangeType: amqp.ExchangeTopic,
		RoutingKey:   fmt.Sprintf("%s.event.%s", v[0], v[1]),
	}
}
