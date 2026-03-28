package repository

import "context"

// TransactionManager abstracts cross-repository transaction handling.
// If fn returns nil the transaction is committed; if fn returns an error it is rolled back.
type TransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
