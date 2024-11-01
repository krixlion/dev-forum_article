package testdata

import (
	"context"
	"os"
	"time"

	"github.com/krixlion/dev_forum-lib/env"
	"github.com/redis/go-redis/v9"
)

func init() {
	if err := initTestData(); err != nil {
		panic(err)
	}
}

func Seed() error {
	if err := env.Load("app"); err != nil {
		return err
	}

	port := os.Getenv("DB_READ_PORT")
	host := os.Getenv("DB_READ_HOST")
	pass := os.Getenv("DB_READ_PASS")

	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: pass,
		DB:       0, // use default DB
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err := rdb.FlushDB(ctx).Err(); err != nil {
		return err
	}

	_, err := rdb.Pipelined(ctx, func(p redis.Pipeliner) error {
		for _, article := range ArticleMaps {
			p.HSet(ctx, "articles:"+article["id"], article)
			p.SAdd(ctx, "articles", article["id"])
			p.SAdd(ctx, "users"+article["user_id"], article["id"])
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
