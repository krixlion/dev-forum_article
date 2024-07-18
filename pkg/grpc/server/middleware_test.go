package server

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/krixlion/dev_forum-article/internal/gentest"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	pb "github.com/krixlion/dev_forum-article/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-article/pkg/storage"
	"github.com/krixlion/dev_forum-article/pkg/storage/storagemocks"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	userMock "github.com/krixlion/dev_forum-user/pkg/grpc/mocks"
	userPb "github.com/krixlion/dev_forum-user/pkg/grpc/v1"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/emptypb"
)

func setUpServer(getter storage.Getter, writer storage.Writer, userClient userPb.UserServiceClient, mq event.Broker) ArticleServer {
	s := MakeArticleServer(Dependencies{
		Services: Services{
			User: userClient,
		},
		Query:      getter,
		Cmd:        writer,
		Broker:     mq,
		Tracer:     nulls.NullTracer{},
		Dispatcher: dispatcher.NewDispatcher(0),
	})
	return s
}

func TestArticleServer_AuthMatcher(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   bool
	}{
		{
			name:   "Test returns false for Get method",
			method: "Get",
			want:   false,
		},
		{
			name:   "Test returns false for Create method",
			method: "GetStream",
			want:   false,
		},
		{
			name:   "Test returns true for Update method",
			method: "Update",
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			callMeta := interceptors.CallMeta{Method: tt.method, Service: "article.ArticleService"}
			got := ArticleServer{}.AuthMatcher().Match(ctx, callMeta)
			if got != tt.want {
				t.Errorf("ArticleServer.AuthMatcher():\n got = %v\n want = %v\n", got, tt.want)
			}
		})
	}
}

func Test_validateCreate(t *testing.T) {
	tests := []struct {
		name       string
		cmd        storagemocks.Writer
		broker     mocks.Broker
		handler    mocks.UnaryHandler
		userClient userMock.UserClient
		req        *pb.CreateArticleRequest
		want       *pb.CreateArticleResponse
		wantErr    bool
	}{
		{
			name: "Test if validation fails on invalid userId",
			handler: func() mocks.UnaryHandler {
				m := mocks.NewUnaryHandler()
				m.On("", mock.Anything).Return().Once()
				return m
			}(),
			userClient: func() userMock.UserClient {
				m := userMock.NewUserClient()
				m.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(&userPb.GetUserResponse{}, errors.New("test err")).Once()
				return m
			}(),
			cmd: func() storagemocks.Writer {
				m := storagemocks.NewWriter()
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

			s := setUpServer(storagemocks.Getter{}, tt.cmd, tt.userClient, tt.broker)

			got, err := s.validateCreate(ctx, tt.req, tt.handler.GetMock())
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleServer.validateCreate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second), cmpopts.EquateEmpty()) && !tt.wantErr {
				t.Errorf("ArticleServer.validateCreate():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_validateUpdate(t *testing.T) {
	tests := []struct {
		name       string
		query      storagemocks.Getter
		handler    mocks.UnaryHandler
		broker     mocks.Broker
		userClient userMock.UserClient
		req        *pb.UpdateArticleRequest
		want       *emptypb.Empty
		wantErr    bool
	}{
		{
			name: "Test if validation fails on nil article",
			handler: func() mocks.UnaryHandler {
				m := mocks.NewUnaryHandler()
				return m
			}(),
			userClient: userMock.NewUserClient(),
			query: func() storagemocks.Getter {
				m := storagemocks.NewGetter()
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

			s := setUpServer(tt.query, storagemocks.Writer{}, tt.userClient, tt.broker)

			got, err := s.validateUpdate(ctx, tt.req, tt.handler.GetMock())
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleServer.validateUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.EquateApproxTime(time.Second)) && !tt.wantErr {
				t.Errorf("ArticleServer.validateUpdate():\n got = %v\n want = %v\n %v", got, tt.want, cmp.Diff(got, tt.want))
			}
		})
	}
}

func Test_validateDelete(t *testing.T) {
	tests := []struct {
		name       string
		query      storagemocks.Getter
		handler    mocks.UnaryHandler
		broker     mocks.Broker
		userClient userMock.UserClient
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
			userClient: func() userMock.UserClient {
				m := userMock.NewUserClient()
				resp := &userPb.GetUserResponse{
					User: &userPb.User{},
				}
				m.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(resp, nil).Once()
				return m
			}(),
			query: func() storagemocks.Getter {
				m := storagemocks.NewGetter()
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

			s := setUpServer(tt.query, storagemocks.Writer{}, tt.userClient, tt.broker)

			_, err := s.validateDelete(ctx, tt.req, tt.handler.GetMock())
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleServer.validateDelete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
