package broker

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/helpers/gentest"
	rabbitmq "github.com/krixlion/dev-forum_rabbitmq"
)

func Test_messageFromEvent(t *testing.T) {

	jsonArticle := gentest.RandomJSONArticle(3, 5)
	e := event.Event{
		AggregateId: "article",
		Type:        event.ArticleCreated,
		Body:        jsonArticle,
		Timestamp:   time.Now(),
	}
	jsonEvent, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}

	tests := []struct {
		desc string
		arg  event.Event
		want rabbitmq.Message
	}{
		{
			desc: "Test if message is correctly processed from simple event",
			arg:  e,
			want: rabbitmq.Message{
				Body:        jsonEvent,
				ContentType: "application/json",
				Timestamp:   e.Timestamp,
				Route: rabbitmq.Route{
					ExchangeName: "article",
					ExchangeType: "topic",
					RoutingKey:   "article.event.created",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got := messageFromEvent(tt.arg)
			if !cmp.Equal(got, tt.want) {
				t.Errorf("MakeMessageFromEvent() = %+v, want %+v", got, tt.want)
				return
			}
		})
	}
}

func Test_routeFromEvent(t *testing.T) {
	type args struct {
		Type event.EventType
	}
	tests := []struct {
		desc string
		args args
		want rabbitmq.Route
	}{
		{
			desc: "Test if returns correct route with simple data.",
			args: args{
				Type: event.ArticleCreated,
			},
			want: rabbitmq.Route{
				ExchangeName: "article",
				ExchangeType: "topic",
				RoutingKey:   "article.event.created",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got := routeFromEvent(tt.args.Type); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeRouteFromEvent() = %v, want %v", got, tt.want)
				return
			}
		})
	}
}
