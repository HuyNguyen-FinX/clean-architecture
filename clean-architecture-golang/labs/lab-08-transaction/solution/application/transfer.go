package application

import (
	"context"
	"errors"

	"example.com/cleanarch/lab08/solution/domain"
)

var ErrInvalidCommand = errors.New("invalid transfer command")

type AccountRepository interface {
	FindByID(context.Context, string) (*domain.Account, error)
	Save(context.Context, *domain.Account) error
}

type Transactor interface {
	WithinTransaction(context.Context, func(context.Context) error) error
}

type TransferMoney struct {
	repo AccountRepository
	tx   Transactor
}

func NewTransferMoney(repo AccountRepository, tx Transactor) *TransferMoney {
	return &TransferMoney{repo: repo, tx: tx}
}

func (uc *TransferMoney) Execute(ctx context.Context, fromID, toID string, amount int64) error {
	if fromID == "" || toID == "" || fromID == toID || amount <= 0 {
		return ErrInvalidCommand
	}
	return uc.tx.WithinTransaction(ctx, func(txCtx context.Context) error {
		from, err := uc.repo.FindByID(txCtx, fromID)
		if err != nil {
			return err
		}
		to, err := uc.repo.FindByID(txCtx, toID)
		if err != nil {
			return err
		}
		if err := from.Withdraw(amount); err != nil {
			return err
		}
		if err := to.Deposit(amount); err != nil {
			return err
		}
		if err := uc.repo.Save(txCtx, from); err != nil {
			return err
		}
		return uc.repo.Save(txCtx, to)
	})
}
