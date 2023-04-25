package storagemocks

import (
	"context"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/stretchr/testify/mock"
)

type Eventstore struct {
	*mock.Mock
}

func NewEventstore() Eventstore {
	return Eventstore{
		Mock: new(mock.Mock),
	}
}

func (m Eventstore) Consume(ctx context.Context, queue string, eventType event.EventType) (<-chan event.Event, error) {
	args := m.Called(ctx, queue, eventType)
	return args.Get(0).(<-chan event.Event), args.Error(1)
}

func (m Eventstore) Get(ctx context.Context, id string) (entity.Article, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.Article), args.Error(1)
}

func (m Eventstore) GetBelongingIDs(ctx context.Context, userId string) ([]string, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]string), args.Error(1)
}

func (m Eventstore) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.Article, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]entity.Article), args.Error(1)
}

func (m Eventstore) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m Eventstore) Create(ctx context.Context, article entity.Article) error {
	args := m.Called(ctx, article)
	return args.Error(0)
}

func (m Eventstore) Update(ctx context.Context, article entity.Article) error {
	args := m.Called(ctx, article)
	return args.Error(0)
}

func (m Eventstore) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
