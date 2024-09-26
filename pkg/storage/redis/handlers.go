package redis

import (
	"context"
	"encoding/json"
	"time"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/tracing"
)

func (db Redis) EventHandlers() map[event.EventType][]event.Handler {
	return map[event.EventType][]event.Handler{
		event.ArticleCreated: {
			db.createArticle(),
		},
		event.ArticleUpdated: {
			db.updateArticle(),
		},
		event.ArticleDeleted: {
			db.deleteArticle(),
		},
		event.UserDeleted: {
			db.deleteAllArticlesBelongingToUser(),
		},
	}
}

func (db Redis) deleteAllArticlesBelongingToUser() event.HandlerFunc {
	return func(e event.Event) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		ctx, span := db.tracer.Start(tracing.InjectMetadataIntoContext(ctx, e.Metadata), "deleteAllArticlesBelongingToUser")
		defer span.End()

		var userId string
		if err := json.Unmarshal(e.Body, &userId); err != nil {
			db.logger.Log(ctx, "Failed to unmarshal event body", "err", err, "event", e)
			tracing.SetSpanErr(span, err)
			return
		}

		ids, err := db.client.SMembers(ctx, addPrefix(usersPrefix, userId)).Result()
		if err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to fetch user ids", "err", err, "event", e)
			return
		}

		if _, err := db.client.Del(ctx, ids...).Result(); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to delete articles", "err", err, "event", e)
			return
		}

	}
}

func (db Redis) createArticle() event.HandlerFunc {
	return func(e event.Event) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		ctx, span := db.tracer.Start(tracing.InjectMetadataIntoContext(ctx, e.Metadata), "query.createArticle")
		defer span.End()

		var article entity.Article
		if err := json.Unmarshal(e.Body, &article); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to parse event", "err", err, "event", e)
			return
		}

		if err := db.Create(ctx, article); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to create article", "err", err, "event", e)
		}
	}
}

func (db Redis) updateArticle() event.HandlerFunc {
	return func(e event.Event) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		ctx, span := db.tracer.Start(tracing.InjectMetadataIntoContext(ctx, e.Metadata), "query.updateArticle")
		defer span.End()

		var article entity.Article
		if err := json.Unmarshal(e.Body, &article); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to parse event", "err", err, "event", e)
			return
		}

		if err := db.Update(ctx, article); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to create article", "err", err, "event", e)
		}
	}
}

func (db Redis) deleteArticle() event.HandlerFunc {
	return func(e event.Event) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		ctx, span := db.tracer.Start(tracing.InjectMetadataIntoContext(ctx, e.Metadata), "query.deleteArticle")
		defer span.End()

		var id string
		if err := json.Unmarshal(e.Body, &id); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to parse event", "err", err, "event", e)
			return
		}

		if err := db.Delete(ctx, id); err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "Failed to create article", "err", err, "event", e)
		}
	}
}
