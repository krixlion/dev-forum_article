package service

import (
	"fmt"
	"net"

	"github.com/krixlion/dev-forum_article/pkg/log"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/pb"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/server"

	"google.golang.org/grpc"
)

type ArticleService struct {
	grpcPort int
	grpcSrv  *grpc.Server
	srv      server.ArticleServer
	logger   log.Logger
}

func NewArticleService(grpcPort int) *ArticleService {
	logger, _ := log.NewLogger()
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
		service.logger.Log("failed to create a listener", "transport", "grpc", "err", err)
	}

	service.logger.Log("listening", "transport", "grpc", "port", service.grpcPort)
	err = service.grpcSrv.Serve(lis)
	if err != nil {
		service.logger.Log("failed to serve", "transport", "grpc", "err", err)
	}
}

func (s *ArticleService) Close() error {
	s.grpcSrv.GracefulStop()
	return s.srv.Close()
}
