package query

import (
	"context"

	"github.com/go-redis/redis/extra/redisotel/v9"
	"github.com/go-redis/redis/v9"
)

type DB struct {
	redis *redis.Client
}

func MakeDB(host, port, pass string) (DB, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: pass,
		DB:       0, // use default DB
	})

	err := redisotel.InstrumentMetrics(rdb)
	if err != nil {
		return DB{}, err
	}

	err = redisotel.InstrumentTracing(rdb)
	if err != nil {
		return DB{}, err
	}
	return DB{
		redis: rdb,
	}, nil
}

func (db DB) Ping(ctx context.Context) error {
	return db.redis.Ping(ctx).Err()
}
