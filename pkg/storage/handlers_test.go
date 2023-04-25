package storage

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-article/pkg/storage/eventstore"
	"github.com/krixlion/dev_forum-article/pkg/storage/query"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/stretchr/testify/mock"
)

func setUpDBAndDispatcher() (Manager, *dispatcher.Dispatcher) {
	env.Load("app")
	cmd_port := os.Getenv("DB_WRITE_PORT")
	cmd_host := os.Getenv("DB_WRITE_HOST")
	cmd_user := os.Getenv("DB_WRITE_USER")
	cmd_pass := os.Getenv("DB_WRITE_PASS")
	cmd, err := eventstore.MakeDB(cmd_port, cmd_host, cmd_user, cmd_pass, nulls.NullLogger{}, nulls.NullTracer{})
	if err != nil {
		panic(err)
	}

	query_port := os.Getenv("DB_READ_PORT")
	query_host := os.Getenv("DB_READ_HOST")
	query_pass := os.Getenv("DB_READ_PASS")
	query, err := query.MakeDB(query_host, query_port, query_pass, nulls.NullLogger{}, nulls.NullTracer{})
	if err != nil {
		panic(err)
	}

	m := Manager{
		cmd:    cmd,
		query:  query,
		logger: nulls.NullLogger{},
		tracer: nulls.NullTracer{},
	}

	d := dispatcher.NewDispatcher(mocks.Broker{Mock: new(mock.Mock)}, 1)
	ch, err := cmd.Consume(context.Background(), "", event.ArticleDeleted)
	if err != nil {
		panic(err)
	}
	d.SetSyncHandler(event.HandlerFunc(m.CatchUp))
	d.AddSyncEventProviders(ch)

	return m, d
}

func TestManager_DeleteUsersArticles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DeleteUsersArticles integration test.")
	}

	tests := []struct {
		name string
		want event.Event
	}{
		{
			name: "Test if everything is called as anticipated",
			want: func() event.Event {
				event, err := event.MakeEvent(event.UserAggregate, event.UserDeleted, "10")
				if err != nil {
					panic(err)
				}
				return event
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			db, d := setUpDBAndDispatcher()
			d.Subscribe(tt.want.Type, db.DeleteUsersArticles())
			go d.Run(context.Background())
			d.Dispatch(tt.want)

			// Wait for the handler to be run in a seperate goroutine.
			time.Sleep(time.Second)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			var userId string
			err := json.Unmarshal(tt.want.Body, &userId)
			if err != nil {
				panic(err)
			}

			got, err := db.GetBelongingIDs(ctx, userId)
			if err != nil {
				t.Errorf("Manager.DeleteUsersArticles(): error = %v", err)
				return
			}

			if !cmp.Equal(got, []string{}, cmpopts.EquateEmpty()) {
				t.Errorf("Manager.DeleteUsersArticles():\n got = %v\n want %v\n", got, []string{})
				return
			}
		})
	}
}
