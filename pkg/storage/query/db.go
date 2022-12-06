package query

import (
	"context"

	"github.com/go-redis/redis/v9"
)

type DB struct {
	redis *redis.Client
}

func MakeDB(host, port, pass string) DB {

	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: pass,
		DB:       0, // use default DB
	})

	return DB{
		redis: rdb,
	}
}

func (db DB) Ping(ctx context.Context) error {
	return db.redis.Ping(ctx).Err()
}
