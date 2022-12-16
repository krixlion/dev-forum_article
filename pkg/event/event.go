package event

import (
	"time"
)

// Events are sent to the queue in JSON format.
type Event struct {
	Entity    string    `json:"entity,omitempty"`
	Type      EventType `json:"type,omitempty"`
	Body      []byte    `json:"body,omitempty"` // Must be marshaled to JSON.
	Timestamp time.Time `json:"timestamp,omitempty"`
}

type EventType string

const (
	ArticleCreated EventType = "article-created"
	ArticleDeleted EventType = "article-deleted"
	ArticleUpdated EventType = "article-updated"
)
