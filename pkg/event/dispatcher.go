package event

type Dispatch struct {
	handlers map[EventType][]Handler
	broker   Publisher
}

func NewDispatcher(broker Publisher) Dispatcher {
	d := &Dispatch{
		handlers: make(map[EventType][]Handler),
		broker:   broker,
	}
	return d
}

func (d *Dispatch) Subscribe(handler Handler, eTypes ...EventType) {
	for _, eType := range eTypes {
		d.handlers[eType] = append(d.handlers[eType], handler)
	}
}

func (d Dispatch) Dispatch(e Event) error {
	err := d.broker.ResilientPublish(e)
	if err != nil {
		return err
	}

	go func() {
		for _, handler := range d.handlers[e.Type] {
			go handler.Handle(e)
		}
	}()

	return nil
}

func (d Dispatch) Close() error {
	return d.broker.Close()
}
