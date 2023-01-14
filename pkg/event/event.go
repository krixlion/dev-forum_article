package event

import (
	"time"
)

// Events are sent to the queue in JSON format.
type Event struct {
	AggregateId string    `json:"aggregate_id,omitempty"`
	Type        EventType `json:"type,omitempty"`
	Body        []byte    `json:"body,omitempty"` // Must be marshaled to JSON.
	Timestamp   time.Time `json:"timestamp,omitempty"`
}

type EventType string

// All event names must be lowercase and follow the structure: "noun-action".
// Eg. article-created, notification-sent, order-accepted.
// For longer names use snake-case naming.
// Eg. changed_password_notification-sent.
const (
	ArticleCreated EventType = "article-created"
	ArticleDeleted EventType = "article-deleted"
	ArticleUpdated EventType = "article-updated"

	UserCreated EventType = "user-created"
	UserDeleted EventType = "user-deleted"
	UserUpdated EventType = "user-updated"
)
