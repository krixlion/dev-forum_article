package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-article/pkg/storage"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/krixlion/dev_forum-proto/article_service/pb"
	userPb "github.com/krixlion/dev_forum-proto/user_service/pb"
	"github.com/stretchr/testify/mock"
)

func setUpServer(ctx context.Context, db storage.CQRStorage, userClient userPb.UserServiceClient, mq event.Broker) ArticleServer {
	s := NewArticleServer(Dependencies{
		Logger: nulls.NullLogger{},
		Services: Services{
			User: userClient,
		},
		Storage:    db,
		Tracer:     nulls.NullTracer{},
		Dispatcher: dispatcher.NewDispatcher(mq, 0),
	})
	return s
}

func Test_validateCreate(t *testing.T) {
	tests := []struct {
		name       string
		storage    mocks.CQRStorage[entity.Article]
		broker     mocks.Broker
		handler    mocks.UnaryHandler
		userClient mocks.UserClient
		req        *pb.CreateArticleRequest
		want       interface{} // *pb.CreateArticleResponse
		wantErr    bool
	}{
		{
			name: "Test if validation fails on invalid userId",
			handler: func() mocks.UnaryHandler {
				m := mocks.NewUnaryHandler()
				m.On("", mock.Anything).Return().Once()
				return m
			}(),
			userClient: func() mocks.UserClient {
				m := mocks.NewUserClient()
				m.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(&userPb.GetUserResponse{}, errors.New("test err")).Once()
				return m
			}(),
			storage: func() mocks.CQRStorage[entity.Article] {
				m := mocks.NewCQRStorage[entity.Article]()
				m.On("Create", mock.Anything).Return(nil).Once()
				return m
			}(),
			broker: mocks.NewBroker(),
			req: &pb.CreateArticleRequest{
				Article: &pb.Article{
					Id:     "Id",
					Title:  "Title",
					UserId: "",
					Body:   "Body",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			s := setUpServer(ctx, tt.storage, tt.userClient, tt.broker)

			got, err := s.validateCreate(ctx, tt.req, tt.handler.GetMock())
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleServer.validateCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second), cmpopts.EquateEmpty()) {
				t.Errorf("ArticleServer.validateCreate():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_validateUpdate(t *testing.T) {
	tests := []struct {
		name       string
		storage    mocks.CQRStorage[entity.Article]
		handler    mocks.UnaryHandler
		broker     mocks.Broker
		userClient mocks.UserClient
		req        *pb.UpdateArticleRequest
		want       interface{} // *pb.UpdateArticleResponse
		wantErr    bool
	}{
		{
			name: "Test if validation fails on invalid article",
			handler: func() mocks.UnaryHandler {
				m := mocks.NewUnaryHandler()
				m.On("", mock.Anything).Return().Once()
				return m
			}(),
			userClient: mocks.NewUserClient(),
			storage: func() mocks.CQRStorage[entity.Article] {
				m := mocks.NewCQRStorage[entity.Article]()
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Article")).Return(nil).Once()
				return m
			}(),
			broker: mocks.NewBroker(),
			req: &pb.UpdateArticleRequest{
				Article: nil,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			s := setUpServer(ctx, tt.storage, tt.userClient, tt.broker)

			got, err := s.validateUpdate(ctx, tt.req, tt.handler.GetMock())
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleServer.validateUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second), cmpopts.EquateEmpty()) {
				t.Errorf("ArticleServer.validateUpdate():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_validateDelete(t *testing.T) {
	tests := []struct {
		name       string
		storage    mocks.CQRStorage[entity.Article]
		handler    mocks.UnaryHandler
		broker     mocks.Broker
		userClient mocks.UserClient
		req        *pb.DeleteArticleRequest
		wantErr    bool
	}{
		{
			name: "Test if returns OK regardless whether Article exists or not",
			handler: func() mocks.UnaryHandler {
				m := mocks.NewUnaryHandler()
				m.On("", mock.Anything).Return().Once()
				return m
			}(),
			userClient: func() mocks.UserClient {
				m := mocks.NewUserClient()
				resp := &userPb.GetUserResponse{
					User: &userPb.User{},
				}
				m.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil).Once()
				return m
			}(),
			storage: func() mocks.CQRStorage[entity.Article] {
				m := mocks.NewCQRStorage[entity.Article]()
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.Article{}, errors.New("not found")).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			req: &pb.DeleteArticleRequest{
				Id: gentest.RandomString(10),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			s := setUpServer(ctx, tt.storage, tt.userClient, tt.broker)

			_, err := s.validateDelete(ctx, tt.req, tt.handler.GetMock())
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleServer.validateDelete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
