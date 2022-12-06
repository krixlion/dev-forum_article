package cmd

import (
	"context"

	"github.com/krixlion/dev-forum_article/pkg/entity"
)

func (db DB) Close() error {
	panic("Close not implemented")
}

func (db DB) Create(context.Context, entity.Article) error {
	panic("Create not implemented")
}

func (db DB) Update(context.Context, entity.Article) error {
	panic("Update not implemented")
}
