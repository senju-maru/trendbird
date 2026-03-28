package persistence

import (
	"context"

	"github.com/trendbird/backend/internal/domain/repository"
	"gorm.io/gorm"
)

type txKey struct{}

var _ repository.TransactionManager = (*transactionManager)(nil)

type transactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new TransactionManager backed by GORM.
func NewTransactionManager(db *gorm.DB) repository.TransactionManager {
	return &transactionManager{db: db}
}

func (tm *transactionManager) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

// GetDB retrieves the *gorm.DB from ctx if within a transaction, otherwise returns the original db.
func GetDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db
}
