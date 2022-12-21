package storage

import (
	"context"
	"encoding/json"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/logging"
	"github.com/krixlion/dev-forum_article/pkg/storage/cmd"
	"github.com/krixlion/dev-forum_article/pkg/storage/query"
)

// DB is a wrapper for the read model and write model to use with Storage interface.
type DB struct {
	cmd    cmd.DB
	query  query.DB
	logger logging.Logger
}

func NewStorage(cmd cmd.DB, query query.DB, logger logging.Logger) Storage {
	return &DB{
		cmd:    cmd,
		query:  query,
		logger: logger,
	}
}

func (storage DB) Close() error {
	err := storage.cmd.Close()
	storage.query.Close()
	return err
}

func (storage DB) Get(ctx context.Context, id string) (entity.Article, error) {
	return storage.query.Get(ctx, id)
}

func (storage DB) GetMultiple(ctx context.Context, offset, limit string) ([]entity.Article, error) {
	return storage.query.GetMultiple(ctx, offset, limit)
}

func (storage DB) Update(ctx context.Context, article entity.Article) error {
	return storage.cmd.Update(ctx, article)
}

func (storage DB) Create(ctx context.Context, article entity.Article) error {
	return storage.cmd.Create(ctx, article)
}

func (storage DB) Delete(ctx context.Context, id string) error {
	return storage.cmd.Delete(ctx, id)
}

func (db DB) CatchUp(ctx context.Context) error {
	db.cmd.Subscribe(ctx, func(ctx context.Context, e event.Event) (err error) {

		if e.Entity != entity.ArticleEntity {
			return nil
		}

		if err = ctx.Err(); err != nil {
			return
		}

		switch e.Type {
		case event.Created:
			var article entity.Article
			err = json.Unmarshal([]byte(e.Body), &article)
			if err != nil {
				return
			}

			err = db.query.Create(ctx, article)
			return

		case event.Deleted:
			var id string
			err = json.Unmarshal([]byte(e.Body), &id)
			if err != nil {
				return
			}

			err = db.query.Delete(ctx, id)
			return

		case event.Updated:
			var article entity.Article
			err = json.Unmarshal([]byte(e.Body), &article)
			if err != nil {
				return
			}

			err = db.query.Update(ctx, article)
			return
		}

		return nil
	})

	return nil
}
