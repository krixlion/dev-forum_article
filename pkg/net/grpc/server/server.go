package server

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/log"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/pb"
	"github.com/krixlion/dev-forum_article/pkg/storage"
	"github.com/krixlion/dev-forum_article/pkg/storage/cmd"
	"github.com/krixlion/dev-forum_article/pkg/storage/query"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ArticleServer struct {
	pb.UnimplementedArticleServiceServer
	repository   storage.Storage
	eventHandler event.Handler
	logger       log.Logger
}

// MakeArticleServer reads DB connection data from
// the environment using os.Getenv() and loads it to the DB struct.
func MakeArticleServer() ArticleServer {
	cmd_port := os.Getenv("DB_WRITE_PORT")
	cmd_host := os.Getenv("DB_WRITE_HOST")
	cmd_user := os.Getenv("DB_WRITE_USER")
	cmd_pass := os.Getenv("DB_WRITE_PASS")

	query_port := os.Getenv("DB_READ_PORT")
	query_host := os.Getenv("DB_READ_HOST")
	query_pass := os.Getenv("DB_READ_PASS")

	logger, _ := log.NewLogger()
	cmd := cmd.MakeDB(cmd_port, cmd_host, cmd_user, cmd_pass)
	query := query.MakeDB(query_host, query_port, query_pass)

	return ArticleServer{
		repository: storage.NewStorage(cmd, query, logger),
		logger:     logger,
	}
}

func (s ArticleServer) Close() error {
	return s.repository.Close()
}

func (s ArticleServer) Create(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleResponse, error) {
	article := entity.MakeArticleFromPb(req.GetArticle())
	err := s.repository.Create(ctx, article)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	json, err := json.Marshal(article)
	if err != nil {
		return nil, err
	}

	event := event.Event{
		Entity:    entity.ArticleEntityName,
		Type:      event.ArticleCreated,
		Body:      json,
		Timestamp: time.Now(),
	}

	s.eventHandler.ResilientPublish(ctx, event)

	return &pb.CreateArticleResponse{
		IsSuccess: true,
	}, nil
}

func (s ArticleServer) Delete(ctx context.Context, req *pb.DeleteArticleRequest) (*pb.DeleteArticleResponse, error) {
	err := s.repository.Delete(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	json, err := json.Marshal(req.GetId())
	if err != nil {
		return nil, err
	}

	event := event.Event{
		Entity:    entity.ArticleEntityName,
		Type:      event.ArticleDeleted,
		Body:      json,
		Timestamp: time.Now(),
	}

	s.eventHandler.ResilientPublish(ctx, event)

	return &pb.DeleteArticleResponse{
		IsSuccess: true,
	}, nil
}

func (s ArticleServer) Update(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleResponse, error) {
	article := entity.MakeArticleFromPb(req.GetArticle())

	err := s.repository.Update(ctx, article)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	json, err := json.Marshal(article)
	if err != nil {
		return nil, err
	}

	event := event.Event{
		Entity:    entity.ArticleEntityName,
		Type:      event.ArticleUpdated,
		Body:      json,
		Timestamp: time.Now(),
	}

	s.eventHandler.ResilientPublish(ctx, event)

	return &pb.UpdateArticleResponse{
		IsSuccess: true,
	}, nil
}

func (s ArticleServer) Get(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	article, err := s.repository.Get(ctx, req.GetArticleId())

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get article: %v", err)
	}

	return &pb.GetArticleResponse{
		Article: &pb.Article{
			Id:     article.Id,
			UserId: article.UserId,
			Title:  article.Title,
			Body:   article.Body,
		},
	}, err
}

func (s ArticleServer) GetStream(req *pb.GetArticlesRequest, stream pb.ArticleService_GetStreamServer) error {
	ctx := stream.Context()
	articles, err := s.repository.GetMultiple(ctx, req.GetOffset(), req.GetLimit())
	if err != nil {
		return err
	}

	for _, v := range articles {
		select {
		case <-ctx.Done():
			return nil
		default:
			article := pb.Article{
				Id:     v.Id,
				UserId: v.UserId,
				Title:  v.Title,
				Body:   v.Body,
			}

			err := stream.Send(&article)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
