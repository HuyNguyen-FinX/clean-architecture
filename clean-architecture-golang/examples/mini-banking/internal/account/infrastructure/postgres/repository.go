package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

var _ application.AccountRepository = (*Repository)(nil)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		panic("postgres: nil pool")
	}
	return &Repository{pool: pool}
}

func (r *Repository) FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	query := "SELECT id, balance_minor, currency, overdraft_limit_minor, status " +
		"FROM accounts WHERE id = $1"

	var row pgx.Row
	if tx, ok := transactionFromContext(ctx); ok {
		row = tx.QueryRow(ctx, query+" FOR UPDATE", id)
	} else {
		row = r.pool.QueryRow(ctx, query, id)
	}

	var stored accountRow
	err := row.Scan(
		&stored.id,
		&stored.balance,
		&stored.currency,
		&stored.overdraftLimit,
		&stored.status,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find account %q: %w", id, err)
	}
	return stored.toDomain()
}

func (r *Repository) Save(ctx context.Context, account *domain.Account) error {
	if account == nil {
		return domain.ErrInvalidAccountID
	}
	query := "UPDATE accounts SET balance_minor = $1, currency = $2, " +
		"overdraft_limit_minor = $3, status = $4, updated_at = now() WHERE id = $5"
	args := []any{
		account.Balance().Amount(),
		account.Balance().Currency().String(),
		account.OverdraftLimit().Amount(),
		string(account.Status()),
		string(account.ID()),
	}

	var (
		tag pgconnCommandTag
		err error
	)
	if tx, ok := transactionFromContext(ctx); ok {
		tag, err = tx.Exec(ctx, query, args...)
	} else {
		tag, err = r.pool.Exec(ctx, query, args...)
	}
	if err != nil {
		return fmt.Errorf("save account %q: %w", account.ID(), err)
	}
	if tag.RowsAffected() != 1 {
		return domain.ErrAccountNotFound
	}
	return nil
}

type pgconnCommandTag interface {
	RowsAffected() int64
}
