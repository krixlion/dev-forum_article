package server

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/helpers/gentest"
	"github.com/krixlion/dev_forum-proto/article_service/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_ArticleFromPB(t *testing.T) {
	id := gentest.RandomString(3)
	userId := gentest.RandomString(3)
	body := gentest.RandomString(3)
	title := gentest.RandomString(3)

	testCases := []struct {
		desc string
		arg  *pb.Article
		want entity.Article
	}{
		{
			desc: "Test if works on simple random data",
			arg: &pb.Article{
				Id:        id,
				UserId:    userId,
				Title:     title,
				Body:      body,
				CreatedAt: timestamppb.New(time.Time{}),
				UpdatedAt: timestamppb.New(time.Time{}),
			},
			want: entity.Article{
				Id:        id,
				UserId:    userId,
				Title:     title,
				Body:      body,
				CreatedAt: time.Time{},
				UpdatedAt: time.Time{},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := articleFromPB(tC.arg)

			if !cmp.Equal(got, tC.want, cmpopts.IgnoreUnexported(pb.Article{})) {
				t.Errorf("Articles are not equal:\n got = %+v\n want = %+v\n", got, tC.want)
				return
			}
		})
	}
}
