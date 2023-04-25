package storagemocks

import (
	"context"

	"github.com/krixlion/dev_forum-article/pkg/entity"
	"github.com/krixlion/dev_forum-lib/event"
	"github.com/stretchr/testify/mock"
)

type CQRStorage struct {
	*mock.Mock
}

func NewCQRStorage() CQRStorage {
	return CQRStorage{
		Mock: new(mock.Mock),
	}
}

// CatchUp is an Event handler used for synchronizing read and write models.
func (m CQRStorage) CatchUp(_ event.Event) {
	panic("not implemented") // TODO: Implement
}

// EventHandlers returns all event handlers specific to an implementation
// which are registered at the composition root.
// These handlers should not be used to sync read and write models
// and should be separated from them, applying other domain events.
func (m CQRStorage) EventHandlers() map[event.EventType][]event.Handler {
	panic("not implemented") // TODO: Implement
}

func (m CQRStorage) Get(ctx context.Context, id string) (entity.Article, error) {
	panic("not implemented") // TODO: Implement
}

func (m CQRStorage) GetBelongingIDs(ctx context.Context, userId string) ([]string, error) {
	panic("not implemented") // TODO: Implement
}

func (m CQRStorage) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.Article, error) {
	panic("not implemented") // TODO: Implement
}

func (m CQRStorage) Close() error {
	panic("not implemented") // TODO: Implement
}

func (m CQRStorage) Create(_ context.Context, _ entity.Article) error {
	panic("not implemented") // TODO: Implement
}

func (m CQRStorage) Update(_ context.Context, _ entity.Article) error {
	panic("not implemented") // TODO: Implement
}

func (m CQRStorage) Delete(ctx context.Context, id string) error {
	panic("not implemented") // TODO: Implement
}
