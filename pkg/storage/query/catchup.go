package query

import (
	"context"
	"encoding/json"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/tracing"
	"go.opentelemetry.io/otel"
)

func (db DB) CatchUp(ctx context.Context, e event.Event) (err error) {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "redis.CatchUp")
	defer span.End()

	if e.Entity != entity.ArticleEntity {
		return nil
	}

	if err = ctx.Err(); err != nil {
		tracing.SetSpanErr(span, err)
		return
	}

	switch e.Type {
	case event.Created:
		var article entity.Article
		err = json.Unmarshal(e.Body, &article)
		if err != nil {
			tracing.SetSpanErr(span, err)
			return
		}

		err = db.Create(ctx, article)
		if err != nil {
			tracing.SetSpanErr(span, err)
			return
		}
		return

	case event.Deleted:
		var id string
		err = json.Unmarshal(e.Body, &id)
		if err != nil {
			tracing.SetSpanErr(span, err)
			return
		}

		err = db.Delete(ctx, id)
		if err != nil {
			tracing.SetSpanErr(span, err)
			return
		}
		return

	case event.Updated:
		var article entity.Article
		err = json.Unmarshal(e.Body, &article)
		if err != nil {
			tracing.SetSpanErr(span, err)
			return
		}

		err = db.Update(ctx, article)
		if err != nil {
			tracing.SetSpanErr(span, err)
		}
		return
	}

	return nil
}
