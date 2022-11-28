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
	PublishOnQueue(context.Context, Event) error
}

type Subscriber interface {
	Subscribe(context.Context, HandlerFunc) error
	SubscribeToQueue(context.Context, HandlerFunc) error
}

// HandleFunc is registered for a subscriber and is invoked for each received event.
type HandlerFunc func(Event) error
