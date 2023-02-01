package mocks

import (
	"context"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/stretchr/testify/mock"
)

type Query struct {
	*mock.Mock
}

func (m Query) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m Query) Get(ctx context.Context, id string) (entity.Article, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.Article), args.Error(1)
}

func (m Query) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.Article, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]entity.Article), args.Error(1)
}

func (m Query) GetBelongingIDs(ctx context.Context, userId string) ([]string, error) {
	panic("not implemented") // TODO: Implement
}
func (m Query) Create(ctx context.Context, a entity.Article) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m Query) Update(ctx context.Context, a entity.Article) error {
	args := m.Called(ctx, a)
	return args.Error(0)
}

func (m Query) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
