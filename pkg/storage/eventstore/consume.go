package eventstore

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/krixlion/dev_forum-lib/event"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
)

// Consume listens to article event streams with given event type for events and sends them through the returned channel.
func (db Eventstore) Consume(ctx context.Context, _ string, eType event.EventType) (<-chan event.Event, error) {
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
				e, err := parseEvent(subEvent)
				if err != nil {
					if err.Error() == "eventstore stream subscription dropped" {
						stream.Close()
						break
					}
				}

				if isZero(e) {
					continue
				}

				options.From = subEvent.EventAppeared.OriginalEvent().Position
				events <- e
			}
		}
	}()

	return events, nil
}

func parseEvent(subEvent *esdb.SubscriptionEvent) (event.Event, error) {
	if subEvent.SubscriptionDropped != nil {
		err := errors.New("eventstore stream subscription dropped")
		return event.Event{}, err
	}

	if subEvent.EventAppeared == nil {
		return event.Event{}, nil
	}

	var e event.Event
	if err := json.Unmarshal(subEvent.EventAppeared.OriginalEvent().Data, &e); err != nil {
		return event.Event{}, err
	}

	return e, nil
}

func isZero(e event.Event) bool {
	// If all fields are zero values.
	return e.AggregateId == "" && e.Body == nil && e.Timestamp == time.Time{} && e.Type == ""
}
