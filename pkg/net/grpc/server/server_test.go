package server_test

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/Krixlion/def-forum_proto/article_service/pb"
	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev-forum_article/pkg/helpers/gentest"
	"github.com/krixlion/dev-forum_article/pkg/net/grpc/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	// bufconn allows the server to call itself
	// great for testing across whole infrastructure
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	server := server.ArticleServer{}
	pb.RegisterArticleServiceServer(s, server)
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func setUpClient(ctx context.Context) pb.ArticleServiceClient {
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client := pb.NewArticleServiceClient(conn)
	return client
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	client := setUpClient(ctx)
	v := gentest.RandomArticle(2, 5)

	article := &pb.Article{
		Id:     v.Id,
		UserId: v.UserId,
		Title:  v.Title,
		Body:   v.Body,
	}
	testCases := []struct {
		desc string
		arg  *pb.GetArticleRequest
		want *pb.GetArticleResponse
	}{
		{
			desc: "",
			arg: &pb.GetArticleRequest{
				ArticleId: article.Id,
			},
			want: &pb.GetArticleResponse{
				Article: article,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			getResponse, err := client.Get(ctx, tC.arg)
			if err != nil {
				t.Fatalf("Failed to Get event, err: %v", err)
			}

			if !cmp.Equal(getResponse.Article, tC.want.Article) {
				t.Fatalf("Articles are not equal. Received = %+v\n, want = %+v\n", getResponse.Article, tC.want.Article)
			}
		})
	}
}

// func TestCreate(t *testing.T) {
// 	// ctx := context.Background()
// 	// client := setUpClient(ctx)

// 	testCases := []struct {
// 		desc string
// 	}{
// 		{
// 			desc: "",
// 		},
// 	}
// 	for _, tC := range testCases {
// 		t.Run(tC.desc, func(t *testing.T) {

// 		})
// 	}
// }
