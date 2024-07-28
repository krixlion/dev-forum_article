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
	services       Services
	query          storage.Getter
	cmd            storage.Writer
	dispatcher     *dispatcher.Dispatcher
	broker         event.Broker
	tokenValidator tokens.Validator
	tracer         trace.Tracer
}

type Dependencies struct {
	Services   Services
	Query      storage.Getter
	Cmd        storage.Writer
	Dispatcher *dispatcher.Dispatcher
	Broker     event.Broker
	Validator  tokens.Validator
	Tracer     trace.Tracer
}

type Services struct {
	User userPb.UserServiceClient
	Auth authPb.AuthServiceClient
}

func MakeArticleServer(d Dependencies) ArticleServer {
	return ArticleServer{
		services:       d.Services,
		query:          d.Query,
		cmd:            d.Cmd,
		broker:         d.Broker,
		dispatcher:     d.Dispatcher,
		tokenValidator: d.Validator,
		tracer:         d.Tracer,
	}
}

func (s ArticleServer) Create(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleResponse, error) {
	ctx, span := s.tracer.Start(ctx, "server.Create")
	defer span.End()

	article := articleFromPB(req.GetArticle())

	if err := s.cmd.Create(ctx, article); err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	event, err := event.MakeEvent(event.ArticleAggregate, event.ArticleCreated, article, tracing.ExtractMetadataFromContext(ctx))
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err := s.broker.ResilientPublish(event); err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.CreateArticleResponse{
		Id: article.Id,
	}, nil
}

func (s ArticleServer) Delete(ctx context.Context, req *pb.DeleteArticleRequest) (*emptypb.Empty, error) {
	ctx, span := s.tracer.Start(ctx, "server.Delete")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	id := req.GetId()

	if err := s.cmd.Delete(ctx, id); err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	event, err := event.MakeEvent(event.ArticleAggregate, event.ArticleDeleted, id, tracing.ExtractMetadataFromContext(ctx))
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err := s.broker.ResilientPublish(event); err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s ArticleServer) Update(ctx context.Context, req *pb.UpdateArticleRequest) (*emptypb.Empty, error) {
	ctx, span := s.tracer.Start(ctx, "server.Update")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	article := articleFromPB(req.GetArticle())

	if err := s.cmd.Update(ctx, article); err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	event, err := event.MakeEvent(event.ArticleAggregate, event.ArticleUpdated, article, tracing.ExtractMetadataFromContext(ctx))
	if err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err := s.broker.ResilientPublish(event); err != nil {
		tracing.SetSpanErr(span, err)
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s ArticleServer) Get(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	ctx, span := s.tracer.Start(ctx, "server.Get")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	article, err := s.query.Get(ctx, req.GetId())
	if err != nil {
		tracing.SetSpanErr(span, err)
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

func (s ArticleServer) GetStream(req *pb.GetArticlesRequest, stream pb.ArticleService_GetStreamServer) error {
	ctx, span := s.tracer.Start(stream.Context(), "server.GetStream")
	defer span.End()

	articles, err := s.query.GetMultiple(ctx, req.GetOffset(), req.GetLimit())
	if err != nil {
		tracing.SetSpanErr(span, err)
		return err
	}

	for _, v := range articles {
		select {
		case <-ctx.Done():
			return nil
		default:
			_, span := s.tracer.Start(stream.Context(), "server.GetStream send")

			article := pb.Article{
				Id:        v.Id,
				UserId:    v.UserId,
				Title:     v.Title,
				Body:      v.Body,
				CreatedAt: timestamppb.New(v.CreatedAt),
				UpdatedAt: timestamppb.New(v.UpdatedAt),
			}

			if err := stream.Send(&article); err != nil {
				tracing.SetSpanErr(span, err)
				span.End()
				return err
			}
			span.End()
		}
	}
	return nil
}
