package mocks

import (
	"context"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/event"
	"github.com/stretchr/testify/mock"
)

type CQRStorage struct {
	*mock.Mock
}

func (m CQRStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m CQRStorage) Get(ctx context.Context, id string) (entity.Article, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.Article), args.Error(1)
}

func (m CQRStorage) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.Article, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]entity.Article), args.Error(1)
}

func (m CQRStorage) GetBelongingIDs(ctx context.Context, userId string) ([]string, error) {
	panic("not implemented") // TODO: Implement
}

func (m CQRStorage) Create(ctx context.Context, a entity.Article) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m CQRStorage) Update(ctx context.Context, a entity.Article) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m CQRStorage) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m CQRStorage) CatchUp(e event.Event) {
	m.Called(e)
}
