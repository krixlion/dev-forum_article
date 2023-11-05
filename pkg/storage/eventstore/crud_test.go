package eventstore

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-article/internal/gentest"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/nulls"
)

func setUpDB() Eventstore {
	env.Load("app")
	port := os.Getenv("DB_WRITE_PORT")
	host := os.Getenv("DB_WRITE_HOST")
	pass := os.Getenv("DB_WRITE_PASS")
	user := os.Getenv("DB_WRITE_USER")

	db, err := MakeDB(port, host, user, pass, nulls.NullLogger{}, nulls.NullTracer{})
	if err != nil {
		log.Fatalf("Failed to make DB, err: %s", err)
	}

	return db
}

func Test_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Create() integration test...")
	}

	type args struct {
		ctx     context.Context
		article entity.Article
	}
	tests := []struct {
		desc    string
		args    args
		wantErr bool
	}{
		{
			desc: "Test if ArticleCreated is correctly emitted on random article",
			args: args{
				ctx: context.Background(),
				article: func() entity.Article {
					a := gentest.RandomArticle(2, 5)
					a.Id = "test"
					return a
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			if err := db.Create(tt.args.ctx, tt.args.article); (err != nil) != tt.wantErr {
				t.Errorf("DB.Create() error = %v\n, wantErr %v\n", err, tt.args.article)
			}
			opts := esdb.ReadStreamOptions{
				Direction: esdb.Backwards,
				From:      esdb.End{},
			}

			stream, err := db.client.ReadStream(tt.args.ctx, addArticlesPrefix(tt.args.article.Id), opts, 1)
			if err != nil {
				t.Errorf("Failed to read stream:\n error = %+v\n", err)
				return
			}
			recvEvent, err := stream.Recv()
			if err != nil {
				t.Errorf("Failed to receive from stream:\n error = %+v\n", err)
				return
			}

			var e event.Event
			if err := json.Unmarshal(recvEvent.OriginalEvent().Data, &e); err != nil {
				t.Errorf("Failed to unmarshal event:\n error = %+v\n", err)
				return
			}

			if e.Type != event.ArticleCreated {
				t.Errorf("Invalid EventType:\n got = %s\n want = %s\n", e.Type, event.ArticleCreated)
				return
			}

			var got entity.Article
			if err := json.Unmarshal(e.Body, &got); err != nil {
				t.Errorf("Failed to unmarshal event body into article:\n error = %+v\n", err)
				return
			}

			if !cmp.Equal(tt.args.article, got) {
				t.Errorf("Articles are not equal:\n got = %+v\n want = %+v\n %v\n", got, tt.args.article, cmp.Diff(got, tt.args.article))
				return
			}
		})
	}
}

func Test_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Update() integration test...")
	}

	type args struct {
		ctx     context.Context
		article entity.Article
	}
	tests := []struct {
		desc    string
		args    args
		wantErr bool
	}{
		{
			desc: "Test if ArticleUpdated event is correctly saved",
			args: args{
				ctx: context.Background(),
				article: func() entity.Article {
					a := gentest.RandomArticle(2, 5)
					a.Id = "test"
					return a
				}(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			if err := db.Update(tt.args.ctx, tt.args.article); (err != nil) != tt.wantErr {
				t.Errorf("DB.Update() error = %v\n, wantErr %v\n", err, tt.wantErr)
			}

			opts := esdb.ReadStreamOptions{
				Direction: esdb.Backwards,
				From:      esdb.End{},
			}

			stream, err := db.client.ReadStream(tt.args.ctx, addArticlesPrefix(tt.args.article.Id), opts, 1)
			if err != nil {
				t.Errorf("Failed to read stream:\n error = %+v\n", err)
				return
			}
			recvEvent, err := stream.Recv()
			if err != nil {
				t.Errorf("Failed to receive from stream:\n error = %+v\n", err)
				return
			}

			var e event.Event
			if err := json.Unmarshal(recvEvent.OriginalEvent().Data, &e); err != nil {
				t.Errorf("Failed to unmarshal event:\n error = %+v\n", err)
				return
			}

			if e.Type != event.ArticleUpdated {
				t.Errorf("Invalid EventType:\n got = %s\n want = %s\n", e.Type, event.ArticleUpdated)
				return
			}

			var got entity.Article
			if err := json.Unmarshal(e.Body, &got); err != nil {
				t.Errorf("Failed to unmarshal event body into article:\n error = %+v\n", err)
				return
			}

			if !cmp.Equal(tt.args.article, got) {
				t.Errorf("Articles are not equal:\n got = %+v\n want = %+v\n %v\n", got, tt.args.article, cmp.Diff(got, tt.args.article))
				return
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Delete() integration test...")
	}

	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		desc    string
		args    args
		wantErr bool
	}{
		{
			desc: "Test if ArticleDeleted event is correctly emitted",
			args: args{
				ctx: context.Background(),
				id:  "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			if err := db.Delete(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DB.Delete() error = %v\n, wantErr %v\n", err, tt.wantErr)
			}

			opts := esdb.ReadStreamOptions{
				Direction: esdb.Backwards,
				From:      esdb.End{},
			}

			stream, err := db.client.ReadStream(tt.args.ctx, addArticlesPrefix(tt.args.id), opts, 1)
			if err != nil {
				t.Errorf("Failed to read stream:\n error = %+v\n", err)
				return
			}
			recvEvent, err := stream.Recv()
			if err != nil {
				t.Errorf("Failed to receive from stream:\n error = %+v\n", err)
				return
			}

			var e event.Event
			if err := json.Unmarshal(recvEvent.OriginalEvent().Data, &e); err != nil {
				t.Errorf("Failed to unmarshal event:\n error = %+v\n", err)
				return
			}

			if e.Type != event.ArticleDeleted {
				t.Errorf("Invalid EventType:\n got = %s\n want = %s\n", e.Type, event.ArticleDeleted)
				return
			}

			var got string
			if err := json.Unmarshal(e.Body, &got); err != nil {
				t.Errorf("Failed to unmarshal event body into ID string:\n error = %+v\n", err)
				return
			}

			if !cmp.Equal(tt.args.id, got) {
				t.Errorf("IDs are not equal:\n got = %+v\n want = %+v\n", got, tt.args.id)
				return
			}
		})
	}
}

func Test_lastRevision(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping lastRevision() integration test...")
	}

	tests := []struct {
		desc    string
		article entity.Article
		wantErr bool
	}{
		{
			desc:    "Test if correctly returns simple article revision.",
			article: gentest.RandomArticle(5, 5),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			if err := db.Create(ctx, tt.article); err != nil {
				t.Errorf("DB.lastRevision() error during prep = %v", err)
				return
			}

			want, err := event.MakeEvent(event.ArticleAggregate, event.ArticleCreated, tt.article)
			if err != nil {
				t.Errorf("DB.lastRevision() error during prep = %v", err)
				return
			}

			resEvent, err := db.lastRevision(ctx, tt.article.Id)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.lastRevision() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			data := resEvent.OriginalEvent().Data
			var got event.Event
			if err = json.Unmarshal(data, &got); err != nil {
				t.Errorf("Failed to unmarshal last revision event, error = %v", err)
				return
			}

			if !cmp.Equal(got, want, cmpopts.EquateApproxTime(time.Second)) {
				t.Errorf("DB.lastRevision():\n got = %v\n want = %v\n %v\n", got, want, cmp.Diff(got, want))
			}
		})
	}
}
