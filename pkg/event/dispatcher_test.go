package event

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev-forum_article/pkg/helpers/gentest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/sync/errgroup"
)

type mockHandler struct {
	*mock.Mock
}

func (h mockHandler) Handle(e Event) {
	h.Called(e)
}

func containsHandler(handlers []Handler, target Handler) bool {
	for _, handler := range handlers {
		if cmp.Equal(handler, target, cmpopts.IgnoreUnexported(mock.Mock{})) {
			return true
		}
	}
	return false
}

func Test_Subscribe(t *testing.T) {
	testCases := []struct {
		desc    string
		handler Handler
		eTypes  []EventType
	}{
		{
			desc:    "Check if simple handler is subscribed succesfully",
			handler: &mockHandler{},
			eTypes:  []EventType{ArticleCreated, ArticleDeleted, UserCreated},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			dispatcher := MakeDispatcher(10)
			dispatcher.Subscribe(tC.handler, tC.eTypes...)

			for _, eType := range tC.eTypes {
				if !containsHandler(dispatcher.handlers[eType], tC.handler) {
					t.Errorf("Handler was not registered succesfully")
				}
			}
		})
	}
}

func Test_mergeChans(t *testing.T) {
	testCases := []struct {
		desc string
		want []Event
	}{
		{
			desc: "Test if receives all events from multiple channels",
			want: []Event{
				{
					AggregateId: gentest.RandomString(5),
				},
				{
					AggregateId: gentest.RandomString(5),
					Type:        ArticleDeleted,
				},
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

			chans := func() (chans []<-chan Event) {
				for _, e := range tC.want {
					v := make(chan Event, 1)
					v <- e
					chans = append(chans, v)
				}
				return
			}()

			out := mergeChans(chans...)
			var got []Event
			for i := 0; i < len(tC.want); i++ {
				got = append(got, <-out)
			}

			if !assert.ElementsMatch(t, got, tC.want) {
				t.Errorf("Events are not equal:\n got = %+v\n want = %+v\n", got, tC.want)
				return
			}
		})
	}
}

func Test_Dispatch(t *testing.T) {
	testCases := []struct {
		desc    string
		arg     Event
		handler mockHandler
	}{
		{
			desc: "Test if handler is called on simple random event",
			arg: Event{
				Type:        ArticleCreated,
				AggregateId: "article",
			},
			handler: func() mockHandler {
				m := mockHandler{new(mock.Mock)}
				m.On("Handle", mock.AnythingOfType("Event")).Return().Times(1)
				return m
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			d := MakeDispatcher(2)
			d.Subscribe(tC.handler, tC.arg.Type)
			d.Dispatch(tC.arg)

			// Wait for the handler to get invoked in a seperate goroutine.
			time.Sleep(time.Millisecond * 5)

			tC.handler.AssertCalled(t, "Handle", tC.arg)
			tC.handler.AssertNumberOfCalls(t, "Handle", 1)
		})
	}
}

func Test_Run(t *testing.T) {
	t.Run("Test if Run() returns on context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		errg, ctx := errgroup.WithContext(ctx)

		d := MakeDispatcher(20)
		errg.Go(func() error {
			d.Run(ctx)
			return ctx.Err()
		})

		before := time.Now()
		cancel()
		err := errg.Wait()
		after := time.Now()
		stopTime := after.Sub(before)

		// If time needed for Run to return was longer than a millisecond or unexpected error was returned.
		if !errors.Is(err, context.Canceled) || stopTime > time.Millisecond {
			t.Errorf("Run did not stop on context cancellation\n Time needed for func to return: %v", stopTime.Seconds())
			return
		}
	})
}
