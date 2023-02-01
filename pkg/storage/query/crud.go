package query

import (
	"context"
	"errors"
	"strconv"

	"github.com/go-redis/redis/v9"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/tracing"
	"go.opentelemetry.io/otel"
)

func (db DB) Close() error {
	return db.redis.Close()
}

func (db DB) Get(ctx context.Context, id string) (entity.Article, error) {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "redis.Get")
	defer span.End()

	id = addPrefix(articlesPrefix, id)
	var data redisArticle

	cmd := db.redis.HGetAll(ctx, id)
	err := scan(cmd, &data)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return entity.Article{}, err
	}

	v, err := data.Article()
	if err != nil {
		tracing.SetSpanErr(span, err)
		return entity.Article{}, err
	}
	return v, nil
}

func (db DB) GetMultiple(ctx context.Context, offset, limit string) (articles []entity.Article, err error) {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "redis.GetMultiple")
	defer span.End()

	var off int64
	if offset != "" {
		off, err = strconv.ParseInt(offset, 10, 0)
		if err != nil {
			return nil, err
		}
	}

	var count int64
	if limit != "" {
		count, err = strconv.ParseInt(limit, 10, 0)
		if err != nil {
			tracing.SetSpanErr(span, err)
			return nil, err
		}
	}

	ids, err := db.redis.Sort(ctx, articlesPrefix, &redis.Sort{
		By:     addPrefix(articlesPrefix, "*->title"),
		Offset: off,
		Count:  count,
		Alpha:  true,
		Order:  "DESC",
	}).Result()

	if err != nil {
		return nil, err
	}

	commands := []*redis.MapStringStringCmd{}
	pipeline := db.redis.Pipeline()

	for _, id := range ids {
		id = addPrefix(articlesPrefix, id)
		commands = append(commands, pipeline.HGetAll(ctx, id))
	}

	pipeline.Exec(ctx)

	for _, cmd := range commands {
		var data redisArticle
		err := scan(cmd, &data)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			tracing.SetSpanErr(span, err)
			return nil, err
		}

		article, err := data.Article()
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func (db DB) GetBelongingIDs(ctx context.Context, userId string) ([]string, error) {
	ids, err := db.redis.SMembers(ctx, addPrefix(usersPrefix, userId)).Result()
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (db DB) Create(ctx context.Context, article entity.Article) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "redis.Create")
	defer span.End()

	prefixedArticleId := addPrefix(articlesPrefix, article.Id)

	_, err := db.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		values := mapArticle(article)
		db.redis.HSet(ctx, prefixedArticleId, values)
		db.redis.SAdd(ctx, articlesPrefix, prefixedArticleId)
		db.redis.SAdd(ctx, addPrefix(usersPrefix, article.UserId), article.Id)
		return nil
	})
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	return nil
}

func (db DB) Update(ctx context.Context, article entity.Article) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "redis.Update")
	defer span.End()

	prefixedId := addPrefix(articlesPrefix, article.Id)
	numOfKeys, err := db.redis.Exists(ctx, prefixedId).Result()
	if err != nil {
		return err
	}

	if numOfKeys <= 0 {
		return redis.Nil
	}

	_, err = db.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		values := mapArticle(article)
		db.redis.HSet(ctx, prefixedId, values)
		return nil
	})
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}
	return nil
}

func (db DB) Delete(ctx context.Context, id string) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "redis.Delete")
	defer span.End()

	_, err := db.redis.TxPipelined(ctx, func(p redis.Pipeliner) error {
		id = addPrefix(articlesPrefix, id)
		db.redis.Del(ctx, id).Result()
		return nil
	})
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	return nil
}
