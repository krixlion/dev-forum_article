package server

import (
	"context"

	"github.com/krixlion/dev_forum-proto/article_service/pb"
	"google.golang.org/grpc"
)

func (s ArticleServer) ValidateRequestInterceptor() grpc.UnaryServerInterceptor {

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		switch info.FullMethod {
		case "/ArticleService/Create":
			return s.validateCreate(ctx, req.(*pb.CreateArticleRequest), handler)
		case "/ArticleService/Update":
			return s.validateUpdate(ctx, req.(*pb.UpdateArticleRequest), handler)
		case "/ArticleService/Delete":
			return s.validateDelete(ctx, req.(*pb.DeleteArticleRequest), handler)
		default:
			return handler(ctx, req)
		}
	}
}

func (s ArticleServer) validateCreate(ctx context.Context, req *pb.CreateArticleRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (s ArticleServer) validateUpdate(ctx context.Context, req *pb.UpdateArticleRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (s ArticleServer) validateDelete(ctx context.Context, req *pb.DeleteArticleRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}
