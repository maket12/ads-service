package fakes

import (
	"container/list"
	"context"
	"sync"

	"github.com/google/uuid"
)

type FakePublisher struct {
	mu    sync.Mutex
	queue *list.List
}

func NewFakePublisher() *FakePublisher {
	return &FakePublisher{queue: list.New()}
}

func (f *FakePublisher) PublishAccountCreate(_ context.Context, accountID uuid.UUID) error {
	f.mu.Lock()
	f.queue.PushBack(accountID)
	f.mu.Unlock()
	return nil
}
