package entity

import "github.com/krixlion/dev-forum_article/pkg/net/grpc/pb"

// This service's entity.
type Article struct {
	Id     string `redis:"id" json:"id,omitempty"`
	UserId string `redis:"user_id" json:"user_id,omitempty"` // Author's ID.
	Title  string `redis:"title" json:"title,omitempty"`
	Body   string `redis:"body" json:"body,omitempty"`
}

func ArticleFromPb(v *pb.Article) Article {
	id := v.GetId()
	userId := v.GetUserId()
	title := v.GetTitle()
	body := v.GetBody()

	return Article{
		Id:     id,
		UserId: userId,
		Title:  title,
		Body:   body,
	}
}
