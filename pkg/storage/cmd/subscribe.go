package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/tracing"
	"go.opentelemetry.io/otel"
)

// Subscribe listens for all events in the eventstore and invokes a
// handlerFunc on each received event.
// It's up to the handlerFunc to process or ignore the event.
func (db DB) Subscribe(ctx context.Context, handle event.HandlerFunc, entity entity.EntityName, types ...event.EventType) error {
	prefix := fmt.Sprintf("%s-", entity)
	stream, err := db.client.SubscribeToAll(ctx, esdb.SubscribeToAllOptions{
		From: esdb.End{},
		Filter: &esdb.SubscriptionFilter{
			Type:     esdb.StreamFilterType,
			Prefixes: []string{prefix},
		},
	})

	if err != nil {
		db.logger.Log(ctx, "Failed to subscribe", "err", err)
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

		ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "esdb.Subscribe")

		if subEvent.SubscriptionDropped != nil {
			err := subEvent.SubscriptionDropped.Error
			db.logger.Log(ctx, "Subscription dropped", "err", err)
			tracing.SetSpanErr(span, err)
			return err
		}

		event := event.Event{}
		data := subEvent.EventAppeared.OriginalEvent().Data

		err := json.Unmarshal(data, &event)
		if err != nil {
			tracing.SetSpanErr(span, err)
			db.logger.Log(ctx, "failed to parse event data", "err", err, "data", data)
			// TODO: append to retry queue
			continue
		}

		if err := handle(ctx, event); err != nil {
			db.logger.Log(ctx, "failed to update read model", "err", err, "event", event)
			tracing.SetSpanErr(span, err)
			// TODO: append to retry queue
		}
		span.End()
	}
}
