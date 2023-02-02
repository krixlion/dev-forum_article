package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/krixlion/dev_forum-lib/event"
)

func (db Manager) EventHandlers() map[event.EventType][]event.Handler {
	return map[event.EventType][]event.Handler{
		event.ArticleCreated: {
			db.DeleteUsersArticles(),
		},
	}
}

func (db Manager) DeleteUsersArticles() event.HandlerFunc {
	return func(e event.Event) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		var userId string
		if err := json.Unmarshal(e.Body, &userId); err != nil {
			db.logger.Log(ctx, "Failed to unmarshal event body", "err", err, "event", e)
			return
		}

		articleIds, err := db.GetBelongingIDs(ctx, userId)
		if err != nil {
			db.logger.Log(ctx, "Failed to get belonging IDs", "err", err)
			return
		}

		for _, id := range articleIds {
			if err := db.Delete(ctx, id); err != nil {
				db.logger.Log(ctx, "Failed to delete article", "err", err)
			}
		}
	}
}
