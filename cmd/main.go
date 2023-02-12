package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/krixlion/dev_forum-article/pkg/grpc/server"
	"github.com/krixlion/dev_forum-article/pkg/service"
	"github.com/krixlion/dev_forum-article/pkg/storage"
	"github.com/krixlion/dev_forum-article/pkg/storage/eventstore"
	"github.com/krixlion/dev_forum-article/pkg/storage/query"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/broker"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/tracing"
	"github.com/krixlion/dev_forum-proto/article_service/pb"
	rabbitmq "github.com/krixlion/dev_forum-rabbitmq"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Hardcoded root dir name.
const projectDir = "app"
const serviceName = "article-service"

var port int

func init() {
	portFlag := flag.Int("p", 50051, "The gRPC server port")
	flag.Parse()
	port = *portFlag
}

func main() {
	env.Load(projectDir)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	shutdownTracing, err := tracing.InitProvider(ctx, serviceName)
	if err != nil {
		logging.Log("Failed to initialize tracing", "err", err)
	}

	service := service.NewArticleService(port, getServiceDependencies())
	service.Run(ctx)

	<-ctx.Done()
	logging.Log("Service shutting down")

	defer func() {
		cancel()
		shutdownTracing()
		err := service.Close()
		if err != nil {
			logging.Log("Failed to shutdown service", "err", err)
		} else {
			logging.Log("Service shutdown properly")
		}
	}()
}

// getServiceDependencies is a Composition root.
// Panics on any non-nil error.
func getServiceDependencies() service.Dependencies {
	tracer := otel.Tracer(serviceName)

	logger, err := logging.NewLogger()
	if err != nil {
		panic(err)
	}

	cmdPort := os.Getenv("DB_WRITE_PORT")
	cmdHost := os.Getenv("DB_WRITE_HOST")
	cmdUser := os.Getenv("DB_WRITE_USER")
	cmdPass := os.Getenv("DB_WRITE_PASS")
	cmd, err := eventstore.MakeDB(cmdPort, cmdHost, cmdUser, cmdPass, logger, tracer)
	if err != nil {
		panic(err)
	}

	queryPort := os.Getenv("DB_READ_PORT")
	queryHost := os.Getenv("DB_READ_HOST")
	queryPass := os.Getenv("DB_READ_PASS")
	query, err := query.MakeDB(queryHost, queryPort, queryPass, logger, tracer)
	if err != nil {
		panic(err)
	}

	mqPort := os.Getenv("MQ_PORT")
	mqHost := os.Getenv("MQ_HOST")
	mqUser := os.Getenv("MQ_USER")
	mqPass := os.Getenv("MQ_PASS")

	mqConfig := rabbitmq.Config{
		QueueSize:         100,
		MaxWorkers:        100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	storage := storage.NewCQRStorage(cmd, query, logger, tracer)

	mq := rabbitmq.NewRabbitMQ(serviceName, mqUser, mqPass, mqHost, mqPort, mqConfig, logger, tracer)
	broker := broker.NewBroker(mq, logger, tracer)
	dispatcher := dispatcher.NewDispatcher(broker, 20)
	dispatcher.SetSyncHandler(event.HandlerFunc(storage.CatchUp))

	for eType, handlers := range storage.EventHandlers() {
		dispatcher.Subscribe(eType, handlers...)
	}

	articleServer := server.NewArticleServer(storage, logger, dispatcher)

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),

		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(zap.L()),
			otelgrpc.UnaryServerInterceptor(),
			articleServer.ValidateRequestInterceptor(),
		),
	)

	reflection.Register(grpcServer)
	pb.RegisterArticleServiceServer(grpcServer, articleServer)

	return service.Dependencies{
		Logger:     logger,
		Broker:     broker,
		GRPCServer: grpcServer,
		SyncEvents: &cmd,
		Storage:    storage,
		Dispatcher: dispatcher,
		ShutdownFunc: func() error {
			grpcServer.GracefulStop()
			return articleServer.Close()
		},
	}
}
