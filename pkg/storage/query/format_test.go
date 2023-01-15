package query

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/helpers/gentest"
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
				t.Fatalf("Wrong output:\n got = %+v\n want = %+v\n", got, tC.want)
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
				"id":      article.Id,
				"user_id": article.UserId,
				"body":    article.Body,
				"title":   article.Title,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := mapArticle(tC.arg)
			if !cmp.Equal(got, tC.want) {
				t.Fatalf("Wrong output:\n got = %+v\n want = %+v\n", got, tC.want)
			}
		})
	}
}

func Test_addArticlesPrefix(t *testing.T) {
	testCases := []struct {
		desc string
		arg  string
		want string
	}{
		{
			desc: "Test if correctly adds prefix to an alias",
			arg:  "*->title",
			want: "articles:*->title",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			got := addArticlesPrefix(tC.arg)
			if got != tC.want {
				t.Fatalf("Failed to add prefix:\n got = %+v\n want = %+v\n", got, tC.want)
			}
		})
	}
}
