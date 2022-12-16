package cmd

import (
	"context"
	"encoding/json"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/krixlion/dev-forum_article/pkg/event"
)

// Subscribe blocks
func (db DB) Subscribe(ctx context.Context, fn event.HandlerFunc) error {

	options := esdb.SubscribeToStreamOptions{
		From: esdb.End{},
	}

	stream, err := db.client.SubscribeToStream(ctx, "some-stream", options)
	if err != nil {
		return err
	}

	defer stream.Close()

	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		subEvent := stream.Recv()

		if subEvent.EventAppeared == nil {
			continue
		}

		if err := subEvent.SubscriptionDropped.Error; err != nil {
			return err
		}

		event := event.Event{}

		err := json.Unmarshal(subEvent.EventAppeared.Event.Data, &event)
		if err != nil {
			return err
		}

		if err := fn(ctx, event); err != nil {
			return err
		}
	}
}
