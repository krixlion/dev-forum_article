package storage

import (
	"context"
	"io"

	"github.com/krixlion/dev-forum_article/pkg/entity"
	"github.com/krixlion/dev-forum_article/pkg/event"
)

type Storage interface {
	Getter
	WriteStorage
	CatchUp(event.Event)
}

type ReadStorage interface {
	Getter
	WriteStorage
}

type Getter interface {
	io.Closer
	Get(ctx context.Context, id string) (entity.Article, error)
	GetMultiple(ctx context.Context, offset, limit string) ([]entity.Article, error)
}

type WriteStorage interface {
	io.Closer
	Create(context.Context, entity.Article) error
	Update(context.Context, entity.Article) error
	Delete(ctx context.Context, id string) error
}

type Eventstore interface {
	event.Consumer
	WriteStorage
}
