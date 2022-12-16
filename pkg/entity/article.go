package entity

import "github.com/krixlion/dev-forum_article/pkg/net/grpc/pb"

const ArticleEntityName = "article"

// This service's entity.
type Article struct {
	Id     string `redis:"id"`
	UserId string `redis:"user_id"` // Author's ID.
	Title  string `redis:"title"`
	Body   string `redis:"body"`
}

func MakeArticleFromPb(v *pb.Article) Article {
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
