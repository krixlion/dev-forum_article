package storage

import (
	"context"
	"encoding/json"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/log"
	"github.com/krixlion/dev-forum_article/pkg/storage/cmd"
	"github.com/krixlion/dev-forum_article/pkg/storage/query"
)

type DB struct {
	cmd    cmd.DB
	query  query.DB
	logger log.Logger
}

func NewStorage(cmd cmd.DB, query query.DB, logger log.Logger) Storage {
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

		if e.Entity != entity.ArticleEntityName {
			return nil
		}

		if err = ctx.Err(); err != nil {
			return
		}

		switch e.Type {
		case event.ArticleCreated:
			var article entity.Article
			err = json.Unmarshal([]byte(e.Body), &article)
			if err != nil {
				return
			}

			err = db.query.Create(ctx, article)
			return

		case event.ArticleDeleted:
			var id string
			err = json.Unmarshal([]byte(e.Body), &id)
			if err != nil {
				return
			}

			err = db.query.Delete(ctx, id)
			return

		case event.ArticleUpdated:
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
