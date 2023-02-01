package mocks

import (
	"context"

	"github.com/krixlion/dev_forum-article/pkg/event"
	"github.com/stretchr/testify/mock"
)

type Broker struct {
	*mock.Mock
}

func (m Broker) Consume(ctx context.Context, queue string, eventType event.EventType) (<-chan event.Event, error) {
	args := m.Called(ctx, queue, eventType)
	return args.Get(0).(<-chan event.Event), args.Error(1)
}

func (m Broker) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m Broker) Publish(ctx context.Context, e event.Event) error {
	args := m.Called(ctx, e)
	return args.Error(0)
}

func (m Broker) ResilientPublish(e event.Event) error {
	args := m.Called(e)
	return args.Error(0)
}
