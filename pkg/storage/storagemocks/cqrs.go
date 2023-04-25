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
func (m CQRStorage) CatchUp(event event.Event) {
	m.Called(event)
}

// EventHandlers returns all event handlers specific to an implementation
// which are registered at the composition root.
// These handlers should not be used to sync read and write models
// and should be separated from them, applying other domain events.
func (m CQRStorage) EventHandlers() map[event.EventType][]event.Handler {
	args := m.Called()
	return args.Get(0).(map[event.EventType][]event.Handler)
}

func (m CQRStorage) Get(ctx context.Context, id string) (entity.Article, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(entity.Article), args.Error(1)
}

func (m CQRStorage) GetBelongingIDs(ctx context.Context, userId string) ([]string, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]string), args.Error(1)
}

func (m CQRStorage) GetMultiple(ctx context.Context, offset string, limit string) ([]entity.Article, error) {
	args := m.Called(ctx, offset, limit)
	return args.Get(0).([]entity.Article), args.Error(1)
}

func (m CQRStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m CQRStorage) Create(ctx context.Context, article entity.Article) error {
	args := m.Called(ctx, article)
	return args.Error(0)
}

func (m CQRStorage) Update(ctx context.Context, article entity.Article) error {
	args := m.Called(ctx, article)
	return args.Error(0)
}

func (m CQRStorage) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
