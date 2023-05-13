package redis

import (
	"context"

	"github.com/krixlion/dev_forum-article/pkg/storage"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	_ "go.opentelemetry.io/otel/metric/global"
	"go.opentelemetry.io/otel/trace"
)

var _ storage.Getter = (*Redis)(nil)

const (
	articlesPrefix = "articles"
	usersPrefix    = "users"
)

type Redis struct {
	redis  *redis.Client
	logger logging.Logger
	tracer trace.Tracer
}

type tracerProvider struct {
	tracer trace.Tracer
}

func (t tracerProvider) Tracer(string, ...trace.TracerOption) trace.Tracer {
	return t.tracer
}

func MakeDB(host, port, pass string, logger logging.Logger, tracer trace.Tracer) (Redis, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     host + ":" + port,
		Password: pass,
		DB:       0, // use default DB
	})

	if err := redisotel.InstrumentMetrics(rdb); err != nil {
		return Redis{}, err
	}

	if err := redisotel.InstrumentTracing(rdb, redisotel.WithTracerProvider(tracerProvider{tracer: tracer})); err != nil {
		return Redis{}, err
	}

	return Redis{
		redis:  rdb,
		logger: logger,
		tracer: tracer,
	}, nil
}

func (db Redis) Ping(ctx context.Context) error {
	return db.redis.Ping(ctx).Err()
}
