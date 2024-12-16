package server

import (
	"context"
	"time"

	pb "github.com/krixlion/dev_forum-article/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-article/pkg/storage"
	authPb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/tracing"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"go.opentelemetry.io/otel/trace"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ArticleServer struct {
	pb.UnimplementedArticleServiceServer
	services    Services
	query       storage.Getter
	cmd         storage.Writer
	dispatcher  *dispatcher.Dispatcher
	broker      event.Broker
	tokenParser tokens.Parser
	tracer      trace.Tracer
}

type Dependencies struct {
	Services   Services
	Query      storage.Getter
	Cmd        storage.Writer
	Dispatcher *dispatcher.Dispatcher
	Broker     event.Broker
	Parser     tokens.Parser
	Tracer     trace.Tracer
}

type Services struct {
	User userPb.UserServiceClient
	Auth authPb.AuthServiceClient
}

func MakeArticleServer(d Dependencies) ArticleServer {
	return ArticleServer{
		services:    d.Services,
		query:       d.Query,
		cmd:         d.Cmd,
		broker:      d.Broker,
		dispatcher:  d.Dispatcher,
		tokenParser: d.Parser,
		tracer:      d.Tracer,
	}
}

func (s ArticleServer) Create(ctx context.Context, req *pb.CreateArticleRequest) (_ *pb.CreateArticleResponse, err error) {
	ctx, span := s.tracer.Start(ctx, "server.Create")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	article := articleFromPB(req.GetArticle())

	if err := s.cmd.Create(ctx, article); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create article: %v", err.Error())
	}

	event, err := event.MakeEvent(event.ArticleAggregate, event.ArticleCreated, article, tracing.ExtractMetadataFromContext(ctx))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to make article-created event: %v", err.Error())
	}

	if err := s.broker.ResilientPublish(event); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to publish the article-created event: %v", err.Error())
	}

	return &pb.CreateArticleResponse{
		Id: article.Id,
	}, nil
}

func (s ArticleServer) Delete(ctx context.Context, req *pb.DeleteArticleRequest) (_ *emptypb.Empty, err error) {
	ctx, span := s.tracer.Start(ctx, "server.Delete")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	id := req.GetId()

	if err := s.cmd.Delete(ctx, id); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete article: %v", err.Error())
	}

	event, err := event.MakeEvent(event.ArticleAggregate, event.ArticleDeleted, id, tracing.ExtractMetadataFromContext(ctx))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to make the article-deleted event: %v", err.Error())
	}

	if err := s.broker.ResilientPublish(event); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to publish the article-deleted event: %v", err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s ArticleServer) Update(ctx context.Context, req *pb.UpdateArticleRequest) (_ *emptypb.Empty, err error) {
	ctx, span := s.tracer.Start(ctx, "server.Update")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	article := articleFromPB(req.GetArticle())

	if err := s.cmd.Update(ctx, article); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update article: %v", err.Error())
	}

	event, err := event.MakeEvent(event.ArticleAggregate, event.ArticleUpdated, article, tracing.ExtractMetadataFromContext(ctx))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to make the article-updated event: %v", err.Error())
	}

	if err := s.broker.ResilientPublish(event); err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to publish the article-updated event: %v", err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s ArticleServer) Get(ctx context.Context, req *pb.GetArticleRequest) (_ *pb.GetArticleResponse, err error) {
	ctx, span := s.tracer.Start(ctx, "server.Get")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	article, err := s.query.Get(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get article: %v", err)
	}

	return &pb.GetArticleResponse{
		Article: &pb.Article{
			Id:        article.Id,
			UserId:    article.UserId,
			Title:     article.Title,
			Body:      article.Body,
			CreatedAt: timestamppb.New(article.CreatedAt),
			UpdatedAt: timestamppb.New(article.UpdatedAt),
		},
	}, err
}

func (s ArticleServer) GetStream(req *pb.GetArticlesRequest, stream pb.ArticleService_GetStreamServer) (err error) {
	ctx, span := s.tracer.Start(stream.Context(), "server.GetStream")
	defer span.End()
	defer tracing.SetSpanErr(span, err)

	articles, err := s.query.GetMultiple(ctx, req.GetOffset(), req.GetLimit())
	if err != nil {
		return err
	}

	for _, v := range articles {
		select {
		case <-ctx.Done():
			return nil
		default:
			err = func() error {
				_, span := s.tracer.Start(stream.Context(), "server.GetStream send")
				defer span.End()
				defer tracing.SetSpanErr(span, err)

				article := pb.Article{
					Id:        v.Id,
					UserId:    v.UserId,
					Title:     v.Title,
					Body:      v.Body,
					CreatedAt: timestamppb.New(v.CreatedAt),
					UpdatedAt: timestamppb.New(v.UpdatedAt),
				}

				return stream.Send(&article)
			}()
			if err != nil {
				return err
			}
		}

	}
	return nil
}
