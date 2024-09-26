package redis

import (
	"context"
	"errors"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-lib/str"
	"github.com/krixlion/dev_forum-lib/tracing"
	"github.com/redis/go-redis/v9"
)

func (db Redis) Close() error {
	return db.redis.Close()
}

func (db Redis) Get(ctx context.Context, id string) (_ entity.Article, err error) {
	ctx, span := db.tracer.Start(ctx, "redis.Get")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	id = addPrefix(articlesPrefix, id)
	var data redisArticle

	cmd := db.redis.HGetAll(ctx, id)
	if err := scan(cmd, &data); err != nil {
		return entity.Article{}, err
	}

	v, err := data.Article()
	if err != nil {
		return entity.Article{}, err
	}
	return v, nil
}

func (db Redis) GetMultiple(ctx context.Context, offset, limit string) (_ []entity.Article, err error) {
	ctx, span := db.tracer.Start(ctx, "redis.GetMultiple")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	off, err := str.ConvertToInt(offset)
	if err != nil {
		return nil, err
	}

	count, err := str.ConvertToInt(limit)
	if err != nil {
		return nil, err
	}

	ids, err := db.redis.Sort(ctx, articlesPrefix, &redis.Sort{
		By:     addPrefix(articlesPrefix, "*->title"),
		Offset: int64(off),
		Count:  int64(count),
		Alpha:  true,
		Order:  "DESC",
	}).Result()
	if err != nil {
		return nil, err
	}

	commands := make([]*redis.MapStringStringCmd, 0, len(ids))
	pipeline := db.redis.Pipeline()

	for _, id := range ids {
		id = addPrefix(articlesPrefix, id)
		commands = append(commands, pipeline.HGetAll(ctx, id))
	}

	if _, err := pipeline.Exec(ctx); err != nil {
		return nil, err
	}

	articles := make([]entity.Article, 0, len(ids))

	for _, cmd := range commands {
		var data redisArticle
		if err := scan(cmd, &data); err != nil {
			if errors.Is(err, redis.Nil) {
				continue
			}
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

func (db Redis) Create(ctx context.Context, article entity.Article) (err error) {
	ctx, span := db.tracer.Start(ctx, "redis.Create")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	prefixedArticleId := addPrefix(articlesPrefix, article.Id)

	_, err = db.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		values := mapArticle(article)
		pipe.HSet(ctx, prefixedArticleId, values)
		pipe.SAdd(ctx, articlesPrefix, article.Id)
		pipe.SAdd(ctx, addPrefix(usersPrefix, article.UserId), article.Id)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (db Redis) Update(ctx context.Context, article entity.Article) (err error) {
	ctx, span := db.tracer.Start(ctx, "redis.Update")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

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
		pipe.HSet(ctx, prefixedId, values)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (db Redis) Delete(ctx context.Context, id string) (err error) {
	ctx, span := db.tracer.Start(ctx, "redis.Delete")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	_, err = db.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		userId, _ := pipe.HGet(ctx, addPrefix(articlesPrefix, id), "user_id").Result()
		pipe.Del(ctx, addPrefix(articlesPrefix, id))
		pipe.SRem(ctx, addPrefix(usersPrefix, userId), id)
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
