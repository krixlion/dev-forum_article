package query

import (
	"context"

	"github.com/go-redis/redis/extra/redisotel/v9"
	"github.com/go-redis/redis/v9"
	"github.com/krixlion/dev-forum_article/pkg/logging"
)

type DB struct {
	redis  *redis.Client
	logger logging.Logger
}

func MakeDB(host, port, pass string, logger logging.Logger) (DB, error) {

	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: pass,
		DB:       0, // use default DB
	})

	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		return DB{}, err
	}

	if err := redisotel.InstrumentTracing(rdb); err != nil {
		return DB{}, err
	}

	return DB{
		redis:  rdb,
		logger: logger,
	}, nil
}

func (db DB) Ping(ctx context.Context) error {
	return db.redis.Ping(ctx).Err()
}
