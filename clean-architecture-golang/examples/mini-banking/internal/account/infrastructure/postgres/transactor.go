package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
)

var (
	_                    application.Transactor = (*Transactor)(nil)
	ErrNestedTransaction                        = errors.New("nested transaction is not supported")
)

type transactionContextKey struct{}

type Transactor struct {
	pool *pgxpool.Pool
}

func NewTransactor(pool *pgxpool.Pool) *Transactor {
	if pool == nil {
		panic("postgres: nil pool")
	}
	return &Transactor{pool: pool}
}

func (t *Transactor) WithinTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) (err error) {
	if _, exists := transactionFromContext(ctx); exists {
		return ErrNestedTransaction
	}

	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			err = errors.Join(err, fmt.Errorf("rollback transaction: %w", rollbackErr))
		}
	}()

	txCtx := context.WithValue(ctx, transactionContextKey{}, tx)
	if err := fn(txCtx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func transactionFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(transactionContextKey{}).(pgx.Tx)
	return tx, ok
}
