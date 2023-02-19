package server_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/grpc/server"
	"github.com/krixlion/dev_forum-article/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-article/pkg/storage"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"
	"github.com/krixlion/dev_forum-proto/article_service/pb"
	userPb "github.com/krixlion/dev_forum-proto/user_service/pb"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// setUpServerWithMiddleware initializes and runs in the background a gRPC
// server allowing only for local calls for testing.
// Returns a client to interact with the server. The server is shutdown when ctx.Done() receives.
// Bufconn allows the server to call itself.
func setUpServerWithMiddleware(ctx context.Context, db storage.CQRStorage, userClient userPb.UserServiceClient, mq event.Broker) pb.ArticleServiceClient {
	lis := bufconn.Listen(1024 * 1024)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	server := server.NewArticleServer(server.Dependencies{
		Logger:     nulls.NullLogger{},
		UserClient: userClient,
		Storage:    db,
		Tracer:     nulls.NullTracer{},
		Dispatcher: dispatcher.NewDispatcher(mq, 0),
	})

	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			server.ValidateRequestInterceptor(),
		),
	)

	pb.RegisterArticleServiceServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited w mith error: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		s.Stop()
	}()

	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := pb.NewArticleServiceClient(conn)
	return client
}

func Test_validateCreate(t *testing.T) {
	tests := []struct {
		name       string
		storage    mocks.CQRStorage[entity.Article]
		broker     mocks.Broker
		userClient mocks.UserClient
		req        *pb.CreateArticleRequest
		want       *pb.CreateArticleResponse
		wantErr    bool
	}{
		{
			name: "Test if validation fails on invalid userId",
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
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
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

			s := setUpServerWithMiddleware(ctx, tt.storage, tt.userClient, tt.broker)

			got, err := s.Create(ctx, tt.req)
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
		broker     mocks.Broker
		userClient mocks.UserClient
		req        *pb.UpdateArticleRequest
		want       *pb.UpdateArticleResponse
		wantErr    bool
	}{
		{
			name:       "Test if validation fails on invalid article",
			userClient: mocks.NewUserClient(),
			storage: func() mocks.CQRStorage[entity.Article] {
				m := mocks.NewCQRStorage[entity.Article]()
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Article")).Return(nil).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
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

			s := setUpServerWithMiddleware(ctx, tt.storage, tt.userClient, tt.broker)

			got, err := s.Update(ctx, tt.req)
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
		broker     mocks.Broker
		userClient mocks.UserClient
		req        *pb.DeleteArticleRequest
		wantErr    bool
	}{
		{
			name: "Test if returns OK regardless whether Article exists or not",
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

			s := setUpServerWithMiddleware(ctx, tt.storage, tt.userClient, tt.broker)

			_, err := s.Delete(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ArticleServer.validateDelete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
