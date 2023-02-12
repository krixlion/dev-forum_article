package eventstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/tracing"
	"go.opentelemetry.io/otel/trace"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
)

// Consume listens for new article event streams with given type and sends them through returned channel.
func (db DB) Consume(ctx context.Context, _ string, eType event.EventType) (<-chan event.Event, error) {
	options := esdb.SubscribeToAllOptions{
		From:           esdb.End{},
		ResolveLinkTos: true,
		Filter: &esdb.SubscriptionFilter{
			Type:  esdb.EventFilterType,
			Regex: string(eType),
		},
	}
	events := make(chan event.Event)

	go func() {
		for {
			stream, err := db.client.SubscribeToAll(ctx, options)
			if err != nil {
				db.logger.Log(ctx, "Failed to subscribe", "err", err)
				time.Sleep(time.Second)
				continue
			}

			for {
				subEvent := stream.Recv()

				if subEvent.SubscriptionDropped != nil {
					db.logger.Log(ctx, "CatchUp subscription dropped", "err", err)
					stream.Close()
					break
				}

				if subEvent.EventAppeared == nil {
					continue
				}

				ctx, span := db.tracer.Start(ctx, "esdb.Consume", trace.WithSpanKind(trace.SpanKindConsumer))

				originalEvent := subEvent.EventAppeared.OriginalEvent()
				options.From = originalEvent.Position

				event := event.Event{}
				data := originalEvent.Data

				err := json.Unmarshal(data, &event)
				if err != nil {
					db.logger.Log(ctx,
						"failed to parse event data",
						"err", err,
						"data", data,
						"streamId", originalEvent.StreamID,
						"eventId", originalEvent.EventID,
						"revision", originalEvent.EventNumber,
					)

					tracing.SetSpanErr(span, err)
					span.End()
					continue
				}
				events <- event
				span.End()
			}
		}
	}()
	return events, nil
}
