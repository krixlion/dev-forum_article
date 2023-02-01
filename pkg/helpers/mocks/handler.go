package mocks

import (
	"github.com/krixlion/dev_forum-article/pkg/event"
	"github.com/stretchr/testify/mock"
)

type Handler struct {
	*mock.Mock
}

func (h Handler) Handle(e event.Event) {
	h.Called(e)
}
