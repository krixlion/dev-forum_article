package query

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

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

	prefixedId := addArticlesPrefix(article.Id)

	_, err := db.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		values := mapArticle(article)
		db.redis.HSet(ctx, prefixedId, values)
		db.redis.SAdd(ctx, articlesPrefix, article.Id)
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

	prefixedId := addArticlesPrefix(article.Id)

	_, err := db.redis.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		values := mapArticle(article)
		db.redis.HSet(ctx, prefixedId, values)
		db.redis.SAdd(ctx, articlesPrefix, article.Id)
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
		id = addArticlesPrefix(id)
		db.redis.Del(ctx, id).Result()
		return nil
	})

	return err
}

type scanCmder interface {
	Scan(dst interface{}) error
}

// scan enhances the Scan method of a scanCmder with these features:
//   - it returns the error redis.Nil when the key does not exist. See https://github.com/go-redis/redis/issues/1668
//   - it supports embedded struct better. See https://github.com/go-redis/redis/issues/2005#issuecomment-1019667052
func scan(s scanCmder, dest ...interface{}) error {
	switch cmd := s.(type) {
	case *redis.MapStringStringCmd:
		if len(cmd.Val()) == 0 {
			return redis.Nil
		}
	case *redis.SliceCmd:
		keyExists := false
		for _, v := range cmd.Val() {
			if v != nil {
				keyExists = true
				break
			}
		}
		if !keyExists {
			return redis.Nil
		}
	}

	for _, d := range dest {
		if err := s.Scan(d); err != nil {
			return err
		}
	}

	return nil
}

func addArticlesPrefix(key string) string {
	return fmt.Sprintf("%s:%s", articlesPrefix, key)
}

func mapArticle(article entity.Article) map[string]string {
	v := reflect.ValueOf(article)
	values := make(map[string]string)

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if s := field.String(); s != "" {
			fieldName := strings.ToLower(v.Type().Field(i).Name)
			values[fieldName] = s
		}
	}
	return values
}
