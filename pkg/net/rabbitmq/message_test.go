package rabbitmq

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/helpers/gentest"
)

func Test_makeMessageFromEvent(t *testing.T) {

	jsonArticle := gentest.RandomJSONArticle()
	e := event.Event{
		Entity:    entity.ArticleEntity,
		Type:      event.Created,
		Body:      jsonArticle,
		Timestamp: time.Now(),
	}
	jsonEvent, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	tests := []struct {
		desc string
		arg  event.Event
		want Message
	}{
		{
			desc: "Test if message is correctly processed from simple event",
			arg:  e,
			want: Message{
				Body:        jsonEvent,
				ContentType: "application/json",
				Timestamp:   e.Timestamp,
				Route: Route{
					ExchangeName: "article",
					ExchangeType: "topic",
					RoutingKey:   "article.event.created",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := makeMessageFromEvent(tt.arg)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("makeMessageFromEvent() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_routingKeyFromEvent(t *testing.T) {
	type args struct {
		e event.Event
	}
	tests := []struct {
		desc     string
		args     args
		wantRKey string
	}{
		{
			desc: "Test if returns correct keys with simple data.",
			args: args{e: event.Event{
				Entity: entity.UserEntity,
				Type:   event.Updated,
			}},
			wantRKey: "user.event.updated",
		},
		{
			desc: "Test if returns correct keys with simple data.",
			args: args{e: event.Event{
				Entity: entity.ArticleEntity,
				Type:   event.Deleted,
			}},
			wantRKey: "article.event.deleted",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if gotRKey := routingKeyFromEvent(tt.args.e); gotRKey != tt.wantRKey {
				t.Errorf("routingKeyFromEvent() = %v, want %v", gotRKey, tt.wantRKey)
			}
		})
	}
}

func Test_makeRouteFromEvent(t *testing.T) {
	type args struct {
		e event.Event
	}
	tests := []struct {
		desc string
		args args
		want Route
	}{
		{
			desc: "Test if returns correct route with simple data.",
			args: args{e: event.Event{
				Entity: entity.ArticleEntity,
				Type:   event.Created,
			}},
			want: Route{
				ExchangeName: "article",
				ExchangeType: "topic",
				RoutingKey:   "article.event.created",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got := makeRouteFromEvent(tt.args.e); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeRouteFromEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}
