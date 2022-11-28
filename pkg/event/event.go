package event

import (
	"time"
)

// Events are sent to the queue in JSON format.
type Event struct {
	Entity    string    `json:"entity,omitempty"`
	Type      EventType `json:"type,omitempty"`
	Body      string    `json:"body,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

type EventType string

const (
	Created EventType = "created"
	Deleted EventType = "deleted"
)
