package eventstore

import (
	"fmt"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
	"github.com/krixlion/dev_forum-article/pkg/storage"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/logging"
	"go.opentelemetry.io/otel/trace"
)

var _ storage.Writer = (*Eventstore)(nil)
var _ event.Consumer = (*Eventstore)(nil)

type Eventstore struct {
	client *esdb.Client
	tracer trace.Tracer
	logger logging.Logger
}

func MakeDB(port, host, user, pass string, logger logging.Logger, tracer trace.Tracer) (Eventstore, error) {
	url := fmt.Sprintf("esdb://%s:%s@%s:%s?tls=false", user, pass, host, port)
	config, err := esdb.ParseConnectionString(url)
	if err != nil {
		return Eventstore{}, err
	}

	client, err := esdb.NewClient(config)
	if err != nil {
		return Eventstore{}, err
	}

	return Eventstore{
		client: client,
		logger: logger,
		tracer: tracer,
	}, nil
}

func (db Eventstore) Close() error {
	return db.client.Close()
}
