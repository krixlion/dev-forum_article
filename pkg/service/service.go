package service

import (
	"context"
	"fmt"
	"net"

	"github.com/krixlion/dev_forum-article/pkg/grpc/server"
	"github.com/krixlion/dev_forum-article/pkg/storage"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"
	"github.com/krixlion/dev_forum-proto/article_service/pb"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

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
	dispatcher      *dispatcher.Dispatcher
	logger          logging.Logger
}

type Dependencies struct {
	Logger     logging.Logger
	Broker     event.Broker
	SyncEvents event.Consumer
	Storage    storage.CQRStorage
	Dispatcher *dispatcher.Dispatcher
}

func NewArticleService(grpcPort int, d Dependencies) *ArticleService {
	srv := server.ArticleServer{
		Storage:    d.Storage,
		Logger:     d.Logger,
		Dispatcher: d.Dispatcher,
	}

	baseSrv := grpc.NewServer(
		grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
		grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
	)

	s := &ArticleService{
		grpcPort:        grpcPort,
		grpcSrv:         baseSrv,
		srv:             srv,
		dispatcher:      d.Dispatcher,
		broker:          d.Broker,
		syncEventSource: d.SyncEvents,
		logger:          d.Logger,
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
		s.dispatcher.AddEventProviders(s.eventProviders(ctx)...)
		s.dispatcher.AddSyncEventProviders(s.syncEventProviders(ctx)...)
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

func (s *ArticleService) syncEventProviders(ctx context.Context) (chans []<-chan event.Event) {

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

func (s *ArticleService) eventProviders(ctx context.Context) (chans []<-chan event.Event) {
	uDeleted, err := s.broker.Consume(ctx, "deleteArticlesAfterUserDeleted", event.UserDeleted)
	if err != nil {
		panic(err)
	}

	return append(chans, uDeleted)
}
