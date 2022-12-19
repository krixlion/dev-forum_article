package rabbitmq_test

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/env"
	"github.com/krixlion/dev-forum_article/pkg/event"
	"github.com/krixlion/dev-forum_article/pkg/net/rabbitmq"

	"github.com/gofrs/uuid"
	"github.com/google/go-cmp/cmp"
	"github.com/matryer/is"
)

const consumer = "TESTING"

var (
	port string
	host string
	user string
	pass string
)

func init() {
	env.Load("app")

	port = os.Getenv("MQ_PORT")
	host = os.Getenv("MQ_HOST")
	user = os.Getenv("MQ_USER")
	pass = os.Getenv("MQ_PASS")
}

func setUpMQ() *rabbitmq.RabbitMQ {

	config := rabbitmq.Config{
		QueueSize:         100,
		ReconnectInterval: time.Second * 2,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}

	return rabbitmq.NewRabbitMQ(consumer, user, pass, host, port, config)
}

func randomText(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	v := make([]rune, length)
	for i := range v {
		v[i] = letters[rand.Intn(len(letters))]
	}
	return string(v)
}

func getRandomArticle() (entity.Article, error) {

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
		Title:  randomText(10),
		Body:   randomText(30),
	}, nil
}

func TestPubSub(t *testing.T) {
	is := is.New(t)
	mq := setUpMQ()

	go func() {
		err := mq.Run()
		defer mq.Close()
		is.NoErr(err)
	}()

	article, err := getRandomArticle()
	is.NoErr(err)

	data, err := json.Marshal(article)
	is.NoErr(err)

	testCases := []struct {
		desc string
		arg  event.Event
	}{
		{
			desc: "Test if a simple message is correctly published and consumed.",
			arg: event.Event{
				Entity:    entity.ArticleEntity,
				Type:      event.Deleted,
				Body:      data,
				Timestamp: time.Now(),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// is needs to have up to date T.
			is := is.New(t)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := mq.Publish(ctx, tC.arg)
			is.NoErr(err)

			events, err := mq.Consume(ctx, "deleteArticle", tC.arg.Entity, tC.arg.Type)
			is.NoErr(err)

			event := <-events
			if !cmp.Equal(tC.arg, event) {
				t.Fatalf("Events are not equal, wanted: %v, received %v", tC.arg, event)
			}
			received := entity.Article{}
			err = json.Unmarshal(event.Body, &received)
			is.NoErr(err)

			is.Equal(received, article)
		})
	}
}

func TestPubSubPipeline(t *testing.T) {
	is := is.New(t)
	mq := setUpMQ()

	go func() {
		err := mq.Run()
		defer mq.Close()
		is.NoErr(err)
	}()

	article, err := getRandomArticle()
	is.NoErr(err)

	data, err := json.Marshal(article)
	is.NoErr(err)

	testCases := []struct {
		desc string
		arg  event.Event
	}{
		{
			desc: "Test if a simple message is correctly published through a pipeline and consumed.",
			arg: event.Event{
				Entity:    entity.ArticleEntity,
				Type:      event.Created,
				Body:      data,
				Timestamp: time.Now(),
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			// is needs to have up to date T.
			is := is.New(t)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := mq.ResilientPublish(ctx, tC.arg)
			is.NoErr(err)

			events, err := mq.Consume(ctx, "createArticle", tC.arg.Entity, tC.arg.Type)
			is.NoErr(err)

			event := <-events
			if !cmp.Equal(tC.arg, event) {
				t.Fatalf("Events are not equal, wanted: %v, received %v", tC.arg, event)
			}

			received := entity.Article{}
			err = json.Unmarshal(event.Body, &received)
			is.NoErr(err)

			is.Equal(received, article)
		})
	}
}
