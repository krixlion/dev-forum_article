package query

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v9"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

const articlesPrefix = "articles"

func addArticlesPrefix(key string) string {
	return fmt.Sprintf("%s:%s", articlesPrefix, key)
}

func (db DB) Close() error {
	return db.redis.Close()
}

func (db DB) Get(ctx context.Context, id string) (entity.Article, error) {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "Get")
	defer span.End()

	id = addArticlesPrefix(id)
	article := entity.Article{}

	err := db.redis.HGetAll(ctx, id).Scan(&article)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return entity.Article{}, err
	}

	return article, nil
}

func (db DB) GetMultiple(ctx context.Context, offset, limit string) ([]entity.Article, error) {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "GetMultiple")
	defer span.End()

	off, err := strconv.ParseInt(offset, 10, 0)
	if err != nil {
		return nil, err
	}

	count, err := strconv.ParseInt(limit, 10, 0)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	ids, err := db.redis.Sort(ctx, articlesPrefix, &redis.Sort{
		By:     addArticlesPrefix("*->title"),
		Offset: off,
		Count:  count,
		Alpha:  true,
	}).Result()

	if err != nil {
		return nil, err
	}

	articles := []entity.Article{}
	commands := []*redis.MapStringStringCmd{}
	pipeline := db.redis.Pipeline()

	for _, id := range ids {
		id = addArticlesPrefix(id)
		commands = append(commands, pipeline.HGetAll(ctx, id))
	}

	pipeline.Exec(ctx)

	for _, cmd := range commands {
		article := entity.Article{}
		err := cmd.Scan(&article)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return nil, err
		}
		articles = append(articles, article)

	}
	return articles, nil
}

func (db DB) Create(ctx context.Context, article entity.Article) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "Get")
	defer span.End()

	prefixedId := addArticlesPrefix(article.Id)

	_, err := db.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		values := map[string]interface{}{
			"id":      article.Id,
			"user_id": article.UserId,
			"title":   article.Title,
			"body":    article.Body,
		}
		db.redis.HSet(ctx, prefixedId, values)
		db.redis.SAdd(ctx, articlesPrefix, article.Id)
		return nil
	})

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	return nil
}

func (db DB) Update(ctx context.Context, article entity.Article) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "Update")
	defer span.End()

	return db.Create(ctx, article)
}

func (db DB) Delete(ctx context.Context, id string) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "Delete")
	defer span.End()

	return db.redis.Del(ctx, id).Err()
}
