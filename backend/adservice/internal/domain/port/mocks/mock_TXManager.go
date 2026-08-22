package mocks

import (
	"context"

	"github.com/avito-tech/go-transaction-manager/trm/v2"
)

type FakeTxManager struct{}

func (f FakeTxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (f FakeTxManager) DoWithSettings(ctx context.Context, s trm.Settings, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
