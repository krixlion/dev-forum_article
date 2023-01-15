package server_test

import (
	"context"
	"errors"
	"log"
	"net"
	"testing"
	"time"

	"github.com/Krixlion/def-forum_proto/article_service/pb"
	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/mock"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/helpers/gentest"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/server"
	"github.com/krixlion/dev-forum_article/pkg/storage"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

// setUpServer initializes and runs in the background a gRPC
// server allowing only for local calls for testing.
// Returns a client to interact with the server.
// The server is shutdown when ctx.Done() receives.
func setUpServer(ctx context.Context, mock storage.CQRStorage) pb.ArticleServiceClient {
	// bufconn allows the server to call itself
	// great for testing across whole infrastructure
	lis := bufconn.Listen(1024 * 1024)
	bufDialer := func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}

	s := grpc.NewServer()
	server := server.ArticleServer{
		Storage: mock,
	}
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
		Id:     v.Id,
		UserId: v.UserId,
		Title:  v.Title,
		Body:   v.Body,
	}

	testCases := []struct {
		desc    string
		arg     *pb.GetArticleRequest
		want    *pb.GetArticleResponse
		wantErr bool
		storage storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.GetArticleRequest{
				ArticleId: article.Id,
			},
			want: &pb.GetArticleResponse{
				Article: article,
			},
			storage: func() (m mockStorage) {
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(v, nil).Times(1)
				return
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.GetArticleRequest{
				ArticleId: "",
			},
			want:    nil,
			wantErr: true,
			storage: func() (m mockStorage) {
				m.On("Get", mock.Anything, mock.AnythingOfType("string")).Return(entity.Article{}, errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()

			client := setUpServer(ctx, tC.storage)

			ctx, cancel := context.WithTimeout(ctx, time.Second)
			defer cancel()

			getResponse, err := client.Get(ctx, tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("Failed to Get article, err: %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if getResponse != tC.want {
				if !cmp.Equal(getResponse.Article, tC.want.Article, cmpopts.IgnoreUnexported(pb.Article{})) {
					t.Errorf("Articles are not equal:\n Got = %+v\n, want = %+v\n", getResponse.Article, tC.want.Article)
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

	testCases := []struct {
		desc     string
		arg      *pb.CreateArticleRequest
		dontWant *pb.CreateArticleResponse
		wantErr  bool
		storage  storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.CreateArticleRequest{
				Article: article,
			},
			dontWant: &pb.CreateArticleResponse{
				Id: article.Id,
			},
			storage: func() (m mockStorage) {
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Article")).Return(nil).Times(1)
				return
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.CreateArticleRequest{
				Article: article,
			},
			dontWant: nil,
			wantErr:  true,
			storage: func() (m mockStorage) {
				m.On("Create", mock.Anything, mock.AnythingOfType("entity.Article")).Return(errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tC.storage)

			createResponse, err := client.Create(ctx, tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("Failed to Get article, err: %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if createResponse != tC.dontWant {
				if cmp.Equal(createResponse.Id, tC.dontWant.Id) {
					t.Errorf("Article IDs was not reassigned:\n Got = %+v\n want = %+v\n", createResponse.Id, tC.dontWant.Id)
					return
				}
				if _, err := uuid.FromString(createResponse.Id); err != nil {
					t.Errorf("Article ID is not correct UUID:\n ID = %+v\n err = %+v", createResponse.Id, err)
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

	testCases := []struct {
		desc    string
		arg     *pb.UpdateArticleRequest
		want    *pb.UpdateArticleResponse
		wantErr bool
		storage storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.UpdateArticleRequest{
				Article: article,
			},
			want: &pb.UpdateArticleResponse{},
			storage: func() (m mockStorage) {
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Article")).Return(nil).Times(1)
				return
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.UpdateArticleRequest{
				Article: article,
			},
			want:    nil,
			wantErr: true,
			storage: func() (m mockStorage) {
				m.On("Update", mock.Anything, mock.AnythingOfType("entity.Article")).Return(errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tC.storage)

			got, err := client.Update(ctx, tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("Failed to Update article, err: %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if got != tC.want {
				if !cmp.Equal(got, tC.want, cmpopts.IgnoreUnexported(pb.UpdateArticleResponse{})) {
					t.Errorf("Wrong response:\n got = %+v\n want = %+v\n", got, tC.want)
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

	testCases := []struct {
		desc    string
		arg     *pb.DeleteArticleRequest
		want    *pb.DeleteArticleResponse
		wantErr bool
		storage storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.DeleteArticleRequest{
				ArticleId: article.Id,
			},
			want: &pb.DeleteArticleResponse{},
			storage: func() (m mockStorage) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(nil).Times(1)
				return
			}(),
		},
		{
			desc: "Test if error is returned properly on storage error",
			arg: &pb.DeleteArticleRequest{
				ArticleId: article.Id,
			},
			want:    nil,
			wantErr: true,
			storage: func() (m mockStorage) {
				m.On("Delete", mock.Anything, mock.AnythingOfType("string")).Return(errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tC.storage)

			got, err := client.Delete(ctx, tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("Failed to Delete article, err: %v", err)
				return
			}

			// Equals false if both are nil or they point to the same memory address
			// so be sure to use seperate structs when providing args in order to prevent SEGV.
			if got != tC.want {
				if !cmp.Equal(got, tC.want, cmpopts.IgnoreUnexported(pb.DeleteArticleResponse{})) {
					t.Errorf("Wrong response:\n got = %+v\n want = %+v\n", got, tC.want)
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
			Id:     v.Id,
			UserId: v.UserId,
			Title:  v.Title,
			Body:   v.Body,
		}
		pbArticles = append(pbArticles, pbArticle)
	}

	testCases := []struct {
		desc    string
		arg     *pb.GetArticlesRequest
		want    []*pb.Article
		wantErr bool
		storage storage.CQRStorage
	}{
		{
			desc: "Test if response is returned properly on simple request",
			arg: &pb.GetArticlesRequest{
				Offset: "0",
				Limit:  "5",
			},
			want: pbArticles,
			storage: func() (m mockStorage) {
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(articles, nil).Times(1)
				return
			}(),
		},
		{
			desc:    "Test if error is returned properly on storage error",
			arg:     &pb.GetArticlesRequest{},
			want:    nil,
			wantErr: true,
			storage: func() (m mockStorage) {
				m.On("GetMultiple", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]entity.Article{}, errors.New("test err")).Times(1)
				return
			}(),
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, shutdown := context.WithCancel(context.Background())
			defer shutdown()
			client := setUpServer(ctx, tC.storage)

			stream, err := client.GetStream(ctx, tC.arg)
			if err != nil {
				t.Errorf("Failed to Get stream, err: %v", err)
				return
			}

			var got []*pb.Article
			for i := 0; i < len(tC.want); i++ {
				article, err := stream.Recv()
				if (err != nil) != tC.wantErr {
					t.Errorf("Failed to receive article from stream, err: %v", err)
					return
				}
				got = append(got, article)
			}

			if !cmp.Equal(got, tC.want, cmpopts.IgnoreUnexported(pb.Article{})) {
				t.Errorf("Articles are not equal:\n Got = %+v\n want = %+v\n", got, tC.want)
				return
			}
		})
	}
}
