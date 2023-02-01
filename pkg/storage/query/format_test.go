package query

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-article/pkg/helpers/gentest"
)

func Test_toLowerSnakeCase(t *testing.T) {
	testCases := []struct {
		desc string
		arg  string
		want string
	}{
		{
			desc: "Test on simple data",
			arg:  "UserId",
			want: "user_id",
		},
		{
			desc: "Test on simple data",
			arg:  "UserMightWanttoFixThat",
			want: "user_might_wantto_fix_that",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := toLowerSnakeCase(tC.arg)
			if !cmp.Equal(got, tC.want) {
				t.Errorf("Wrong output:\n got = %+v\n want = %+v\n", got, tC.want)
				return
			}
		})
	}
}

func Test_mapArticle(t *testing.T) {
	article := gentest.RandomArticle(2, 5)

	testCases := []struct {
		desc string
		arg  entity.Article
		want map[string]string
	}{
		{
			desc: "Test on simple random data",
			arg:  article,
			want: map[string]string{
				"id":         article.Id,
				"user_id":    article.UserId,
				"body":       article.Body,
				"title":      article.Title,
				"created_at": article.CreatedAt.Format(time.RFC3339),
				"updated_at": article.UpdatedAt.Format(time.RFC3339),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := mapArticle(tC.arg)
			if !cmp.Equal(got, tC.want) {
				t.Errorf("Wrong output:\n got = %+v\n want = %+v\n", got, tC.want)
				return
			}
		})
	}
}

func Test_addPrefix(t *testing.T) {
	type args struct {
		prefix string
		key    string
	}
	testCases := []struct {
		desc string
		args args
		want string
	}{
		{
			desc: "Test if correctly adds prefix to an alias",
			args: args{
				key:    "*->title",
				prefix: articlesPrefix,
			},
			want: "articles:*->title",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := addPrefix(tC.args.prefix, tC.args.key)
			if got != tC.want {
				t.Errorf("Failed to add prefix:\n got = %+v\n want = %+v\n", got, tC.want)
				return
			}
		})
	}
}
