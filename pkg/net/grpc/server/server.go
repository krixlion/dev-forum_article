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
	"github.com/krixlion/dev-forum_article/pkg/event/broker"
	"github.com/krixlion/dev-forum_article/pkg/logging"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/pb"
	"github.com/krixlion/dev-forum_article/pkg/net/rabbitmq"
	"github.com/krixlion/dev-forum_article/pkg/storage"
	"github.com/krixlion/dev-forum_article/pkg/storage/eventstore"
	"github.com/krixlion/dev-forum_article/pkg/storage/query"
	"github.com/krixlion/dev-forum_article/pkg/tracing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ArticleServer struct {
	pb.UnimplementedArticleServiceServer
	storage    storage.Storage
	dispatcher event.Dispatcher
	logger     logging.Logger
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
	mq := rabbitmq.NewRabbitMQ(consumer, mq_user, mq_pass, mq_host, mq_port, logger, config)
	broker := broker.NewBroker(mq, logger)
	dispatcher := event.NewDispatcher(broker)
	cmd := eventstore.MakeDB(cmd_port, cmd_host, cmd_user, cmd_pass, logger)
	query, err := query.MakeDB(query_host, query_port, query_pass, logger)
	if err != nil {
		panic(err)
	}

	dispatcher.Subscribe(event.HandlerFunc(query.CatchUp), event.ArticleCreated, event.ArticleDeleted, event.ArticleUpdated)

	return ArticleServer{
		storage:    storage.NewStorage(cmd, query, logger),
		dispatcher: dispatcher,
		logger:     logger,
	}
}

func (s ArticleServer) Close() error {
	var errMsg string
	err := s.dispatcher.Close()
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

	json, err := json.Marshal(article)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	event := event.Event{
		AggregateId: "article",
		Type:        event.ArticleCreated,
		Body:        json,
		Timestamp:   time.Now(),
	}

	if err := s.dispatcher.Dispatch(event); err != nil {
		s.logger.Log(ctx, "Failed to dispatch event", "err", err)
	}

	if err = s.storage.Create(ctx, article); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.CreateArticleResponse{
		Id: id.String(),
	}, nil
}

func (s ArticleServer) Delete(ctx context.Context, req *pb.DeleteArticleRequest) (*pb.DeleteArticleResponse, error) {
	json, err := json.Marshal(req.GetId())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	event := event.Event{
		AggregateId: "article",
		Type:        event.ArticleDeleted,
		Body:        json,
		Timestamp:   time.Now(),
	}

	if err := s.dispatcher.Dispatch(event); err != nil {
		s.logger.Log(ctx, "Failed to dispatch event", "err", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := s.storage.Delete(ctx, req.GetId()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return &pb.DeleteArticleResponse{}, nil
}

func (s ArticleServer) Update(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleResponse, error) {
	article := entity.ArticleFromPb(req.GetArticle())
	json, err := json.Marshal(article)
	if err != nil {
		return nil, err
	}

	event := event.Event{
		AggregateId: "article",
		Type:        event.ArticleUpdated,
		Body:        json,
		Timestamp:   time.Now(),
	}

	if err := s.dispatcher.Dispatch(event); err != nil {
		s.logger.Log(ctx, "Failed to dispatch event", "err", err)
		return nil, status.Error(codes.Internal, err.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	if err := s.storage.Update(ctx, article); err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return &pb.UpdateArticleResponse{}, nil
}

func (s ArticleServer) Get(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

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
	ctx, cancel := context.WithTimeout(stream.Context(), time.Second*10)
	defer cancel()

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
