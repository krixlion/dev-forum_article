package server

import (
	"context"

	"github.com/krixlion/dev_forum-lib/tracing"
	"github.com/krixlion/dev_forum-proto/article_service/pb"
	userPb "github.com/krixlion/dev_forum-proto/user_service/pb"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	ctx, span := s.tracer.Start(ctx, "grpc.validateCreate", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	article := articleFromPB(req.GetArticle())

	userResp, err := s.userClient.Get(ctx, &userPb.GetUserRequest{Id: article.Id})
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	if userResp.GetUser().GetId() != "" {
		return handler(ctx, req)
	}

	err = status.Error(codes.FailedPrecondition, "User with provided ID does not exist.")
	tracing.SetSpanErr(span, err)
	return nil, err
}

func (s ArticleServer) validateUpdate(ctx context.Context, req *pb.UpdateArticleRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

func (s ArticleServer) validateDelete(ctx context.Context, req *pb.DeleteArticleRequest, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}
