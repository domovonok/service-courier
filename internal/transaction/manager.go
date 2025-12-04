package transaction

import (
	"context"
)

type Tx interface {
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type Manager interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
