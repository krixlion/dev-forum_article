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
	"github.com/krixlion/dev_forum-auth/pkg/grpc/auth"
	authPb "github.com/krixlion/dev_forum-auth/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-auth/pkg/tokens"
	"github.com/krixlion/dev_forum-auth/pkg/tokens/validator"
	"github.com/krixlion/dev_forum-lib/cert"
	"github.com/krixlion/dev_forum-lib/env"
	"github.com/krixlion/dev_forum-lib/event/broker"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-lib/tracing"
	rabbitmq "github.com/krixlion/dev_forum-rabbitmq"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

// Hardcoded root dir name.
const projectDir = "app"
const serviceName = "article-service"

var port int
var isTLS bool

func init() {
	portFlag := flag.Int("p", 50051, "The gRPC server port")
	insecureFlag := flag.Bool("insecure", false, "Whether to not use TLS over gRPC")
	flag.Parse()
	port = *portFlag
	isTLS = !(*insecureFlag)
}

func main() {
	env.Load(projectDir)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	deps, err := getServiceDependencies(ctx, serviceName, isTLS)
	if err != nil {
		logging.Log("Failed to initialize service dependencies", "err", err)
		return
	}
	service := service.NewArticleService(port, deps)
	service.Run(ctx)

	<-ctx.Done()
	logging.Log("Service shutting down")

	defer func() {
		cancel()

		if err := service.Close(); err != nil {
			logging.Log("Failed to shutdown service", "err", err)
			return
		}

		logging.Log("Service shutdown successful")
	}()
}

// getServiceDependencies is a Composition root.
// Panics on any non-nil error.
func getServiceDependencies(ctx context.Context, serviceName string, isTLS bool) (service.Dependencies, error) {
	clientCreds := insecure.NewCredentials()
	serverCreds := insecure.NewCredentials()

	if isTLS {
		caCertPool, err := cert.LoadCaPool(os.Getenv("TLS_CA_PATH"))
		if err != nil {
			return service.Dependencies{}, err
		}

		clientCreds = credentials.NewClientTLSFromCert(caCertPool, "")

		serverCert, err := cert.LoadX509KeyPair(os.Getenv("TLS_CERT_PATH"), os.Getenv("TLS_KEY_PATH"))
		if err != nil {
			return service.Dependencies{}, err
		}

		serverCreds = credentials.NewServerTLSFromCert(&serverCert)
	}

	shutdownTracing, err := tracing.InitProvider(ctx, serviceName, os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if err != nil {
		return service.Dependencies{}, err
	}

	tracer := otel.Tracer(serviceName)

	logger, err := logging.NewLogger()
	if err != nil {
		return service.Dependencies{}, err
	}

	cmd, err := eventstore.MakeDB(os.Getenv("DB_WRITE_PORT"), os.Getenv("DB_WRITE_HOST"), os.Getenv("DB_WRITE_USER"), os.Getenv("DB_WRITE_PASS"), logger, tracer)
	if err != nil {
		return service.Dependencies{}, err
	}

	query, err := redis.MakeDB(os.Getenv("DB_READ_HOST"), os.Getenv("DB_READ_PORT"), os.Getenv("DB_READ_PASS"), logger, tracer)
	if err != nil {
		return service.Dependencies{}, err
	}

	mqConfig := rabbitmq.Config{
		QueueSize:         100,
		MaxWorkers:        100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	messageQueue := rabbitmq.NewRabbitMQ(serviceName, os.Getenv("MQ_USER"), os.Getenv("MQ_PASS"), os.Getenv("MQ_HOST"), os.Getenv("MQ_PORT"), mqConfig,
		rabbitmq.WithLogger(logger),
		rabbitmq.WithTracer(tracer),
	)
	broker := broker.NewBroker(messageQueue, logger, tracer)
	dispatcher := dispatcher.NewDispatcher(20)
	dispatcher.Register(query)

	userConn, err := grpc.NewClient(os.Getenv("USER_SERVICE_SERVICE_HOST")+":"+os.Getenv("USER_SERVICE_SERVICE_PORT"),
		grpc.WithTransportCredentials(clientCreds),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return service.Dependencies{}, err
	}
	userClient := userPb.NewUserServiceClient(userConn)

	authConn, err := grpc.NewClient(os.Getenv("AUTH_SERVICE_SERVICE_HOST")+":"+os.Getenv("AUTH_SERVICE_SERVICE_PORT"),
		grpc.WithTransportCredentials(clientCreds),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return service.Dependencies{}, err
	}
	authClient := authPb.NewAuthServiceClient(authConn)

	tokenValidator, err := validator.NewValidator(tokens.DefaultIssuer, validator.DefaultRefreshFunc(authClient, tracer))
	if err != nil {
		return service.Dependencies{}, err
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

	grpcServer := grpc.NewServer(
		grpc.Creds(serverCreds),
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainStreamInterceptor(
			grpc_recovery.StreamServerInterceptor(),
		),
		grpc.ChainUnaryInterceptor(
			grpc_recovery.UnaryServerInterceptor(),
			grpc_auth.UnaryServerInterceptor(auth.NewAuthFunc(tokenValidator, tracer)),
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
			return errors.Join(userConn.Close(), authConn.Close(), articleServer.Close(), shutdownTracing(), logger.Sync())
		},
	}, nil
}
