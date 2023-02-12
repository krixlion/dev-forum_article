package server

import (
	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-proto/article_service/pb"
)

func articleFromPB(v *pb.Article) entity.Article {
	return entity.Article{
		Id:        v.GetId(),
		UserId:    v.GetUserId(),
		Title:     v.GetTitle(),
		Body:      v.GetBody(),
		CreatedAt: v.GetCreatedAt().AsTime(),
		UpdatedAt: v.GetUpdatedAt().AsTime(),
	}
}
