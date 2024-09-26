package server

import (
	"context"
	"fmt"
	"html"
	"slices"

	"github.com/gofrs/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	pb "github.com/krixlion/dev_forum-article/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-lib/tracing"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthMatcher returns a callback used to determine whether current gRPC path should be protected by auth middleware.
func (ArticleServer) AuthMatcher() selector.Matcher {
	return selector.MatchFunc(func(ctx context.Context, callMeta interceptors.CallMeta) bool {
		// List of paths excluded from auth middleware.
		disabledAuthPaths := []string{
			"/article.ArticleService/GetStream",
			"/article.ArticleService/Get",
		}
		return !slices.Contains(disabledAuthPaths, callMeta.FullMethod())
	})
}

func (s ArticleServer) ValidateRequestInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		switch info.FullMethod {
		case "/article.ArticleService/Create":
			return s.validateCreate(ctx, req.(*pb.CreateArticleRequest), handler)
		case "/article.ArticleService/Update":
			return s.validateUpdate(ctx, req.(*pb.UpdateArticleRequest), handler)
		case "/article.ArticleService/Delete":
			return s.validateDelete(ctx, req.(*pb.DeleteArticleRequest), handler)
		default:
			return handler(ctx, req)
		}
	}
}

func (server ArticleServer) validateCreate(ctx context.Context, req *pb.CreateArticleRequest, handler grpc.UnaryHandler) (_ interface{}, err error) {
	ctx, span := server.tracer.Start(ctx, "server.validateCreate", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	article := req.GetArticle()

	if article == nil {
		return nil, status.Error(codes.FailedPrecondition, "Article not provided")
	}

	// Sanitize user input.
	// Assign a new ID: do not let users assign custom ID to articles.
	id, err := uuid.NewV4()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	article.Id = id.String()

	// Escape html content.
	article.Body = html.EscapeString(article.GetBody())
	article.Title = html.EscapeString(article.GetTitle())

	userResp, err := server.services.User.Get(ctx, &userPb.GetUserRequest{Id: article.GetUserId()})
	if err != nil {
		return nil, fmt.Errorf("failed to verify user id: %w", err)
	}

	if userResp.GetUser().GetId() == "" {
		return nil, status.Error(codes.FailedPrecondition, "User with provided ID does not exist")
	}

	return handler(ctx, req)
}

func (server ArticleServer) validateUpdate(ctx context.Context, req *pb.UpdateArticleRequest, handler grpc.UnaryHandler) (_ interface{}, err error) {
	ctx, span := server.tracer.Start(ctx, "server.validateUpdate")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	article := req.GetArticle()

	if article == nil {
		return nil, status.Error(codes.FailedPrecondition, "Article not provided")
	}

	// Sanitize user input.
	// It is not allowed to change article ownership.
	article.UserId = ""

	// Escape html content.
	article.Body = html.EscapeString(article.GetBody())
	article.Title = html.EscapeString(article.GetTitle())

	return handler(ctx, req)
}

func (server ArticleServer) validateDelete(ctx context.Context, req *pb.DeleteArticleRequest, handler grpc.UnaryHandler) (interface{}, error) {
	ctx, span := server.tracer.Start(ctx, "server.validateDelete")
	defer span.End()

	id := req.GetId()

	if id == "" {
		err := status.Error(codes.FailedPrecondition, "Article id not provided")
		tracing.SetSpanErr(span, err)
		return nil, err
	}

	if _, err := server.query.Get(ctx, id); err != nil {
		tracing.SetSpanErr(span, err)
		// Do not let the user know whether the entity with provided ID existed before deleting or not.
		return nil, nil
	}

	return handler(ctx, req)
}
