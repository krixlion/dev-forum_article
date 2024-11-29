package redis

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev_forum-article/internal/gentest"
	"github.com/krixlion/dev_forum-article/pkg/entity"
)

func Test_mapArticle(t *testing.T) {
	article := gentest.RandomArticle(2, 5)

	tests := []struct {
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
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := mapArticle(tt.arg)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("mapArticle:\n got = %+v\n want = %+v\n", got, tt.want)
			}
		})
	}
}

func Test_addPrefix(t *testing.T) {
	type args struct {
		prefix string
		key    string
	}
	tests := []struct {
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
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := addPrefix(tt.args.prefix, tt.args.key)
			if got != tt.want {
				t.Errorf("addPrefix:\n got = %+v\n want = %+v\n", got, tt.want)
			}
		})
	}
}
