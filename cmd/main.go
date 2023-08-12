package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/krixlion/dev_forum-article/pkg/grpc/server"
	pb "github.com/krixlion/dev_forum-article/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-article/pkg/service"
	"github.com/krixlion/dev_forum-article/pkg/storage/eventstore"
	"github.com/krixlion/dev_forum-article/pkg/storage/redis"
	authPb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-auth/pkg/tokens/validator"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/event/broker"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/tls"
	"github.com/krixlion/dev_forum-lib/tracing"
	rabbitmq "github.com/krixlion/dev_forum-rabbitmq"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
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

	service := service.NewArticleService(port, getServiceDependencies(ctx))
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
func getServiceDependencies(ctx context.Context) service.Dependencies {
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
	query, err := redis.MakeDB(queryHost, queryPort, queryPass, logger, tracer)
	if err != nil {
		panic(err)
	}

	mqPort := os.Getenv("MQ_PORT")
	mqHost := os.Getenv("MQ_HOST")
	mqUser := os.Getenv("MQ_USER")
	mqPass := os.Getenv("MQ_PASS")
	consumer := serviceName
	mqConfig := rabbitmq.Config{
		QueueSize:         100,
		MaxWorkers:        100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	messageQueue := rabbitmq.NewRabbitMQ(
		consumer,
		mqUser,
		mqPass,
		mqHost,
		mqPort,
		mqConfig,
		rabbitmq.WithLogger(logger),
		rabbitmq.WithTracer(tracer),
	)
	broker := broker.NewBroker(messageQueue, logger, tracer)
	dispatcher := dispatcher.NewDispatcher(20)
	dispatcher.Register(query)

	tlsCaPath := os.Getenv("TLS_CA_PATH")
	clientCredentials, err := tls.LoadCA(tlsCaPath)
	if err != nil {
		panic(err)
	}

	userConn, err := grpc.Dial("user-service:50051",
		grpc.WithTransportCredentials(clientCredentials),
		grpc.WithChainUnaryInterceptor(
			otelgrpc.UnaryClientInterceptor(),
		),
	)
	if err != nil {
		panic(err)
	}
	userClient := userPb.NewUserServiceClient(userConn)

	authConn, err := grpc.Dial("auth-service:50053",
		grpc.WithTransportCredentials(clientCredentials),
		grpc.WithChainUnaryInterceptor(
			otelgrpc.UnaryClientInterceptor(),
		),
	)
	if err != nil {
		panic(err)
	}
	authClient := authPb.NewAuthServiceClient(authConn)

	tokenValidator, err := validator.NewValidator("http://auth-service", validator.DefaultRefreshFunc(authClient, tracer))
	if err != nil {
		panic(err)
	}

	go tokenValidator.Run(ctx)

	articleServer := server.MakeArticleServer(server.Dependencies{
		Services: server.Services{
			User: userClient,
			Auth: authClient,
		},
		Validator:  tokenValidator,
		Query:      query,
		Cmd:        cmd,
		Dispatcher: dispatcher,
		Broker:     broker,
		Tracer:     tracer,
	})

	tlsCertPath := os.Getenv("TLS_CERT_PATH")
	tlsKeyPath := os.Getenv("TLS_KEY_PATH")
	credentials, err := tls.LoadServerCredentials(tlsCertPath, tlsKeyPath)
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(
		grpc.Creds(credentials),
		grpc.ChainStreamInterceptor(
			otelgrpc.StreamServerInterceptor(),
			grpc_recovery.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(),
			otelgrpc.UnaryServerInterceptor(),
			grpc_auth.UnaryServerInterceptor(articleServer.AuthFunc),
			articleServer.ValidateRequestInterceptor(),
		),
	)

	reflection.Register(grpcServer)
	pb.RegisterArticleServiceServer(grpcServer, articleServer)

	return service.Dependencies{
		Logger:     logger,
		Broker:     broker,
		GRPCServer: grpcServer,
		Dispatcher: dispatcher,
		ShutdownFunc: func() error {
			grpcServer.GracefulStop()

			return errors.Join(
				userConn.Close(),
				authConn.Close(),
				articleServer.Close(),
			)
		},
	}
}
