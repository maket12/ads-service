package fakes

import (
	"container/list"
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/maket12/ads-service/backend/adservice/internal/domain/model"
)

type FakePublisher struct {
	mu    sync.Mutex
	queue *list.List
}

func NewFakePublisher() *FakePublisher {
	return &FakePublisher{queue: list.New()}
}

func (f *FakePublisher) PublishAdPublished(_ context.Context, ad *model.Ad) error {
	f.mu.Lock()
	f.queue.PushBack(ad)
	f.mu.Unlock()
	return nil
}

func (f *FakePublisher) PublishAdUpdated(_ context.Context, ad *model.Ad) error {
	f.mu.Lock()
	f.queue.PushBack(ad)
	f.mu.Unlock()
	return nil
}

func (f *FakePublisher) PublishAdRejected(ctx context.Context, adID uuid.UUID) error {
	f.mu.Lock()
	f.queue.PushBack(adID)
	f.mu.Unlock()
	return nil
}

func (f *FakePublisher) PublishAdDeleted(ctx context.Context, adID uuid.UUID) error {
	f.mu.Lock()
	f.queue.PushBack(adID)
	f.mu.Unlock()
	return nil
}
