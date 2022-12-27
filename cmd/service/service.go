package service

import (
	"context"
	"fmt"
	"net"

	"github.com/krixlion/dev-forum_article/pkg/logging"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/pb"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/server"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"

	"google.golang.org/grpc"
)

type ArticleService struct {
	grpcPort int
	grpcSrv  *grpc.Server
	srv      server.ArticleServer
	logger   logging.Logger
}

func NewArticleService(grpcPort int) *ArticleService {
	logger, _ := logging.NewLogger()
	s := &ArticleService{
		grpcPort: grpcPort,
		grpcSrv: grpc.NewServer(
			grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
			grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
		),
		srv:    server.MakeArticleServer(),
		logger: logger,
	}
	pb.RegisterArticleServiceServer(s.grpcSrv, s.srv)
	return s
}

func (service *ArticleService) Run(ctx context.Context) {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", service.grpcPort))
	if err != nil {
		service.logger.Log(ctx, "failed to create a listener", "transport", "grpc", "err", err)
	}

	go func() {
		if err := service.srv.Run(ctx); err != nil {
			service.logger.Log(ctx, "Failed to run server", "err", err)
		}
	}()

	service.logger.Log(ctx, "listening", "transport", "grpc", "port", service.grpcPort)
	err = service.grpcSrv.Serve(lis)
	if err != nil {
		service.logger.Log(ctx, "failed to serve", "transport", "grpc", "err", err)
	}
}

func (s *ArticleService) Close() error {
	s.grpcSrv.GracefulStop()
	return s.srv.Close()
}
