package storage

import (
	"context"
	"io"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-lib/event"
)

// CQRStorage (Command-Query Responsibility Segregation Storage) is
// a standard storage that can apply sync events using CatchUp() method.
type CQRStorage interface {
	// CatchUp is an Event handler used for synchronizing read and write models.
	CatchUp(event.Event)

	// EventHandlers returns all event handlers specific to an implementation
	// which are registered at the composition root.
	// These handlers should not be used to sync read and write models
	// and should be separated from them, applying other domain events.
	EventHandlers() map[event.EventType][]event.Handler

	Storage
}

type Eventstore interface {
	event.Consumer
	Writer
}

type Storage interface {
	Getter
	Writer
}

type Getter interface {
	Get(ctx context.Context, id string) (entity.Article, error)
	GetBelongingIDs(ctx context.Context, userId string) ([]string, error) // Get article ids belonging to a user.
	GetMultiple(ctx context.Context, offset, limit string) ([]entity.Article, error)

	io.Closer
}

type Writer interface {
	Create(context.Context, entity.Article) error
	Update(context.Context, entity.Article) error
	Delete(ctx context.Context, id string) error

	io.Closer
}
