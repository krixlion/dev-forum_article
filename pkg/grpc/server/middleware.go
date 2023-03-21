package server

import (
	"context"

	"github.com/gofrs/uuid"
	pb "github.com/krixlion/dev_forum-article/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-lib/tracing"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
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

	if req.GetArticle() == nil {
		err := status.Error(codes.FailedPrecondition, "Article not provided")
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	// Sanitize user input.
	// Assign a new ID: do not let users assign custom ID to articles.
	id, err := uuid.NewV4()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	req.GetArticle().Id = id.String()

	article := articleFromPB(req.GetArticle())
	userResp, err := s.services.User.Get(ctx, &userPb.GetUserRequest{Id: article.UserId})
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	if userResp.GetUser().GetId() == "" {
		err := status.Error(codes.FailedPrecondition, "User with provided ID does not exist")
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	return handler(ctx, req)
}

func (s ArticleServer) validateUpdate(ctx context.Context, req *pb.UpdateArticleRequest, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := s.tracer.Start(ctx, "validateUpdate")
	defer span.End()

	article := req.GetArticle()

	if article == nil {
		err := status.Error(codes.FailedPrecondition, "Article not provided")
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	// Sanitize user input.
	// It is not allowed to change article ownership.
	article.UserId = ""

	return handler(ctx, req)
}

func (s ArticleServer) validateDelete(ctx context.Context, req *pb.DeleteArticleRequest, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := s.tracer.Start(ctx, "validateDelete")
	defer span.End()

	id := req.GetId()

	if id == "" {
		err := status.Error(codes.FailedPrecondition, "Article id not provided")
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	if _, err := s.storage.Get(ctx, id); err != nil {
		tracing.SetSpanErr(span, err)
		// Do not let user whether entity with provided ID existed before deleting or not.
		return nil, nil
	}

	return handler(ctx, req)
}
