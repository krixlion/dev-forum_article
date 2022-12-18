package event

import (
	"time"

	"github.com/krixlion/dev-forum_article/pkg/entity"
)

// Events are sent to the queue in JSON format.
type Event struct {
	Entity    entity.EntityName `json:"entity,omitempty"`
	Type      EventType         `json:"type,omitempty"`
	Body      []byte            `json:"body,omitempty"` // Must be marshaled to JSON.
	Timestamp time.Time         `json:"timestamp,omitempty"`
}

type EventType string

const (
	Created EventType = "created"
	Deleted EventType = "deleted"
	Updated EventType = "updated"
)
