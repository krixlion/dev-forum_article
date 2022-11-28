package entity

import "github.com/krixlion/dev-forum_article/pkg/net/grpc/pb"

// This service's entity.
type Article struct {
	id     string
	userId string // Author's ID.
	Title  string
	Body   string
}

func (v *Article) UserId() string {
	if v != nil {
		return v.userId
	}

	return ""
}

func (v *Article) Id() string {
	if v != nil {
		return v.id
	}

	return ""
}

func MakeArticleFromPb(v *pb.Article) Article {
	id := v.GetId()
	userId := v.GetUserId()
	title := v.GetTitle()
	body := v.GetBody()

	return Article{
		id:     id,
		userId: userId,
		Title:  title,
		Body:   body,
	}
}
