package service

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/Krixlion/def-forum_proto/article_service/pb"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/event/broker"
	"github.com/krixlion/dev-forum_article/pkg/logging"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/server"
	"github.com/krixlion/dev-forum_article/pkg/net/rabbitmq"
	"github.com/krixlion/dev-forum_article/pkg/storage"
	"github.com/krixlion/dev-forum_article/pkg/storage/eventstore"
	"github.com/krixlion/dev-forum_article/pkg/storage/query"
	"github.com/krixlion/dev-forum_article/pkg/tracing"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type ArticleService struct {
	grpcPort int
	grpcSrv  *grpc.Server
	srv      server.ArticleServer

	// Consumer for events used to update and sync the read model.
	syncEventSource event.Consumer
	broker          event.Broker
	dispatcher      event.Dispatcher

	logger logging.Logger
}

func NewArticleService(grpcPort int) *ArticleService {
	logger, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}

	cmd_port := os.Getenv("DB_WRITE_PORT")
	cmd_host := os.Getenv("DB_WRITE_HOST")
	cmd_user := os.Getenv("DB_WRITE_USER")
	cmd_pass := os.Getenv("DB_WRITE_PASS")
	cmd, err := eventstore.MakeDB(cmd_port, cmd_host, cmd_user, cmd_pass, logger)
	if err != nil {
		panic(err)
	}

	query_port := os.Getenv("DB_READ_PORT")
	query_host := os.Getenv("DB_READ_HOST")
	query_pass := os.Getenv("DB_READ_PASS")
	query, err := query.MakeDB(query_host, query_port, query_pass, logger)
	if err != nil {
		panic(err)
	}
	storage := storage.NewStorage(cmd, query, logger)

	mq_port := os.Getenv("MQ_PORT")
	mq_host := os.Getenv("MQ_HOST")
	mq_user := os.Getenv("MQ_USER")
	mq_pass := os.Getenv("MQ_PASS")

	consumer := tracing.ServiceName
	mqConfig := rabbitmq.Config{
		QueueSize:         100,
		MaxWorkers:        100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	tracer := otel.Tracer(tracing.ServiceName)

	mq := rabbitmq.NewRabbitMQ(consumer, mq_user, mq_pass, mq_host, mq_port, mqConfig, logger, tracer)
	broker := broker.NewBroker(mq, logger)
	dispatcher := event.MakeDispatcher()
	dispatcher.Subscribe(event.HandlerFunc(storage.CatchUp), event.ArticleCreated, event.ArticleDeleted, event.ArticleUpdated)

	srv := server.ArticleServer{
		Storage: storage,
		Logger:  logger,
	}

	baseSrv := grpc.NewServer(
		// grpc.UnaryInterceptor(srv.Interceptor),
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	s := &ArticleService{
		grpcPort: grpcPort,
		grpcSrv:  baseSrv,

		srv: srv,

		dispatcher:      dispatcher,
		broker:          broker,
		syncEventSource: &cmd,

		logger: logger,
	}
	reflection.Register(s.grpcSrv)
	pb.RegisterArticleServiceServer(s.grpcSrv, s.srv)
	return s
}

func (s *ArticleService) Run(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		return
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.grpcPort))
	if err != nil {
		s.logger.Log(ctx, "failed to create a listener", "transport", "grpc", "err", err)
	}

	go func() {
		s.dispatcher.AddEventSources(s.SyncEventSources(ctx)...)
		s.dispatcher.Run(ctx)
	}()

	s.logger.Log(ctx, "listening", "transport", "grpc", "port", s.grpcPort)
	err = s.grpcSrv.Serve(lis)
	if err != nil {
		s.logger.Log(ctx, "failed to serve", "transport", "grpc", "err", err)
	}
}

func (s *ArticleService) Close() error {
	s.grpcSrv.GracefulStop()
	return s.srv.Close()
}

func (s *ArticleService) SyncEventSources(ctx context.Context) (chans []<-chan event.Event) {

	aCreated, err := s.syncEventSource.Consume(ctx, "", event.ArticleCreated)
	if err != nil {
		panic(err)
	}

	aDeleted, err := s.syncEventSource.Consume(ctx, "", event.ArticleDeleted)
	if err != nil {
		panic(err)
	}

	aUpdated, err := s.syncEventSource.Consume(ctx, "", event.ArticleUpdated)
	if err != nil {
		panic(err)
	}

	return append(chans, aCreated, aDeleted, aUpdated)
}
