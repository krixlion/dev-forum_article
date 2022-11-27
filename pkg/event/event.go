package event

import "time"

type Event struct {
	Type      string
	Body      string
	Timestamp time.Time
}
