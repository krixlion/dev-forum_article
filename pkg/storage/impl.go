package storage

import (
	"context"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/logging"
)

// DB is a wrapper for the read model and write model to use with Storage interface.
type DB struct {
	cmd    Eventstore
	query  Querer
	logger logging.Logger
}

func NewStorage(cmd Eventstore, query Querer, logger logging.Logger) Storage {
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

func (db DB) ListenAndCatchUp(ctx context.Context) error {
	types := []event.EventType{
		event.Created,
		event.Deleted,
		event.Updated,
	}
	return db.cmd.Subscribe(ctx, event.HandlerFunc(db.query.CatchUp), entity.ArticleEntity, types...)
}
