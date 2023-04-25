package storagemocks

import (
	"context"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/stretchr/testify/mock"
)

type Storage struct {
	*mock.Mock
}

func NewStorage() Storage {
	return Storage{
		Mock: new(mock.Mock),
	}
}

func (m Storage) Get(ctx context.Context, id string) (entity.Article, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.Article), args.Error(1)
}

func (m Storage) GetBelongingIDs(ctx context.Context, userId string) ([]string, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]string), args.Error(1)
}

func (m Storage) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.Article, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]entity.Article), args.Error(1)
}

func (m Storage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m Storage) Create(ctx context.Context, article entity.Article) error {
	args := m.Called(ctx, article)
	return args.Error(0)
}

func (m Storage) Update(ctx context.Context, article entity.Article) error {
	args := m.Called(ctx, article)
	return args.Error(0)
}

func (m Storage) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
