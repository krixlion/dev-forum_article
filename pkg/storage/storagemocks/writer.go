package storagemocks

import (
	"context"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/stretchr/testify/mock"
)

type Writer struct {
	*mock.Mock
}

func NewWriter() Writer {
	return Writer{
		Mock: new(mock.Mock),
	}
}

func (m Writer) Create(ctx context.Context, article entity.Article) error {
	args := m.Called(ctx, article)
	return args.Error(0)
}

func (m Writer) Update(ctx context.Context, article entity.Article) error {
	args := m.Called(ctx, article)
	return args.Error(0)
}

func (m Writer) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
