package event

import (
	"context"
	"io"
)

type Handler interface {
	Consumer
	Publisher
	Subscriber
	io.Closer
}

type Consumer interface {
	Consume(context.Context) (<-chan Event, error)
	ConsumeOnQueue(context.Context) (<-chan Event, error)
}

type Publisher interface {
	// Exchanges and queues should be maintained internally depending on the type of event.
	Publish(context.Context, Event) error

	// Resilient publish should return only parsing error and on any other error retry each event until it succeeds.
	ResilientPublish(context.Context, Event) error
}

type Subscriber interface {
	// Subscribe takes in a context, function which will handle the event and event types to subscribe for.
	Subscribe(context.Context, HandlerFunc) error
}

// HandleFunc is registered for a subscriber and is invoked for each received event.
type HandlerFunc func(context.Context, Event) error
