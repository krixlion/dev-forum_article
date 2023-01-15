package entity

import (
	"github.com/krixlion/dev_forum-proto/article_service/pb"
)

// This service's entity.
type Article struct {
	Id     string `redis:"id" json:"id,omitempty"`
	UserId string `redis:"user_id" json:"user_id,omitempty"` // Author's ID.
	Title  string `redis:"title" json:"title,omitempty"`
	Body   string `redis:"body" json:"body,omitempty"`
}

func ArticleFromPB(v *pb.Article) Article {
	return Article{
		Id:     v.GetId(),
		UserId: v.GetUserId(),
		Title:  v.GetTitle(),
		Body:   v.GetBody(),
	}
}
