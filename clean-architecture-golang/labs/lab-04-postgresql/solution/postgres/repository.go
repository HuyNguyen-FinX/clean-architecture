package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/cleanarch/lab04/solution/application"
	"example.com/cleanarch/lab04/solution/domain"
)

var _ application.AccountRepository = (*Repository)(nil)

type Repository struct {
	pool *pgxpool.Pool
}

type accountRow struct {
	id        string
	balance   int64
	currency  string
	overdraft int64
	status    string
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		panic("postgres: nil pool")
	}
	return &Repository{pool: pool}
}

func (r *Repository) FindByID(ctx context.Context, id string) (*domain.Account, error) {
	row := r.pool.QueryRow(ctx,
		"SELECT id, balance_minor, currency, overdraft_minor, status FROM accounts WHERE id=$1",
		id,
	)
	var stored accountRow
	err := row.Scan(&stored.id, &stored.balance, &stored.currency, &stored.overdraft, &stored.status)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, application.ErrAccountNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find account %q: %w", id, err)
	}
	account, err := stored.toDomain()
	if err != nil {
		return nil, fmt.Errorf("map account %q: %w", id, err)
	}
	return account, nil
}

func (r *Repository) Save(ctx context.Context, account *domain.Account) error {
	tag, err := r.pool.Exec(ctx,
		"UPDATE accounts SET balance_minor=$1, currency=$2, overdraft_minor=$3, status=$4 WHERE id=$5",
		account.Balance(), account.Currency(), account.Overdraft(), account.Status(), account.ID(),
	)
	if err != nil {
		return fmt.Errorf("save account %q: %w", account.ID(), err)
	}
	if tag.RowsAffected() != 1 {
		return application.ErrAccountNotFound
	}
	return nil
}

func (row accountRow) toDomain() (*domain.Account, error) {
	return domain.RehydrateAccount(row.id, row.balance, row.currency, row.overdraft, row.status)
}
