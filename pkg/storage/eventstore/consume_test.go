package eventstore

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/helpers/gentest"
)

func Test_Consume(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Consume() integration test.")
	}

	type args struct {
		ctx   context.Context
		eType event.EventType
	}
	testCases := []struct {
		desc    string
		args    args
		want    event.Event
		wantErr bool
	}{
		{
			desc: "Test if correctly returns ArticleCreated events",
			args: args{
				ctx:   context.Background(),
				eType: event.ArticleCreated,
			},
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
			desc: "Test if correctly returns ArticleUpdated events",
			args: args{
				ctx:   context.Background(),
				eType: event.ArticleUpdated,
			},
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
			desc: "Test if correctly returns ArticleDeleted events",
			args: args{
				ctx:   context.Background(),
				eType: event.ArticleDeleted,
			},
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
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			stream, err := db.Consume(tC.args.ctx, "", tC.args.eType)
			if (err != nil) != tC.wantErr {
				t.Errorf("DB.Consume():\n error = %v\n, wantErr %v\n", err, tC.wantErr)
				return
			}

			var article entity.Article
			var id string
			if tC.want.Type != event.ArticleDeleted {
				err = json.Unmarshal(tC.want.Body, &article)
				if err != nil {
					t.Errorf("Failed to unmarshal random JSON article:\n error = %+v\n", err)
				}
			} else {
				err = json.Unmarshal(tC.want.Body, &id)
				if err != nil {
					t.Errorf("Failed to unmarshal random JSON id:\n error = %+v\n", err)
				}
			}

			switch tC.want.Type {
			case event.ArticleCreated:
				err = db.Create(tC.args.ctx, article)
			case event.ArticleDeleted:
				err = db.Delete(tC.args.ctx, id)
			case event.ArticleUpdated:
				err = db.Update(tC.args.ctx, article)

			}
			if err != nil {
				t.Errorf("Failed to emit %s event:\n error = %+v\n", tC.want.Type, err)
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()

			select {
			case got := <-stream:
				if !cmp.Equal(got, tC.want, cmpopts.EquateApproxTime(time.Second)) {
					t.Errorf("DB.Consume():\n got = %v\n want = %v\n Difference =  %s\n", got, tC.want, cmp.Diff(got, tC.want))
				}
			case <-ctx.Done():
				t.Error("Timed out waiting for an event")
				return
			}
		})
	}
}
