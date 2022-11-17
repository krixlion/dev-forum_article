package server

import (
	"context"

	"github.com/krixlion/dev-forum_article/pkg/grpc/pb"
)

type ArticleServer struct {
	pb.UnimplementedArticleServiceServer
}

func (srv ArticleServer) Close(context.Context) error {
	panic("Unimplemented method")
}

func (srv ArticleServer) Create(_ context.Context, _ *pb.CreateArticleRequest) (*pb.CreateArticleResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (srv ArticleServer) Update(_ context.Context, _ *pb.UpdateArticleRequest) (*pb.UpdateArticleResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (srv ArticleServer) Get(_ context.Context, _ *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	panic("not implemented") // TODO: Implement
}

func (srv ArticleServer) GetStream(_ *pb.GetArticleRequest, _ pb.ArticleService_GetStreamServer) error {
	panic("not implemented") // TODO: Implement
}
