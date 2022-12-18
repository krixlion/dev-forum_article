package event

import (
	"context"
	"io"

	"github.com/krixlion/dev-forum_article/pkg/entity"
)

type Handler interface {
	Consumer
	Publisher
	io.Closer
}

type Consumer interface {
	// Command is queue's name
	Consume(ctx context.Context, command string, entity entity.EntityName, eventType EventType) (<-chan Event, error)
}

type Publisher interface {
	// Exchanges and queues should be maintained internally depending on the type of event.
	Publish(context.Context, Event) error

	// Resilient publish should return only parsing error and on any other error retry each event until it succeeds.
	ResilientPublish(context.Context, Event) error
}

type Subscriber interface {
	// Subscribe takes in a context, entity and event type to subscribe for and a function to invoke on each event.
	Subscribe(context.Context, HandlerFunc, entity.EntityName, ...EventType) error
}

// HandleFunc is registered for a subscriber and is invoked for each received event.
type HandlerFunc func(context.Context, Event) error

// type Command interface {
// 	Name()
// 	Execute() error
// }
