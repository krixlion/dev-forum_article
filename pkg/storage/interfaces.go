package storage

import (
	"context"

	"github.com/krixlion/dev_forum-article/pkg/entity"
)

type Getter interface {
	Get(ctx context.Context, id string) (entity.Article, error)
	GetMultiple(ctx context.Context, offset, limit string) ([]entity.Article, error)
}

type Writer interface {
	Create(context.Context, entity.Article) error
	Update(context.Context, entity.Article) error
	Delete(ctx context.Context, id string) error
}
