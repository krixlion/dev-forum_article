package entity_test

import (
	"testing"

	"github.com/Krixlion/def-forum_proto/article_service/pb"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/helpers/gentest"
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
				Id:     id,
				UserId: userId,
				Title:  title,
				Body:   body,
			},
			want: entity.Article{
				Id:     id,
				UserId: userId,
				Title:  title,
				Body:   body,
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := entity.ArticleFromPB(tC.arg)

			if !cmp.Equal(got, tC.want, cmpopts.IgnoreUnexported(pb.Article{})) {
				t.Errorf("Articles are not equal:\n got = %+v\n want = %+v\n", got, tC.want)
				return
			}
		})
	}
}
