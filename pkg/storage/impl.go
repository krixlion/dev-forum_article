package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/logging"
)

// DB is a wrapper for the read model and write model to use with Storage interface.
type DB struct {
	cmd    Eventstore
	query  ReadStorage
	logger logging.Logger
}

func NewStorage(cmd Eventstore, query ReadStorage, logger logging.Logger) Storage {
	return &DB{
		cmd:    cmd,
		query:  query,
		logger: logger,
	}
}

func (storage DB) Close() error {
	var errMsg string

	if err := storage.cmd.Close(); err != nil {
		errMsg = fmt.Sprintf("%s, failed to close eventStore: %s", errMsg, err)
	}

	if err := storage.query.Close(); err != nil {
		errMsg = fmt.Sprintf("failed to close readStorage: %s", err)
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
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

// CatchUp handles events required to keep the read model consistent.
func (db DB) CatchUp(e event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "redis.CatchUp")
	// defer span.End()

	switch e.Type {
	case event.ArticleCreated:
		var article entity.Article
		if err := json.Unmarshal(e.Body, &article); err != nil {
			// // tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to parse event",
				"err", err,
				"event", e,
			)
			return
		}

		if err := db.query.Create(ctx, article); err != nil {
			// // tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to create article",
				"err", err,
				"event", e,
			)
		}
		return

	case event.ArticleDeleted:
		var id string
		if err := json.Unmarshal(e.Body, &id); err != nil {
			// // tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to parse event",
				"err", err,
				"event", e,
			)
		}

		if err := db.query.Delete(ctx, id); err != nil {
			// // tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to delete article",
				"err", err,
				"event", e,
			)
		}
		return

	case event.ArticleUpdated:
		var article entity.Article
		if err := json.Unmarshal(e.Body, &article); err != nil {
			// // tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to parse event",
				"err", err,
				"event", e,
			)
			return
		}

		if err := db.query.Update(ctx, article); err != nil {
			// // tracing.SetSpanErr(span, err)

			db.logger.Log(ctx, "Failed to update article",
				"err", err,
				"event", e,
			)
		}
		return
	}
}
