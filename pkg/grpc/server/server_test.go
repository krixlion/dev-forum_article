package server_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/grpc/server"
	pb "github.com/krixlion/dev_forum-article/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-article/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-article/pkg/storage/storagemocks"

	"github.com/krixlion/dev_forum-lib/event/dispatcher"
	"github.com/krixlion/dev_forum-lib/mocks"
	"github.com/krixlion/dev_forum-lib/nulls"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// setUpServer initializes and runs in the background a gRPC
// server allowing only for local calls for testing.
// Returns a client to interact with the server.
// The server is shutdown when ctx.Done() receives.
func setUpServer(ctx context.Context, storage storagemocks.CQRStorage, broker mocks.Broker) pb.ArticleServiceClient {
	// bufconn allows the server to call itself
	// great for testing across whole infrastructure
	lis := bufconn.Listen(1024 * 1024)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	s := grpc.NewServer()
	server := server.NewArticleServer(server.Dependencies{
		Storage:    storage,
		Dispatcher: dispatcher.NewDispatcher(broker, 2),
		Logger:     nulls.NullLogger{},
		Tracer:     nulls.NullTracer{},
	})

	pb.RegisterArticleServiceServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
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

func Test_Get(t *testing.T) {
	v := gentest.RandomArticle(2, 5)
	article := &pb.Article{
		Id:        v.Id,
		UserId:    v.UserId,
		Title:     v.Title,
		Body:      v.Body,
		CreatedAt: timestamppb.New(v.CreatedAt),
		UpdatedAt: timestamppb.New(v.UpdatedAt),
	}

	tests := []struct {
		broker  mocks.Broker
		desc    string
		arg     *pb.GetArticleRequest
		want    *pb.GetArticleResponse
		wantErr bool
		storage storagemocks.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.GetArticleRequest{
				Id: article.Id,
			},
			want: &pb.GetArticleResponse{
				Article: article,
			},
			storage: func() storagemocks.CQRStorage {
				m := storagemocks.NewCQRStorage()
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(v, nil).Times(1)
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.GetArticleRequest{
				Id: "",
			},
			want:    nil,
			wantErr: true,
			storage: func() storagemocks.CQRStorage {
				m := storagemocks.NewCQRStorage()
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.Article{}, errors.New("test err")).Times(1)
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()

			client := setUpServer(ctx, tt.storage, tt.broker)

			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			got, err := client.Get(ctx, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed to Get article, err: %v", err)
				return
			}

			// Compare in order to avoid nil pointer dereference.
			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if got != tt.want {
				if !cmp.Equal(got.Article, tt.want.Article, cmpopts.IgnoreUnexported(pb.Article{}, timestamppb.Timestamp{})) {
					t.Errorf("Articles are not equal:\n Got = %+v\n, want = %+v\n", got.Article, tt.want.Article)
					return
				}
			}
		})
	}
}

func Test_Create(t *testing.T) {
	v := gentest.RandomArticle(2, 5)
	article := &pb.Article{
		Id:     v.Id,
		UserId: v.UserId,
		Title:  v.Title,
		Body:   v.Body,
	}

	tests := []struct {
		broker   mocks.Broker
		desc     string
		arg      *pb.CreateArticleRequest
		dontWant *pb.CreateArticleResponse
		wantErr  bool
		storage  storagemocks.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.CreateArticleRequest{
				Article: article,
			},
			dontWant: &pb.CreateArticleResponse{
				Id: article.Id,
			},
			storage: func() storagemocks.CQRStorage {
				m := storagemocks.NewCQRStorage()
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Article")).Return(nil).Times(1)
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.CreateArticleRequest{
				Article: article,
			},
			dontWant: nil,
			wantErr:  true,
			storage: func() storagemocks.CQRStorage {
				m := storagemocks.NewCQRStorage()
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Article")).Return(errors.New("test err")).Times(1)
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tt.storage, tt.broker)

			got, err := client.Create(ctx, tt.arg)
			if err != nil {
				tt.broker.AssertNumberOfCalls(t, "ResilientPublish", 0)

				if !tt.wantErr {
					t.Errorf("Failed to Create article, err: %v", err)
					return
				}
			} else {
				tt.broker.AssertNumberOfCalls(t, "ResilientPublish", 1)
			}
			tt.storage.AssertNumberOfCalls(t, "Create", 1)

			// Compare in order to avoid nil pointer dereference.
			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if got != tt.dontWant {
				if _, err := uuid.FromString(got.Id); err != nil {
					t.Errorf("Article ID is not correct UUID:\n ID = %+v\n err = %+v", got.Id, err)
					return
				}
			}
		})
	}
}

