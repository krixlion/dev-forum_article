package service

import (
	"fmt"
	"net"

	"github.com/krixlion/dev-forum_article/pkg/log"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/pb"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/server"
	"go.uber.org/zap"

	"google.golang.org/grpc"
)

type ArticleService struct {
	grpcPort int
	grpcSrv  *grpc.Server
	srv      server.ArticleServer
	logger   *zap.SugaredLogger
}

func NewArticleService(grpcPort int) *ArticleService {
	logger, _ := log.MakeZapLogger()
	s := &ArticleService{
		grpcPort: grpcPort,
		grpcSrv:  grpc.NewServer(),
		srv:      server.MakeArticleServer(),
		logger:   logger,
	}
	pb.RegisterArticleServiceServer(s.grpcSrv, s.srv)
	return s
}

func (service *ArticleService) Run() {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", service.grpcPort))
	if err != nil {
		service.logger.Infow("transport", "grpc", "msg", "failed to create a listener", "err", err)
	}

	service.logger.Infow("listening", "transport", "grpc", "port", service.grpcPort)
	err = service.grpcSrv.Serve(lis)
	if err != nil {
		service.logger.Infow("failed to serve", "transport", "grpc", "err", err)
	}
}

func (s *ArticleService) Close() error {
	s.grpcSrv.GracefulStop()
	return s.srv.Close()
}
