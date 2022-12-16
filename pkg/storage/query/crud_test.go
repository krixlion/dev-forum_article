package query_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/env"
	"github.com/krixlion/dev-forum_article/pkg/storage/query"
)

func init() {
	env.Load("app")
}

func pingRedis(ctx context.Context, t *testing.T, db query.DB) {
	err := db.Ping(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v", err)
	}
}

func TestCRUDOnSimpleData(t *testing.T) {

	port := os.Getenv("DB_READ_PORT")
	host := os.Getenv("DB_READ_HOST")
	pass := os.Getenv("DB_READ_PASS")
	db := query.MakeDB(host, port, pass)

	pingRedis(context.Background(), t, db)

	id, err := uuid.NewV1()
	if err != nil {
		t.Errorf("Failed to create article id: %v", err)
	}

	mockArticle := entity.Article{
		Id:     id.String(),
		UserId: "1",
		Title:  "title lol",
		Body:   "body lol",
	}

	type test struct {
		desc     string
		arg      entity.Article
		expected entity.Article
	}
	testCases := []test{
		{
			desc:     "Check if created article is later returned correctly.",
			arg:      mockArticle,
			expected: mockArticle,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.desc, func(t *testing.T) {
			if err := db.Create(context.Background(), mockArticle); err != nil {
				t.Errorf("Failed to Create article: %v", err)
			}

			article, err := db.Get(context.Background(), mockArticle.Id)
			if err != nil {
				t.Errorf("Failed to Get article: %v", err)
			}

			if !reflect.DeepEqual(mockArticle, article) {
				t.Errorf("Articles are not equal, received: %+v\n", article)
			}

			if err := db.Delete(context.Background(), mockArticle.Id); err != nil {
				t.Errorf("Failed to Delete article: %v", err)
			}
		})
	}
}
