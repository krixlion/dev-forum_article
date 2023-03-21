package mocks

import (
	"context"

	pb "github.com/krixlion/dev_forum-article/pkg/grpc/v1"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ pb.ArticleServiceClient = (*ArticleClient)(nil)

type ArticleClient struct {
	*mock.Mock
}

func NewArticleClient() ArticleClient {
	return ArticleClient{
		Mock: new(mock.Mock),
	}
}
func (m ArticleClient) Create(ctx context.Context, in *pb.CreateArticleRequest, opts ...grpc.CallOption) (*pb.CreateArticleResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.CreateArticleResponse), args.Error(1)
}

func (m ArticleClient) Update(ctx context.Context, in *pb.UpdateArticleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m ArticleClient) Delete(ctx context.Context, in *pb.DeleteArticleRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*emptypb.Empty), args.Error(1)
}

func (m ArticleClient) Get(ctx context.Context, in *pb.GetArticleRequest, opts ...grpc.CallOption) (*pb.GetArticleResponse, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(*pb.GetArticleResponse), args.Error(1)
}

func (m ArticleClient) GetStream(ctx context.Context, in *pb.GetArticlesRequest, opts ...grpc.CallOption) (pb.ArticleService_GetStreamClient, error) {
	args := m.Called(ctx, in, opts)
	return args.Get(0).(pb.ArticleService_GetStreamClient), args.Error(1)
}
