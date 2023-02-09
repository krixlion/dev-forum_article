package eventstore

import (
	"fmt"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/krixlion/dev_forum-lib/logging"
	"go.opentelemetry.io/otel/trace"
)

type DB struct {
	client *esdb.Client
	tracer trace.Tracer
	logger logging.Logger
}

func formatConnString(port, host, user, pass string) string {
	return fmt.Sprintf("esdb://%s:%s@%s:%s?tls=false", user, pass, host, port)
}

func MakeDB(port, host, user, pass string, logger logging.Logger, tracer trace.Tracer) (DB, error) {
	url := formatConnString(port, host, user, pass)
	config, err := esdb.ParseConnectionString(url)
	if err != nil {
		return DB{}, err
	}

	client, err := esdb.NewClient(config)
	if err != nil {
		return DB{}, err
	}

	return DB{
		client: client,
		logger: logger,
		tracer: tracer,
	}, nil
}

func (db DB) Close() error {
	return db.client.Close()
}
