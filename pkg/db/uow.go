package db

import (
	"context"
	"database/sql"
)

// UnitOfWork is the transaction management interface (similar to .NET's IUnitOfWork).
// It provides transaction boundaries for consistent data operations.
type UnitOfWork interface {
	// Begin starts a new transaction.
	Begin(ctx context.Context) error

	// Commit saves all changes within the transaction.
	Commit(ctx context.Context) error

	// Rollback reverts all changes within the transaction.
	Rollback(ctx context.Context) error

	// Tx returns the underlying *sql.Tx for direct access if needed.
	Tx() *sql.Tx
}

// UnitOfWorkImpl implements the UnitOfWork interface.
type UnitOfWorkImpl struct {
	db *sql.DB
	tx *sql.Tx
}

// NewUnitOfWork creates a new unit of work (transaction manager).
func NewUnitOfWork(db *sql.DB) UnitOfWork {
	return &UnitOfWorkImpl{db: db}
}

// Begin starts a new transaction.
func (uow *UnitOfWorkImpl) Begin(ctx context.Context) error {
	tx, err := uow.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	uow.tx = tx
	return nil
}

// Commit saves all changes within the transaction.
func (uow *UnitOfWorkImpl) Commit(ctx context.Context) error {
	if uow.tx == nil {
		return nil
	}
	return uow.tx.Commit()
}

// Rollback reverts all changes within the transaction.
func (uow *UnitOfWorkImpl) Rollback(ctx context.Context) error {
	if uow.tx == nil {
		return nil
	}
	return uow.tx.Rollback()
}

// Tx returns the underlying *sql.Tx.
func (uow *UnitOfWorkImpl) Tx() *sql.Tx {
	return uow.tx
}
