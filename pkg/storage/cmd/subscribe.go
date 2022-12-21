package cmd

import (
	"context"
	"encoding/json"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

// Subscribe blocks.
func (db DB) Subscribe(ctx context.Context, fn event.HandlerFunc) error {
	ctx, span := otel.Tracer(tracing.ServiceName).Start(ctx, "Subscribe")
	defer span.End()

	options := esdb.SubscribeToStreamOptions{
		From: esdb.End{},
	}

	stream, err := db.client.SubscribeToStream(ctx, "articles", options)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	defer stream.Close()

	for {
		if err := ctx.Err(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		subEvent := stream.Recv()
		if subEvent.EventAppeared == nil {
			continue
		}

		if err := subEvent.SubscriptionDropped.Error; err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		event := event.Event{}

		err := json.Unmarshal(subEvent.EventAppeared.OriginalEvent().Data, &event)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}

		if err := fn(ctx, event); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return err
		}
	}
}
