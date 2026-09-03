package postgres

import (
	"fmt"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

type accountRow struct {
	id             string
	balance        int64
	currency       string
	overdraftLimit int64
	status         string
}

func (row accountRow) toDomain() (*domain.Account, error) {
	id, err := domain.NewAccountID(row.id)
	if err != nil {
		return nil, fmt.Errorf("account id: %w", err)
	}
	balance, err := domain.NewMoney(row.balance, row.currency)
	if err != nil {
		return nil, fmt.Errorf("account balance: %w", err)
	}
	overdraft, err := domain.NewMoney(row.overdraftLimit, row.currency)
	if err != nil {
		return nil, fmt.Errorf("account overdraft: %w", err)
	}
	account, err := domain.RehydrateAccount(id, balance, overdraft, domain.AccountStatus(row.status))
	if err != nil {
		return nil, fmt.Errorf("rehydrate account %q: %w", row.id, err)
	}
	return account, nil
}
