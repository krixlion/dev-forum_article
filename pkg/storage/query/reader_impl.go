package query

import (
	"context"

	"github.com/krixlion/dev-forum_article/pkg/entity"
)

func (db DB) Get(ctx context.Context, id string) (entity.Article, error) {
	panic("Get not implemented")
}

func (db DB) GetMultiple(ctx context.Context, offset, limit string) ([]entity.Article, error) {
	panic("GetMultiple not implemented")
}
