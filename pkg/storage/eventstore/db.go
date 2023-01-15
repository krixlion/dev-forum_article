package eventstore

import (
	"fmt"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/krixlion/dev_forum-article/pkg/logging"
)

type DB struct {
	logger logging.Logger
	client *esdb.Client
	url    string
}

func formatConnString(port, host, user, pass string) string {
	return fmt.Sprintf("esdb://%s:%s@%s:%s?tls=false", user, pass, host, port)
}

func MakeDB(port, host, user, pass string, logger logging.Logger) (DB, error) {
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
		url:    url,
		client: client,
		logger: logger,
	}, nil
}

func (db DB) Close() error {
	return db.client.Close()
}
