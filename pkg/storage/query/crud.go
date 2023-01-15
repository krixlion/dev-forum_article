package query

import (
	"context"
	"errors"
	"strconv"

	"github.com/go-redis/redis/v9"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/tracing"
	"go.opentelemetry.io/otel"
)

const articlesPrefix = "articles"

func (db DB) Close() error {
	return db.redis.Close()
}

func (db DB) Get(ctx context.Context, id string) (entity.Article, error) {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "redis.Get")
	defer span.End()

	id = addArticlesPrefix(id)
	article := entity.Article{}

	cmd := db.redis.HGetAll(ctx, id)
	err := scan(cmd, &article)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return entity.Article{}, err
	}

	return article, nil
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
		By:     addArticlesPrefix("*->title"),
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
		id = addArticlesPrefix(id)
		commands = append(commands, pipeline.HGetAll(ctx, id))
	}

	pipeline.Exec(ctx)

	for _, cmd := range commands {
		article := entity.Article{}
		err := scan(cmd, &article)
		if err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
			tracing.SetSpanErr(span, err)
			return nil, err
		}
		articles = append(articles, article)
	}
	return articles, nil
}

func (db DB) Create(ctx context.Context, article entity.Article) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "redis.Create")
	defer span.End()

	err := db.create(ctx, article)
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	return nil
}

func (db DB) Update(ctx context.Context, article entity.Article) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "redis.Update")
	defer span.End()

	numOfKeys, err := db.redis.Exists(ctx, addArticlesPrefix(article.Id)).Result()
	if err != nil {
		return err
	}

	if numOfKeys <= 0 {
		return redis.Nil
	}

	err = db.create(ctx, article)
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
		id = addArticlesPrefix(id)
		db.redis.Del(ctx, id).Result()
		return nil
	})

	return err
}

func (db DB) create(ctx context.Context, article entity.Article) error {
	prefixedId := addArticlesPrefix(article.Id)

	_, err := db.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		values := mapArticle(article)
		db.redis.HSet(ctx, prefixedId, values)
		db.redis.SAdd(ctx, articlesPrefix, article.Id)
		return nil
	})

	return err
}
