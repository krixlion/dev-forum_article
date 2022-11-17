package service

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/krixlion/dev-forum_article/pkg/grpc/pb"
	"github.com/krixlion/dev-forum_article/pkg/grpc/server"
	"github.com/krixlion/dev-forum_article/pkg/log"

	"google.golang.org/grpc"
)

var (
	port int
)

func init() {
	portFlag := flag.Int("port", 50051, "The server port")
	flag.Parse()
	port = *portFlag
}

func Run() {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.PrintLn("transport", "grpc", "msg", "failed to create a listener", "err", err)
	}

	grpcSrv := grpc.NewServer()
	srv := server.ArticleServer{}

	defer func() {
		err := srv.Close(context.Background())
		if err != nil {
			log.PrintLn("msg", "failed to gracefully close connections", "err", err)
		}
	}()

	pb.RegisterArticleServiceServer(grpcSrv, srv)

	log.PrintLn("transport", "grpc", "msg", "listening")
	err = grpcSrv.Serve(lis)
	if err != nil {
		log.PrintLn("transport", "grpc", "msg", "failed to serve", "err", err)
	}
}
