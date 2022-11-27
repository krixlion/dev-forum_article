package service

import (
	"context"
	"flag"
	"fmt"
	"net"

	"github.com/krixlion/dev-forum_article/pkg/log"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/pb"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/server"

	"google.golang.org/grpc"
)

var (
	port int
)

func init() {
	portFlag := flag.Int("p", 50051, "The gRPC server port")
	flag.Parse()
	port = *portFlag
}

func Run() {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.PrintLn("transport", "grpc", "msg", "failed to create a listener", "err", err)
	}

	grpcSrv := grpc.NewServer()
	srv := server.MakeArticleServer()

	defer func() {
		err := srv.Close(context.Background())
		if err != nil {
			log.PrintLn("msg", "failed to gracefully close connections", "err", err)
		}
	}()

	pb.RegisterArticleServiceServer(grpcSrv, srv)

	log.PrintLn("transport", "grpc", "msg", "listening", "port", port)
	err = grpcSrv.Serve(lis)
	if err != nil {
		log.PrintLn("transport", "grpc", "msg", "failed to serve", "err", err)
	}
}

func Close() {

}
