package query

import (
	"time"

	"github.com/krixlion/dev_forum-article/pkg/entity"
)

type redisArticle struct {
	Id        string `redis:"id"`
	UserId    string `redis:"user_id"`
	Title     string `redis:"title"`
	Body      string `redis:"body"`
	CreatedAt string `redis:"created_at"`
	UpdatedAt string `redis:"updated_at"`
}

func (a redisArticle) CreatedAtTime() (time.Time, error) {
	return time.Parse(time.RFC3339, a.CreatedAt)
}

func (a redisArticle) UpdatedAtTime() (time.Time, error) {
	return time.Parse(time.RFC3339, a.UpdatedAt)
}

func (a redisArticle) Article() (entity.Article, error) {
	createdAt, err := a.CreatedAtTime()
	if err != nil {
		return entity.Article{}, err
	}

	updatedAt, err := a.UpdatedAtTime()
	if err != nil {
		return entity.Article{}, err
	}

	return entity.Article{
		Id:        a.Id,
		UserId:    a.UserId,
		Title:     a.Title,
		Body:      a.Body,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}
