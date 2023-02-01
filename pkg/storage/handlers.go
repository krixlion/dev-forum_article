package storage

import (
	"context"
	"encoding/json"

	"github.com/krixlion/dev_forum-article/pkg/event"
)

func (db Manager) DeleteUsersArticles(e event.Event) event.HandlerFunc {
	return func(e event.Event) {
		ctx, cancel := context.WithCancel(context.Background())
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
