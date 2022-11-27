package server

import (
	"context"
	"os"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/pb"
	"github.com/krixlion/dev-forum_article/pkg/storage"
	"github.com/krixlion/dev-forum_article/pkg/storage/cmd"
	"github.com/krixlion/dev-forum_article/pkg/storage/query"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ArticleServer struct {
	pb.UnimplementedArticleServiceServer
	query storage.Reader
	cmd   storage.Writer
}

// MakeArticleServer reads DB connection data from
// the environment using os.Getenv() and loads it to the DB struct.
// This struct is then returned.
func MakeArticleServer() ArticleServer {
	cmd_port := os.Getenv("DB_WRITE_PORT")
	cmd_host := os.Getenv("DB_WRITE_HOST")
	cmd_user := os.Getenv("DB_WRITE_USER")
	cmd_pass := os.Getenv("DB_WRITE_PASS")

	query_port := os.Getenv("DB_READ_PORT")
	query_host := os.Getenv("DB_READ_HOST")
	query_user := os.Getenv("DB_READ_USER")
	query_pass := os.Getenv("DB_READ_PASS")

	return ArticleServer{
		cmd:   cmd.MakeDB(cmd_port, cmd_host, cmd_user, cmd_pass),
		query: query.MakeDB(query_port, query_host, query_user, query_pass),
	}
}

func (s ArticleServer) Close(context.Context) error {
	panic("Unimplemented Close method")
}

func (s ArticleServer) Create(ctx context.Context, req *pb.CreateArticleRequest) (*pb.CreateArticleResponse, error) {
	article := entity.MakeArticleFromPb(req.GetArticle())

	err := s.cmd.Create(ctx, article)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return nil, nil
}

func (s ArticleServer) Update(ctx context.Context, req *pb.UpdateArticleRequest) (*pb.UpdateArticleResponse, error) {
	article := entity.MakeArticleFromPb(req.GetArticle())

	err := s.cmd.Update(ctx, article)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	}

	return nil, nil
}

func (s ArticleServer) Get(ctx context.Context, req *pb.GetArticleRequest) (*pb.GetArticleResponse, error) {
	s.query.Get(ctx, req.GetArticleId())
	return nil, nil
}

func (s ArticleServer) GetStream(req *pb.GetArticlesRequest, stream pb.ArticleService_GetStreamServer) error {
	ctx := stream.Context()
	articles, err := s.query.GetMultiple(ctx, req.GetOffset(), req.GetLimit())
	if err != nil {
		return err
	}

	for _, v := range articles {
		if err := ctx.Err(); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		default:
			article := pb.Article{
				Id:     v.Id(),
				UserId: v.UserId(),
				Title:  v.Title,
				Body:   v.Body,
			}
			err := stream.Send(&article)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
