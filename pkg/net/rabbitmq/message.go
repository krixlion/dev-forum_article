package rabbitmq

import "time"

const ContentTypeJson = "application/json"
const ContentTypeText = "text/plain"

type Message struct {
	Body        string
	ContentType string
	Timestamp   time.Time
	OnQueue     bool // Whether a Message should be published directly to a queue.
	Enqueue     bool // Whether a Message should be enqueued on failure for later retry.
	Route
}

type Route struct {
	Exchange   string
	RoutingKey string
	QueueName  string
}
