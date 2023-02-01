package mocks

import (
	"context"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/event"
	"github.com/stretchr/testify/mock"
)

type Cmd struct {
	*mock.Mock
}

func (m Cmd) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m Cmd) Consume(ctx context.Context, queue string, eventType event.EventType) (<-chan event.Event, error) {
	args := m.Called(ctx, queue, eventType)
	return args.Get(0).(<-chan event.Event), args.Error(1)
}

func (m Cmd) Create(ctx context.Context, a entity.Article) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m Cmd) Update(ctx context.Context, a entity.Article) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m Cmd) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
