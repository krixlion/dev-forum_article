package storage_test

import (
	"context"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/stretchr/testify/mock"
)

type mockQuery struct {
	*mock.Mock
}

func (m mockQuery) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m mockQuery) Get(ctx context.Context, id string) (entity.Article, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.Article), args.Error(1)
}

func (m mockQuery) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.Article, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]entity.Article), args.Error(1)
}

func (m mockQuery) Create(ctx context.Context, a entity.Article) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m mockQuery) Update(ctx context.Context, a entity.Article) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m mockQuery) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockCmd struct {
	*mock.Mock
}

func (m mockCmd) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m mockCmd) Consume(ctx context.Context, queue string, eventType event.EventType) (<-chan event.Event, error) {
	args := m.Called(ctx, queue, eventType)
	return args.Get(0).(<-chan event.Event), args.Error(1)
}

func (m mockCmd) Create(ctx context.Context, a entity.Article) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m mockCmd) Update(ctx context.Context, a entity.Article) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m mockCmd) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
