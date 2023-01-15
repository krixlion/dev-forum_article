package eventstore

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/EventStore/EventStore-Client-Go/v3/esdb"
	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/env"
	"github.com/krixlion/dev_forum-article/pkg/event"
	"github.com/krixlion/dev_forum-article/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-article/pkg/helpers/nulls"
)

var (
	port string
	host string
	user string
	pass string
)

func init() {
	env.Load("app")
	port = os.Getenv("DB_WRITE_PORT")
	host = os.Getenv("DB_WRITE_HOST")
	pass = os.Getenv("DB_WRITE_PASS")
	user = os.Getenv("DB_WRITE_USER")
}

func setUpDB() DB {
	db, err := MakeDB(port, host, user, pass, nulls.NullLogger{})
	if err != nil {
		log.Fatalf("Failed to make DB, err: %s", err)
	}

	return db
}

func Test_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Create() integration test.")
	}

	type args struct {
		ctx     context.Context
		article entity.Article
	}
	testCases := []struct {
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
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			if err := db.Create(tC.args.ctx, tC.args.article); (err != nil) != tC.wantErr {
				t.Errorf("DB.Create() error = %v\n, wantErr %v\n", err, tC.args.article)
			}
			opts := esdb.ReadStreamOptions{
				Direction: esdb.Backwards,
				From:      esdb.End{},
			}

			stream, err := db.client.ReadStream(tC.args.ctx, addArticlesPrefix(tC.args.article.Id), opts, 1)
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

			if !cmp.Equal(tC.args.article, got) {
				t.Errorf("Articles are not equal:\n got = %+v\n want = %+v\n", got, tC.args.article)
				return
			}
		})
	}
}

func Test_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Update() integration test.")
	}

	type args struct {
		ctx     context.Context
		article entity.Article
	}
	testCases := []struct {
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
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			if err := db.Update(tC.args.ctx, tC.args.article); (err != nil) != tC.wantErr {
				t.Errorf("DB.Update() error = %v\n, wantErr %v\n", err, tC.wantErr)
			}

			opts := esdb.ReadStreamOptions{
				Direction: esdb.Backwards,
				From:      esdb.End{},
			}

			stream, err := db.client.ReadStream(tC.args.ctx, addArticlesPrefix(tC.args.article.Id), opts, 1)
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

			if !cmp.Equal(tC.args.article, got) {
				t.Errorf("Articles are not equal:\n got = %+v\n want = %+v\n", got, tC.args.article)
				return
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Delete() integration test.")
	}

	type args struct {
		ctx context.Context
		id  string
	}
	testCases := []struct {
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
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()

			if err := db.Delete(tC.args.ctx, tC.args.id); (err != nil) != tC.wantErr {
				t.Errorf("DB.Delete() error = %v\n, wantErr %v\n", err, tC.wantErr)
			}

			opts := esdb.ReadStreamOptions{
				Direction: esdb.Backwards,
				From:      esdb.End{},
			}

			stream, err := db.client.ReadStream(tC.args.ctx, addArticlesPrefix(tC.args.id), opts, 1)
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

			if !cmp.Equal(tC.args.id, got) {
				t.Errorf("IDs are not equal:\n got = %+v\n want = %+v\n", got, tC.args.id)
				return
			}
		})
	}
}

func Test_lastRevision(t *testing.T) {
	type args struct {
		ctx       context.Context
		articleId string
	}
	tests := []struct {
		desc    string
		args    args
		want    event.Event
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			db := setUpDB()
			defer db.Close()
			got, err := db.lastRevision(tt.args.ctx, tt.args.articleId)
			if (err != nil) != tt.wantErr {
				t.Errorf("DB.lastRevision() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want) {
				t.Errorf("DB.lastRevision() = %v, want %v", got, tt.want)
			}
		})
	}
}
