package rabbitmq_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/env"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/net/rabbitmq"
)

const consumer = "TESTING"

var mq *rabbitmq.RabbitMQ

func init() {
	env.Load("app")

	port := os.Getenv("MQ_PORT")
	host := os.Getenv("MQ_HOST")
	user := os.Getenv("MQ_USER")
	pass := os.Getenv("MQ_PASS")

	config := rabbitmq.Config{
		QueueSize:         100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	mq = rabbitmq.NewRabbitMQ(consumer, user, pass, host, port, config)

	go mq.Run()
}

func getTestArticle() (entity.Article, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return entity.Article{}, err
	}

	userId, err := uuid.NewV4()
	if err != nil {
		return entity.Article{}, err
	}

	return entity.Article{
		Id:     id.String(),
		UserId: userId.String(),
		Title:  "test title",
		Body:   "test body",
	}, nil

}

func TestPubSub(t *testing.T) {

	article, err := getTestArticle()
	if err != nil {
		t.Fatalf("Failed to generate article, err: %s", err)
	}

	data, err := json.Marshal(article)
	if err != nil {
		t.Fatalf("Failed to marshal article, err: %s", err)
	}

	tests := []struct {
		name  string
		event event.Event
	}{
		{
			name: "Test if a simple message is correctly published and consumed.",
			event: event.Event{
				Entity: entity.ArticleEntity,
				Type:   event.Created,
				Body:   data,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
			defer cancel()
			err := mq.Publish(ctx, test.event)
			if err != nil {
				t.Fatalf("Failed to publish event, err: %s", err)
			}

			events, err := mq.Consume(ctx, "createArticle", test.event.Entity, test.event.Type)
			if err != nil {
				t.Fatalf("Failed to consume events, err: %s", err)
			}

			event := <-events
			if !reflect.DeepEqual(test.event, event) {
				t.Fatalf("Events are not equal, received: %+v\n", event)
			}
		})
	}
}
