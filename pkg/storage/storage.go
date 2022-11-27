package storage

import (
	"context"

	"github.com/krixlion/dev-forum_article/pkg/entity"
)

type Reader interface {
	Get(ctx context.Context, id string) (entity.Article, error)
	GetMultiple(ctx context.Context, offset, limit string) ([]entity.Article, error)
}

type Writer interface {
	Create(context.Context, entity.Article) error
	Update(context.Context, entity.Article) error
}
