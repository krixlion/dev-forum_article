package server

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	pb "github.com/krixlion/dev_forum-article/pkg/grpc/v1"
	"github.com/krixlion/dev_forum-article/pkg/helpers/gentest"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_articleFromPB(t *testing.T) {
	id := gentest.RandomString(3)
	userId := gentest.RandomString(3)
	body := gentest.RandomString(3)
	title := gentest.RandomString(3)

	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := articleFromPB(tt.arg)

			if !cmp.Equal(got, tt.want, cmpopts.IgnoreUnexported(pb.Article{})) {
				t.Errorf("Articles are not equal:\n got = %+v\n want = %+v\n", got, tt.want)
				return
			}
		})
	}
}
