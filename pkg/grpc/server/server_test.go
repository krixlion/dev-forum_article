package server_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"

	"github.com/krixlion/dev_forum-article/internal/gentest"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/grpc/server"
	pb "github.com/krixlion/dev_forum-article/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-article/pkg/storage"
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
// The server is cancel when ctx.Done() receives.
func setUpServer(ctx context.Context, query storage.Getter, cmd storage.Writer, broker mocks.Broker) pb.ArticleServiceClient {
	// bufconn allows the server to call itself
	// great for testing across whole infrastructure
	lis := bufconn.Listen(1024 * 1024)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	s := grpc.NewServer()
	server := server.MakeArticleServer(server.Dependencies{
		Query:      query,
		Cmd:        cmd,
		Broker:     broker,
		Dispatcher: dispatcher.NewDispatcher(2),
		Tracer:     nulls.NullTracer{},
	})

	pb.RegisterArticleServiceServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	conn, err := grpc.NewClient("passthrough://bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}

	go func() {
		<-ctx.Done()
		s.Stop()
		if err := conn.Close(); err != nil {
			log.Fatalf("Failed to close client conn, err: %v", err)
		}
	}()

	return pb.NewArticleServiceClient(conn)
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
		name    string
		arg     *pb.GetArticleRequest
		broker  mocks.Broker
		query   storagemocks.Getter
		want    *pb.GetArticleResponse
		wantErr bool
	}{
		{
			name: "Test if response is returned properly on simple request",
			arg: &pb.GetArticleRequest{
				Id: article.Id,
			},
			want: &pb.GetArticleResponse{
				Article: article,
			},
			query: func() storagemocks.Getter {
				m := storagemocks.NewGetter()
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(v, nil).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
		{
			name: "Test if error is returned properly on storage error",
			arg: &pb.GetArticleRequest{
				Id: "",
			},
			query: func() storagemocks.Getter {
				m := storagemocks.NewGetter()
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.Article{}, errors.New("test err")).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			client := setUpServer(ctx, tt.query, storagemocks.NewWriter(), tt.broker)

			got, err := client.Get(ctx, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed to Get article, err: %v", err)
				return
			}

			if tt.wantErr {
				return
			}

			if !cmp.Equal(got.Article, tt.want.Article, cmpopts.IgnoreUnexported(pb.Article{}, timestamppb.Timestamp{})) {
				t.Errorf("Articles are not equal:\n got = %+v\n, want = %+v\n", got.Article, tt.want.Article)
				return
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
		name    string
		arg     *pb.CreateArticleRequest
		broker  mocks.Broker
		cmd     storagemocks.Writer
		wantErr bool
	}{
		{
			name: "Test if response is returned properly on simple request",
			arg: &pb.CreateArticleRequest{
				Article: article,
			},
			cmd: func() storagemocks.Writer {
				m := storagemocks.NewWriter()
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Article")).Return(nil).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			wantErr: false,
		},
		{
			name: "Test if error is returned properly on storage error",
			arg: &pb.CreateArticleRequest{
				Article: article,
			},
			cmd: func() storagemocks.Writer {
				m := storagemocks.NewWriter()
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Article")).Return(errors.New("test err")).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			client := setUpServer(ctx, storagemocks.NewGetter(), tt.cmd, tt.broker)

			got, err := client.Create(ctx, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed to Create article, err: %v", err)
				return
			}

			if tt.wantErr {
				return
			}

			if _, err := uuid.FromString(got.Id); err != nil {
				t.Errorf("Article ID is not a correct UUID:\n got = %+v\n err = %+v", got.Id, err)
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
		name    string
		arg     *pb.UpdateArticleRequest
		broker  mocks.Broker
		cmd     storagemocks.Writer
		want    *emptypb.Empty
		wantErr bool
	}{
		{
			name: "Test if response is returned properly on simple request",
			arg: &pb.UpdateArticleRequest{
				Article: article,
			},
			cmd: func() storagemocks.Writer {
				m := storagemocks.NewWriter()
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Article")).Return(nil).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			want:    &emptypb.Empty{},
			wantErr: false,
		},
		{
			name: "Test if error is returned properly on storage error",
			arg: &pb.UpdateArticleRequest{
				Article: article,
			},
			cmd: func() storagemocks.Writer {
				m := storagemocks.NewWriter()
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Article")).Return(errors.New("test err")).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			client := setUpServer(ctx, storagemocks.NewGetter(), tt.cmd, tt.broker)

			got, err := client.Update(ctx, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed to Update article, err: %v", err)
				return
			}

			if tt.wantErr {
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(emptypb.Empty{})) {
				t.Errorf("Wrong response:\n got = %+v\n want = %+v\n", got, tt.want)
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	v := gentest.RandomArticle(2, 5)
	article := &pb.Article{
		Id: v.Id,
	}

	tests := []struct {
		name    string
		arg     *pb.DeleteArticleRequest
		cmd     storagemocks.Writer
		broker  mocks.Broker
		want    *emptypb.Empty
		wantErr bool
	}{
		{
			name: "Test if response is returned properly on simple request",
			arg: &pb.DeleteArticleRequest{
				Id: article.Id,
			},
			want: &emptypb.Empty{},
			cmd: func() storagemocks.Writer {
				m := storagemocks.NewWriter()
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
		},
		{
			name: "Test if error is returned properly on storage error",
			arg: &pb.DeleteArticleRequest{
				Id: article.Id,
			},
			want:    nil,
			wantErr: true,
			cmd: func() storagemocks.Writer {
				m := storagemocks.NewWriter()
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("test err")).Once()
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
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			client := setUpServer(ctx, storagemocks.NewGetter(), tt.cmd, tt.broker)

			got, err := client.Delete(ctx, tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Failed to Delete article, err: %v", err)
				return
			}
			if tt.wantErr {
				return
			}

			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(emptypb.Empty{})) {
				t.Errorf("Wrong response:\n got = %+v\n want = %+v\n", got, tt.want)
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
		name    string
		arg     *pb.GetArticlesRequest
		query   storagemocks.Getter
		broker  mocks.Broker
		want    []*pb.Article
		wantErr bool
	}{
		{
			name: "Test if response is returned properly on simple request",
			arg: &pb.GetArticlesRequest{
				Offset: "0",
				Limit:  "5",
			},
			query: func() storagemocks.Getter {
				m := storagemocks.NewGetter()
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(articles, nil).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			want:    pbArticles,
			wantErr: false,
		},
		{
			name: "Test if error is returned properly on storage error",
			arg:  &pb.GetArticlesRequest{},
			query: func() storagemocks.Getter {
				m := storagemocks.NewGetter()
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]entity.Article{}, errors.New("test err")).Once()
				return m
			}(),
			broker: func() mocks.Broker {
				m := mocks.NewBroker()
				m.On("ResilientPublish", mock.AnythingOfType("event.Event")).Return(nil).Once()
				return m
			}(),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			client := setUpServer(ctx, tt.query, storagemocks.NewWriter(), tt.broker)

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
			}
		})
	}
}
