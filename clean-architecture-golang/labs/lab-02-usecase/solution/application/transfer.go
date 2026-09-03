package application

import (
	"context"
	"errors"
	"fmt"

	"example.com/cleanarch/lab02/solution/domain"
)

var ErrInvalidCommand = errors.New("invalid transfer command")

type AccountRepository interface {
	FindByID(context.Context, domain.AccountID) (*domain.Account, error)
	Save(context.Context, *domain.Account) error
}

type Transactor interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}

type TransferCommand struct {
	From   domain.AccountID
	To     domain.AccountID
	Amount int64
}

type TransferMoney struct {
	accounts AccountRepository
	tx       Transactor
}

func NewTransferMoney(accounts AccountRepository, tx Transactor) *TransferMoney {
	if accounts == nil {
		panic("application: nil account repository")
	}
	if tx == nil {
		panic("application: nil transactor")
	}
	return &TransferMoney{accounts: accounts, tx: tx}
}

func (uc *TransferMoney) Execute(ctx context.Context, cmd TransferCommand) error {
	if cmd.From == "" || cmd.To == "" || cmd.From == cmd.To || cmd.Amount <= 0 {
		return ErrInvalidCommand
	}

	return uc.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		from, err := uc.accounts.FindByID(txCtx, cmd.From)
		if err != nil {
			return fmt.Errorf("load source: %w", err)
		}
		to, err := uc.accounts.FindByID(txCtx, cmd.To)
		if err != nil {
			return fmt.Errorf("load destination: %w", err)
		}
		if err := from.Withdraw(cmd.Amount); err != nil {
			return fmt.Errorf("withdraw: %w", err)
		}
		if err := to.Deposit(cmd.Amount); err != nil {
			return fmt.Errorf("deposit: %w", err)
		}
		if err := uc.accounts.Save(txCtx, from); err != nil {
			return fmt.Errorf("save source: %w", err)
		}
		if err := uc.accounts.Save(txCtx, to); err != nil {
			return fmt.Errorf("save destination: %w", err)
		}
		return nil
	})
}
