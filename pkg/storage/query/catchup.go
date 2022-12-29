package query

import (
	"context"
	"encoding/json"
	"time"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/tracing"
	"go.opentelemetry.io/otel"
)

// CatchUp handles events required to keep the read model consistent.
func (db DB) CatchUp(e event.Event) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "redis.CatchUp")
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

		if err := db.Create(ctx, article); err != nil {
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

		if err := db.Delete(ctx, id); err != nil {
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

		if err := db.Update(ctx, article); err != nil {
			tracing.SetSpanErr(span, err)

			db.logger.Log(ctx, "Failed to update article",
				"err", err,
				"event", e,
			)
		}
		return
	}
}
