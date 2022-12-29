package rabbitmq_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/krixlion/dev-forum_article/pkg/env"
	"github.com/krixlion/dev-forum_article/pkg/helpers/gentest"
	"github.com/krixlion/dev-forum_article/pkg/logging"
	"github.com/krixlion/dev-forum_article/pkg/net/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/google/go-cmp/cmp"
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
	logger, _ := logging.NewLogger()
	config := rabbitmq.Config{
		QueueSize:         100,
		ReconnectInterval: time.Millisecond * 100,
		MaxRequests:       30,
		ClearInterval:     time.Second * 5,
		ClosedTimeout:     time.Second * 15,
	}
	mq := rabbitmq.NewRabbitMQ(consumer, user, pass, host, port, logger, config)
	return mq
}

func TestPubSub(t *testing.T) {
	mq := setUpMQ()
	defer mq.Close()

	article := gentest.RandomArticle(3, 5)
	data, err := json.Marshal(article)
	if err != nil {
		t.Fatalf("Failed to marshal article, input: %+v, err: %s", article, err)
	}

	testCases := []struct {
		desc    string
		arg     rabbitmq.Message
		wantErr bool
	}{
		{
			desc: "Test if a simple message is correctly published and consumed.",
			arg: rabbitmq.Message{
				Body:        data,
				ContentType: rabbitmq.ContentTypeJson,
				Timestamp:   time.Now(),
				Route: rabbitmq.Route{
					ExchangeName: "",
					ExchangeType: amqp.ExchangeTopic,
					RoutingKey:   "",
				},
			},
			wantErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := mq.Publish(ctx, tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("RabbitMQ.Publish() error = %+v\n, wantErr = %+v\n", err, tC.wantErr)
			}

			events, err := mq.Consume(ctx, "deleteArticle", tC.arg.Route)
			if (err != nil) != tC.wantErr {
				t.Errorf("RabbitMQ.Consume() error = %+v\n, wantErr = %+v\n", err, tC.wantErr)
			}

			event := <-events
			if !cmp.Equal(tC.arg, event) {
				t.Fatalf("Events are not equal, want = %+v\n, got = %+v\n", tC.arg, event)
			}
		})
	}
}

func TestPubSubPipeline(t *testing.T) {
	mq := setUpMQ()
	defer mq.Close()

	article := gentest.RandomArticle(3, 5)
	data, err := json.Marshal(article)
	if err != nil {
		t.Fatalf("Failed to marshal article, input = %+v\n, err = %s", article, err)
	}

	testCases := []struct {
		desc    string
		arg     rabbitmq.Message
		wantErr bool
	}{
		{
			desc: "Test if a simple message is correctly published through a pipeline and consumed.",
			arg: rabbitmq.Message{
				Body:        data,
				ContentType: rabbitmq.ContentTypeJson,
				Timestamp:   time.Now(),
				Route: rabbitmq.Route{
					ExchangeName: "",
					ExchangeType: amqp.ExchangeTopic,
					RoutingKey:   "",
				},
			},
			wantErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := mq.Enqueue(tC.arg)
			if (err != nil) != tC.wantErr {
				t.Errorf("RabbitMQ.Enqueue() error = %+v\n, wantErr = %+v\n", err, tC.wantErr)
			}

			events, err := mq.Consume(ctx, "createArticle", tC.arg.Route)
			if (err != nil) != tC.wantErr {
				t.Errorf("RabbitMQ.Consume() error = %+v\n, wantErr = %+v\n", err, tC.wantErr)
			}

			event := <-events
			if !cmp.Equal(tC.arg, event) {
				t.Fatalf("Events are not equal, got = %+v\n  want = %+v\n,", event, tC.arg)
			}
		})
	}
}
