package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/event"
	"github.com/krixlion/dev_forum-article/pkg/logging"
	"github.com/krixlion/dev_forum-article/pkg/tracing"
	"go.opentelemetry.io/otel"
)

// DB is a wrapper for the read model and write model to use with Storage interface.
type DB struct {
	cmd    Eventstore
	query  Storage
	logger logging.Logger
}

func NewCQRStorage(cmd Eventstore, query Storage, logger logging.Logger) CQRStorage {
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
		errMsg = fmt.Sprintf("failed to close storage: %s", err)
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}

func (storage DB) Get(ctx context.Context, id string) (entity.Article, error) {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "storage.Get")
	defer span.End()

	article, err := storage.query.Get(ctx, id)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return entity.Article{}, err
	}
	return article, nil
}

func (storage DB) GetMultiple(ctx context.Context, offset, limit string) ([]entity.Article, error) {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "storage.GetMultiple")
	defer span.End()

	articles, err := storage.query.GetMultiple(ctx, offset, limit)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}
	return articles, nil
}

func (storage DB) Update(ctx context.Context, article entity.Article) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "storage.Update")
	defer span.End()

	if err := storage.cmd.Update(ctx, article); err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}
	return nil
}

func (storage DB) Create(ctx context.Context, article entity.Article) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "storage.Create")
	defer span.End()

	if err := storage.cmd.Create(ctx, article); err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}
	return nil
}

func (storage DB) Delete(ctx context.Context, id string) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "storage.Delete")
	defer span.End()

	if err := storage.cmd.Delete(ctx, id); err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}
	return nil
}

// CatchUp handles events required to keep the read model consistent.
func (db DB) CatchUp(e event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "storage.CatchUp")
	defer span.End()

	switch e.Type {
	case event.ArticleCreated:
		var article entity.Article
		if err := json.Unmarshal(e.Body, &article); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to parse event",
				"err", err,
				"event", e,
			)
			return
		}

		if err := db.query.Create(ctx, article); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to create article",
				"err", err,
				"event", e,
			)
		}
		return

	case event.ArticleDeleted:
		var id string
		if err := json.Unmarshal(e.Body, &id); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to parse event",
				"err", err,
				"event", e,
			)
		}

		if err := db.query.Delete(ctx, id); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to delete article",
				"err", err,
				"event", e,
			)
		}
		return

	case event.ArticleUpdated:
		var article entity.Article
		if err := json.Unmarshal(e.Body, &article); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to parse event",
				"err", err,
				"event", e,
			)
			return
		}

		if err := db.query.Update(ctx, article); err != nil {
			tracing.SetSpanErr(span, err)

			db.logger.Log(ctx, "Failed to update article",
				"err", err,
				"event", e,
			)
		}
		return
	}
}
