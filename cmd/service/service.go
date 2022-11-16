package service

import (
	"github.com/krixlion/dev-forum_article/pkg/grpc/pb"
	"github.com/krixlion/dev-forum_article/pkg/log"
	"github.com/krixlion/dev-forum_article/pkg/server"

	"flag"
	"fmt"
	"net"

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
		log.PrintLn("transport", "net/tcp", "msg", "failed listening", "err", err)
	}

	grpcServer := grpc.NewServer()
	eventstore := server.Server{}

	defer func() {
		// err := eventstore.Close(context.Background())
		// if err != nil {
		// 	log.PrintLn("msg", "failed to gracefully close connections", "err", err)
		// }
	}()

	pb.RegisterArticleServiceServer(grpcServer, eventstore)

	log.PrintLn("transport", "net/tcp", "msg", "listening")
	err = grpcServer.Serve(lis)
	if err != nil {
		log.PrintLn("transport", "grpc", "msg", "failed to serve", "err", err)
	}
}