func Test_Update(t *testing.T) {
	v := gentest.RandomArticle(2, 5)
	article := &pb.Article{
		Id:     v.Id,
		UserId: v.UserId,
		Title:  v.Title,
		Body:   v.Body,
	}

	tests := []struct {
		broker  mocks.Broker
		desc    string
		arg     *pb.UpdateArticleRequest
		want    *emptypb.Empty
		wantErr bool
		storage storagemocks.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.UpdateArticleRequest{
				Article: article,
			},
			want: &emptypb.Empty{},
			storage: func() storagemocks.CQRStorage {
				m := storagemocks.NewCQRStorage()
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Article")).Return(nil).Times(1)
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.UpdateArticleRequest{
				Article: article,
			},
			want:    nil,
			wantErr: true,
			storage: func() storagemocks.CQRStorage {
				m := storagemocks.NewCQRStorage()
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Article")).Return(errors.New("test err")).Times(1)
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tt.storage, tt.broker)

			got, err := client.Update(ctx, tt.arg)
			if err != nil {
				tt.broker.AssertNumberOfCalls(t, "ResilientPublish", 0)

				if !tt.wantErr {
					t.Errorf("Failed to Update article, err: %v", err)
					return
				}
			} else {
				tt.broker.AssertNumberOfCalls(t, "ResilientPublish", 1)
			}
			tt.storage.AssertNumberOfCalls(t, "Update", 1)

			// Compare in order to avoid nil pointer dereference.
			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if got != tt.want {
				if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(emptypb.Empty{})) {
					t.Errorf("Wrong response:\n got = %+v\n want = %+v\n", got, tt.want)
					return
				}
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	v := gentest.RandomArticle(2, 5)
	article := &pb.Article{
		Id:     v.Id,
		UserId: v.UserId,
		Title:  v.Title,
		Body:   v.Body,
	}

	tests := []struct {
		broker  mocks.Broker
		desc    string
		arg     *pb.DeleteArticleRequest
		want    *emptypb.Empty
		wantErr bool
		storage storagemocks.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.DeleteArticleRequest{
				Id: article.Id,
			},
			want: &emptypb.Empty{},
			storage: func() storagemocks.CQRStorage {
				m := storagemocks.NewCQRStorage()
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Times(1)
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.DeleteArticleRequest{
				Id: article.Id,
			},
			want:    nil,
			wantErr: true,
			storage: func() storagemocks.CQRStorage {
				m := storagemocks.NewCQRStorage()
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("test err")).Times(1)
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tt.storage, tt.broker)

			got, err := client.Delete(ctx, tt.arg)
			if err != nil {
				tt.broker.AssertNumberOfCalls(t, "ResilientPublish", 0)

				if !tt.wantErr {
					t.Errorf("Failed to Delete article, err: %v", err)
					return
				}
			} else {
				tt.broker.AssertNumberOfCalls(t, "ResilientPublish", 1)
			}
			tt.storage.AssertNumberOfCalls(t, "Delete", 1)

			// Compare in order to avoid nil pointer dereference.
			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if got != tt.want {
				if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(emptypb.Empty{})) {
					t.Errorf("Wrong response:\n got = %+v\n want = %+v\n", got, tt.want)
					return
				}
			}
		})
	}
}

func Test_GetStream(t *testing.T) {
	var articles []entity.Article
	for i := 0; i < 5; i++ {
		article := gentest.RandomArticle(2, 5)
		articles = append(articles, article)
	}

	var pbArticles []*pb.Article
	for _, v := range articles {
		pbArticle := &pb.Article{
			Id:        v.Id,
			UserId:    v.UserId,
			Title:     v.Title,
			Body:      v.Body,
			CreatedAt: timestamppb.New(v.CreatedAt),
			UpdatedAt: timestamppb.New(v.UpdatedAt),
		}
		pbArticles = append(pbArticles, pbArticle)
	}

	tests := []struct {
		broker  mocks.Broker
		desc    string
		arg     *pb.GetArticlesRequest
		want    []*pb.Article
		wantErr bool
		storage storagemocks.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.GetArticlesRequest{
				Offset: "0",
				Limit:  "5",
			},
			want: pbArticles,
			storage: func() storagemocks.CQRStorage {
				m := storagemocks.NewCQRStorage()
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(articles, nil).Times(1)
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
		{
			desc:    "Test if error is returned properly on storage error",
			arg:     &pb.GetArticlesRequest{},
			want:    nil,
			wantErr: true,
			storage: func() storagemocks.CQRStorage {
				m := storagemocks.NewCQRStorage()
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]entity.Article{}, errors.New("test err")).Times(1)
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tt.storage, tt.broker)

			stream, err := client.GetStream(ctx, tt.arg)
			if err != nil {
				t.Errorf("Failed to Get stream, err: %v", err)
				return
			}

			var got []*pb.Article
			for i := 0; i < len(tt.want); i++ {
				article, err := stream.Recv()
				if (err != nil) != tt.wantErr {
					t.Errorf("Failed to receive article from stream, err: %v", err)
					return
				}
				got = append(got, article)
			}

			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(pb.Article{}, timestamppb.Timestamp{})) {
				t.Errorf("Articles are not equal:\n Got = %+v\n want = %+v\n", got, tt.want)
				return
			}
		})
	}
}
