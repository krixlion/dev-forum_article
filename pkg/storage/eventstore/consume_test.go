package eventstore

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-article/internal/gentest"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-lib/event"
)

func Test_Consume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Consume() integration test.")
	}

	tests := []struct {
		desc    string
		eType   event.EventType
		want    event.Event
		wantErr bool
	}{
		{
			desc:  "Test if correctly returns ArticleCreated events",
			eType: event.ArticleCreated,
			want: event.Event{
				Type:        event.ArticleCreated,
				AggregateId: "article",
				Body: func() []byte {
					a := gentest.RandomArticle(2, 5)
					a.Id = "test"
					json, err := json.Marshal(a)
					if err != nil {
						panic("Failed to marshal article, err = " + err.Error())
					}
					return json
				}(),
				Timestamp: time.Now().Round(0),
			},
		},
		{
			desc:  "Test if correctly returns ArticleUpdated events",
			eType: event.ArticleUpdated,
			want: event.Event{
				Type:        event.ArticleUpdated,
				AggregateId: "article",
				Body: func() []byte {
					a := gentest.RandomArticle(2, 5)
					a.Id = "test"
					json, err := json.Marshal(a)
					if err != nil {
						panic("Failed to marshal article, err = " + err.Error())
					}
					return json
				}(),
				Timestamp: time.Now().Round(0),
			},
		},
		{
			desc:  "Test if correctly returns ArticleDeleted events",
			eType: event.ArticleDeleted,
			want: event.Event{
				Type:        event.ArticleDeleted,
				AggregateId: "article",
				Body: func() []byte {
					id, err := json.Marshal("test")
					if err != nil {
						panic("Failed to unmarshal, err = " + err.Error())
					}
					return id
				}(),
				Timestamp: time.Now().Round(0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			stream, err := db.Consume(ctx, "", tt.eType)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.Consume():\n error = %v\n, wantErr %v\n", err, tt.wantErr)
				return
			}

			var article entity.Article
			var id string
			if tt.want.Type != event.ArticleDeleted {
				err = json.Unmarshal(tt.want.Body, &article)
				if err != nil {
					t.Errorf("Failed to unmarshal random JSON article:\n error = %+v\n", err)
				}
			} else {
				err = json.Unmarshal(tt.want.Body, &id)
				if err != nil {
					t.Errorf("Failed to unmarshal random JSON id:\n error = %+v\n", err)
				}
			}

			switch tt.want.Type {
			case event.ArticleCreated:
				err = db.Create(ctx, article)
			case event.ArticleDeleted:
				err = db.Delete(ctx, id)
			case event.ArticleUpdated:
				err = db.Update(ctx, article)
			}

			if err != nil {
				t.Errorf("Failed to emit %s event:\n error = %+v\n", tt.want.Type, err)
			}

			select {
			case got := <-stream:
				if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second*2)) {
					t.Errorf("DB.Consume():\n got = %v\n want = %v\n Difference =  %s\n", got, tt.want, cmp.Diff(got, tt.want))
				}
			case <-ctx.Done():
				t.Error("Timed out waiting for an event")
				return
			}
		})
	}
}
