package event

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type nullHandler struct{}

func (nullHandler) Handle(Event) {}

func containsHandler(handlers []Handler, target Handler) bool {
	for _, handler := range handlers {
		if cmp.Equal(handler, target) {
			return true
		}
	}
	return false
}

func TestSubscribe(t *testing.T) {
	testCases := []struct {
		desc    string
		handler Handler
		eTypes  []EventType
	}{
		{
			desc:    "Check if simple handler is subscribed succesfully",
			handler: nullHandler{},
			eTypes:  []EventType{ArticleCreated, ArticleDeleted, UserCreated},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			dispatcher := MakeDispatcher()
			dispatcher.Subscribe(tC.handler, tC.eTypes...)

			for _, eType := range tC.eTypes {
				if !containsHandler(dispatcher.handlers[eType], tC.handler) {
					t.Errorf("Handler was not registered succesfully")
				}
			}
		})
	}
}
