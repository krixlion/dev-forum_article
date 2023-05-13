package testdata

import (
	"strconv"
	"time"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/helpers/gentest"
)

var (
	ArticleMaps map[string]map[string]string

	Articles map[string]entity.Article

	ArticleMap = map[string]string{
		"id":         gentest.RandomString(5),
		"user_id":    gentest.RandomString(5),
		"title":      "title-" + gentest.RandomString(5),
		"body":       "body-" + gentest.RandomString(5),
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}
)

func initTestData() error {
	count := 8
	ArticleMaps = make(map[string]map[string]string, count)
	Articles = make(map[string]entity.Article, count)

	for i := 1; i <= count; i++ {
		id := strconv.Itoa(i)
		ArticleMaps[id] = map[string]string{
			"id":         id,
			"user_id":    id,
			"title":      "title-" + id,
			"body":       "body-" + id,
			"created_at": "2023-01-31T22:58:24Z",
			"updated_at": "2023-01-31T22:58:24Z",
		}

		t, err := time.Parse(time.RFC3339, "2023-01-31T22:58:24Z")
		if err != nil {
			return err
		}

		Articles[id] = entity.Article{
			Id:        id,
			UserId:    id,
			Title:     "title-" + id,
			Body:      "body-" + id,
			CreatedAt: t,
			UpdatedAt: t,
		}
	}

	return nil
}
