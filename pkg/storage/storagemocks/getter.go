package storagemocks

import (
	"context"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/stretchr/testify/mock"
)

type Getter struct {
	*mock.Mock
}

func NewGetter() Getter {
	return Getter{
		Mock: new(mock.Mock),
	}
}

func (m Getter) Get(ctx context.Context, id string) (entity.Article, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.Article), args.Error(1)
}

func (m Getter) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.Article, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]entity.Article), args.Error(1)
}
