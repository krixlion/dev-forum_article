package service

import (
	"context"
	"fmt"
	"net"

	"github.com/krixlion/dev_forum-article/pkg/storage"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/logging"

	"google.golang.org/grpc"
)

type ArticleService struct {
	grpcPort   int
	grpcServer *grpc.Server

	// Consumer for events used to update and sync the read model.
	syncEventSource event.Consumer
	broker          event.Broker
	dispatcher      *dispatcher.Dispatcher
	logger          logging.Logger
	shutdown        func() error
}

type Dependencies struct {
	Logger       logging.Logger
	Broker       event.Broker
	GRPCServer   *grpc.Server
	SyncEvents   event.Consumer
	Storage      storage.CQRStorage
	Dispatcher   *dispatcher.Dispatcher
	ShutdownFunc func() error
}

func NewArticleService(grpcPort int, d Dependencies) *ArticleService {
	s := &ArticleService{
		grpcPort:        grpcPort,
		dispatcher:      d.Dispatcher,
		grpcServer:      d.GRPCServer,
		broker:          d.Broker,
		syncEventSource: d.SyncEvents,
		logger:          d.Logger,
		shutdown:        d.ShutdownFunc,
	}

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
	if err = s.grpcServer.Serve(lis); err != nil {
		s.logger.Log(ctx, "failed to serve", "transport", "grpc", "err", err)
	}
}

func (s *ArticleService) Close() error {
	return s.shutdown()
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
