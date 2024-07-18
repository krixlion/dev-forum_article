package mocks

import (
	"context"

	pb "github.com/krixlion/dev_forum-article/pkg/grpc/v1"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"
)

var _ pb.ArticleService_GetStreamClient = (*ArticleStreamClient)(nil)

type ArticleStreamClient struct {
	*mock.Mock
}

func NewArticleStreamClient() ArticleStreamClient {
	return ArticleStreamClient{
		new(mock.Mock),
	}
}

func (m ArticleStreamClient) Recv() (*pb.Article, error) {
	returnVals := m.Called()
	return returnVals.Get(0).(*pb.Article), returnVals.Error(1)
}

func (m ArticleStreamClient) CloseSend() error {
	returnVals := m.Called()
	return returnVals.Error(0)
}

func (m ArticleStreamClient) Header() (metadata.MD, error) {
	returnVals := m.Called()
	return returnVals.Get(0).(metadata.MD), returnVals.Error(1)
}

func (m ArticleStreamClient) Trailer() metadata.MD {
	returnVals := m.Called()
	return returnVals.Get(0).(metadata.MD)
}

func (m ArticleStreamClient) Context() context.Context {
	returnVals := m.Called()
	return returnVals.Get(0).(context.Context)
}

func (m ArticleStreamClient) SendMsg(msg any) error {
	returnVals := m.Called(msg)
	return returnVals.Error(0)
}

func (m ArticleStreamClient) RecvMsg(msg any) error {
	returnVals := m.Called(msg)
	return returnVals.Error(0)
}
