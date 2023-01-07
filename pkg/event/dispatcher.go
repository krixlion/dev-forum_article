package event

import (
	"context"
	"sync"
)

type Dispatcher struct {
	handlers map[EventType][]Handler
	events   <-chan Event
}

func MakeDispatcher() Dispatcher {
	return Dispatcher{
		handlers: make(map[EventType][]Handler),
	}
}

// AddEventSources registers provided channels as an event source.
// This method is not thread safe and should be called before Run().
func (d *Dispatcher) AddEventSources(sources ...<-chan Event) {
	out := mergeChans(sources...)
	d.events = out
}

func (d *Dispatcher) Run(ctx context.Context) {
	for {
		select {
		case event := <-d.events:
			d.Dispatch(event)
		case <-ctx.Done():
			return
		}
	}
}

func (d *Dispatcher) Subscribe(handler Handler, eTypes ...EventType) {
	for _, eType := range eTypes {
		d.handlers[eType] = append(d.handlers[eType], handler)
	}
}

func (d Dispatcher) Dispatch(e Event) {
	for _, handler := range d.handlers[e.Type] {
		go handler.Handle(e)
	}
}

func mergeChans(cs ...<-chan Event) <-chan Event {
	out := make(chan Event)

	wg := sync.WaitGroup{}
	wg.Add(len(cs))

	for _, c := range cs {
		go func(c <-chan Event) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}

	return out
}
