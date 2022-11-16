package server

import "github.com/krixlion/dev-forum_article/pkg/grpc/pb"

type Server struct {
	pb.UnimplementedArticleServiceServer
}
