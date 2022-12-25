package query_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/env"
	"github.com/krixlion/dev-forum_article/pkg/helpers/gentest"
	"github.com/krixlion/dev-forum_article/pkg/storage/query"
)

var (
	port string
	host string
	pass string
)

func init() {
	env.Load("app")
	port = os.Getenv("DB_READ_PORT")
	host = os.Getenv("DB_READ_HOST")
	pass = os.Getenv("DB_READ_PASS")
}

func setUpDB() query.DB {
	db, err := query.MakeDB(host, port, pass)
	if err != nil {
		log.Fatalf("Failed to make DB, err: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err = db.Ping(ctx)
	if err != nil {
		log.Fatalf("Failed to ping to DB: %v", err)
	}

	return db
}

func TestCRUD(t *testing.T) {
	db := setUpDB()
	defer db.Close()

	article := gentest.RandomArticle(3, 5)

	type testCase struct {
		desc          string
		arg           entity.Article
		want          entity.Article
		wantCreateErr bool
		wantGetErr    bool
		wantDelErr    bool
	}
	testCases := []testCase{
		{
			desc: "Check if created article is later returned and deleted correctly.",
			arg:  article,
			want: article,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			if err := db.Create(ctx, article); (err != nil) != tC.wantCreateErr {
				t.Errorf("db.Create() err = %v", err)
			}

			got, err := db.Get(ctx, article.Id)
			if (err != nil) != tC.wantGetErr {
				t.Errorf("db.Get() err = %v", err)
			}

			if !cmp.Equal(got, tC.want) {
				t.Errorf("Articles are not equal, got = %+v\n, want = %+v\n", got, tC.want)
			}

			if err := db.Delete(ctx, article.Id); (err != nil) != tC.wantDelErr {
				t.Errorf("db.Delete() err = %v", err)
			}

			v, err := db.Get(ctx, article.Id)
			if err != nil {
				t.Fatalf("Failed to db.Get() after db.Del(), err = %v", err)
			}
			if !cmp.Equal(v, entity.Article{}) {
				t.Errorf("Failed to delete article after testing, got = %v, want = %v", v, entity.Article{})
			}
		})
	}
}
