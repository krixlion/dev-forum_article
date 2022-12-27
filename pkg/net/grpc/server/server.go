package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gofrs/uuid"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/logging"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/pb"
	"github.com/krixlion/dev-forum_article/pkg/net/rabbitmq"
	"github.com/krixlion/dev-forum_article/pkg/storage"
	"github.com/krixlion/dev-forum_article/pkg/storage/cmd"
	"github.com/krixlion/dev-forum_article/pkg/storage/query"
	"github.com/krixlion/dev-forum_article/pkg/tracing"
	"golang.org/x/sync/errgroup"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ArticleServer struct {
	pb.UnimplementedArticleServiceServer
	storage      storage.Storage
	eventHandler event.Handler
	logger       logging.Logger
}

// MakeArticleServer reads connection data from the environment
// using os.Getenv() and loads it to the conn structs.
func MakeArticleServer() ArticleServer {
	cmd_port := os.Getenv("DB_WRITE_PORT")
	cmd_host := os.Getenv("DB_WRITE_HOST")
	cmd_user := os.Getenv("DB_WRITE_USER")
	cmd_pass := os.Getenv("DB_WRITE_PASS")

	query_port := os.Getenv("DB_READ_PORT")
	query_host := os.Getenv("DB_READ_HOST")
	query_pass := os.Getenv("DB_READ_PASS")

	mq_port := os.Getenv("MQ_PORT")
	mq_host := os.Getenv("MQ_HOST")
	mq_user := os.Getenv("MQ_USER")
	mq_pass := os.Getenv("MQ_PASS")

	consumer := tracing.ServiceName

	config := rabbitmq.Config{
		QueueSize:         100,
		MaxWorkers:        100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	logger, _ := logging.NewLogger()
	cmd := cmd.MakeDB(cmd_port, cmd_host, cmd_user, cmd_pass)
	query, err := query.MakeDB(query_host, query_port, query_pass)

	if err != nil {
		panic(err)
	}

	return ArticleServer{
		storage:      storage.NewStorage(cmd, query, logger),
		logger:       logger,
		eventHandler: rabbitmq.NewRabbitMQ(consumer, mq_user, mq_pass, mq_host, mq_port, config),
	}
}

func (s ArticleServer) Run(ctx context.Context) error {
	errg, ctx := errgroup.WithContext(ctx)
	errg.Go(func() error {
		return s.eventHandler.Run()
	})

	errg.Go(func() error {
		return s.storage.ListenAndCatchUp(ctx)
	})

	return errg.Wait()
}

func (s ArticleServer) Close() error {
	var errMsg string
	err := s.eventHandler.Close()
	if err != nil {
		errMsg = fmt.Sprintf("failed to close eventHandler: %s", err)
	}

	err = s.storage.Close()
	if err != nil {
		errMsg = fmt.Sprintf("%s, failed to close storage: %s", errMsg, err)
	}

	if errMsg != "" {
		return errors.New(errMsg)
	}

	return nil
}

func (s ArticleServer) Create(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleResponse, error) {
	article := entity.ArticleFromPb(req.GetArticle())
	id, err := uuid.NewV4()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	// Assign new UUID to article about to be created.
	article.Id = id.String()

	err = s.storage.Create(ctx, article)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	json, err := json.Marshal(article)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	event := event.Event{
		Entity:    entity.ArticleEntity,
		Type:      event.Created,
		Body:      json,
		Timestamp: time.Now(),
	}

	if err := s.eventHandler.ResilientPublish(ctx, event); err != nil {
		s.logger.Log(ctx, "Failed to publish event", "err", err)
	}

	return &pb.CreateArticleResponse{
		Id: id.String(),
	}, nil
}

func (s ArticleServer) Delete(ctx context.Context, req *pb.DeleteArticleRequest) (*pb.DeleteArticleResponse, error) {
	err := s.storage.Delete(ctx, req.GetId())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	json, err := json.Marshal(req.GetId())
	if err != nil {
		return nil, err
	}

	event := event.Event{
		Entity:    entity.ArticleEntity,
		Type:      event.Deleted,
		Body:      json,
		Timestamp: time.Now(),
	}

	if err := s.eventHandler.ResilientPublish(ctx, event); err != nil {
		s.logger.Log(ctx, "Failed to publish event", "err", err)
	}

	return &pb.DeleteArticleResponse{}, nil
}

func (s ArticleServer) Update(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleResponse, error) {
	article := entity.ArticleFromPb(req.GetArticle())

	err := s.storage.Update(ctx, article)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	json, err := json.Marshal(article)
	if err != nil {
		return nil, err
	}

	event := event.Event{
		Entity:    entity.ArticleEntity,
		Type:      event.Updated,
		Body:      json,
		Timestamp: time.Now(),
	}

	if err := s.eventHandler.ResilientPublish(ctx, event); err != nil {
		s.logger.Log(ctx, "Failed to publish event", "err", err)
	}

	return &pb.UpdateArticleResponse{}, nil
}

func (s ArticleServer) Get(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	article, err := s.storage.Get(ctx, req.GetArticleId())

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
