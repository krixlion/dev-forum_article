package eventstore

import (
	"fmt"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/krixlion/dev-forum_article/pkg/logging"
)

type DB struct {
	logger logging.Logger
	client *esdb.Client
	// eventHandler event.Handler
	url string
}

func formatConnString(port, host, user, pass string) string {
	return fmt.Sprintf("esdb://%s:%s@%s:%s?tls=false", user, pass, host, port)
}

func MakeDB(port, host, user, pass string, logger logging.Logger) DB {
	url := formatConnString(port, host, user, pass)
	settings, err := esdb.ParseConnectionString(url)
	if err != nil {
		panic(err)
	}

	client, _ := esdb.NewClient(settings)

	return DB{
		url:    url,
		client: client,
		logger: logger,
	}
}

func (db DB) Close() error {
	return db.client.Close()
}
