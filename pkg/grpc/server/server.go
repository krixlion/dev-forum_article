package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	pb "github.com/krixlion/dev_forum-article/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-article/pkg/storage"
	authPb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
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
	storage        storage.CQRStorage
	dispatcher     *dispatcher.Dispatcher
	tokenValidator tokens.Validator
	logger         logging.Logger
	tracer         trace.Tracer
}

type Dependencies struct {
	Services   Services
	Storage    storage.CQRStorage
	Dispatcher *dispatcher.Dispatcher
	Validator  tokens.Validator
	Logger     logging.Logger
	Tracer     trace.Tracer
}

type Services struct {
	User userPb.UserServiceClient
	Auth authPb.AuthServiceClient
}

func NewArticleServer(d Dependencies) ArticleServer {
	return ArticleServer{
		services:       d.Services,
		storage:        d.Storage,
		dispatcher:     d.Dispatcher,
		tokenValidator: d.Validator,
		logger:         d.Logger,
		tracer:         d.Tracer,
	}
}

func (s ArticleServer) Close() error {
	var errMsg string

	err := s.storage.Close()
	if err != nil {
		errMsg = fmt.Sprintf("%s, failed to close storage: %s", errMsg, err)
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}

func (s ArticleServer) Create(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleResponse, error) {
	ctx, span := s.tracer.Start(ctx, "server.Create")
	defer span.End()

	article := articleFromPB(req.GetArticle())

	if err := s.storage.Create(ctx, article); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	event, err := event.MakeEvent(event.ArticleAggregate, event.ArticleCreated, article)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err := s.dispatcher.ResilientPublish(event); err != nil {
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

	if err := s.storage.Delete(ctx, id); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	event, err := event.MakeEvent(event.ArticleAggregate, event.ArticleDeleted, id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err := s.dispatcher.ResilientPublish(event); err != nil {
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

	if err := s.storage.Update(ctx, article); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	event, err := event.MakeEvent(event.ArticleAggregate, event.ArticleUpdated, article)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if err := s.dispatcher.ResilientPublish(event); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s ArticleServer) Get(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	ctx, span := s.tracer.Start(ctx, "server.Get")
	defer span.End()

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	article, err := s.storage.Get(ctx, req.GetId())
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

func (s ArticleServer) GetStream(req *pb.GetArticlesRequest, stream pb.ArticleService_GetStreamServer) error {
	ctx := stream.Context()
	ctx, span := s.tracer.Start(ctx, "server.GetStream")
	defer span.End()

	articles, err := s.storage.GetMultiple(ctx, req.GetOffset(), req.GetLimit())
	if err != nil {
		return err
	}

	for _, v := range articles {
		select {
		case <-ctx.Done():
			return nil
		default:
			article := pb.Article{
				Id:        v.Id,
				UserId:    v.UserId,
				Title:     v.Title,
				Body:      v.Body,
				CreatedAt: timestamppb.New(v.CreatedAt),
				UpdatedAt: timestamppb.New(v.UpdatedAt),
			}

			if err := stream.Send(&article); err != nil {
				return err
			}
		}
	}
	return nil
}
