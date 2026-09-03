package application

import (
	"context"
	"errors"
	"fmt"

	"example.com/cleanarch/lab03/solution/domain"
)

var ErrAccountNotFound = errors.New("account not found")

type AccountRepository interface {
	FindByID(context.Context, domain.AccountID) (*domain.Account, error)
	Save(context.Context, *domain.Account) error
}

type DepositMoney struct {
	accounts AccountRepository
}

func NewDepositMoney(accounts AccountRepository) *DepositMoney {
	if accounts == nil {
		panic("application: nil account repository")
	}
	return &DepositMoney{accounts: accounts}
}

func (uc *DepositMoney) Execute(ctx context.Context, id domain.AccountID, amount int64) error {
	account, err := uc.accounts.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("load account: %w", err)
	}
	if err := account.Deposit(amount); err != nil {
		return fmt.Errorf("deposit: %w", err)
	}
	if err := uc.accounts.Save(ctx, account); err != nil {
		return fmt.Errorf("save account: %w", err)
	}
	return nil
}
