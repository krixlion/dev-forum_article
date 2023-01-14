package event

import (
	"context"
	"sync"
)

type Dispatcher struct {
	maxWorkers int
	handlers   map[EventType][]Handler
	events     <-chan Event
	mu         sync.Mutex
}

func MakeDispatcher(maxWorkers int) Dispatcher {
	return Dispatcher{
		maxWorkers: maxWorkers,
		handlers:   make(map[EventType][]Handler),
	}
}

// AddEventSources registers provided channels as an event source.
// This method is not thread safe and should be called before Run().
func (d *Dispatcher) AddEventSources(sources ...<-chan Event) {
	d.events = mergeChans(sources...)
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
	d.mu.Lock()
	for _, eType := range eTypes {
		d.handlers[eType] = append(d.handlers[eType], handler)
	}
	d.mu.Unlock()
}

func (d *Dispatcher) Dispatch(e Event) {
	limit := make(chan struct{}, d.maxWorkers)
	for _, handler := range d.handlers[e.Type] {
		limit <- struct{}{}
		go func(handler Handler) {
			handler.Handle(e)
			<-limit
		}(handler)
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
