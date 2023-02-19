package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/tracing"
)

// EventHandlers is an Event Handlers provider.
func (db Manager) EventHandlers() map[event.EventType][]event.Handler {
	return map[event.EventType][]event.Handler{
		event.UserDeleted: {
			db.DeleteUsersArticles(),
		},
	}
}

func (db Manager) DeleteUsersArticles() event.HandlerFunc {
	return func(e event.Event) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		ctx, span := db.tracer.Start(ctx, "DeleteUsersArticles")
		defer span.End()

		var userId string
		if err := json.Unmarshal(e.Body, &userId); err != nil {
			db.logger.Log(ctx, "Failed to unmarshal event body", "err", err, "event", e)
			tracing.SetSpanErr(span, err)
			return
		}

		articleIds, err := db.GetBelongingIDs(ctx, userId)
		if err != nil {
			db.logger.Log(ctx, "Failed to get belonging IDs", "err", err)
			tracing.SetSpanErr(span, err)
			return
		}
		for _, id := range articleIds {
			if err := db.Delete(ctx, id); err != nil {
				db.logger.Log(ctx, "Failed to delete article", "err", err)
				tracing.SetSpanErr(span, err)
			}
		}
	}
}
