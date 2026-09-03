package application

import (
	"context"
	"errors"

	"example.com/cleanarch/lab11/solution/domain"
)

var ErrNotFound = errors.New("account not found")

type Transaction interface {
	Find(string) (*domain.Account, error)
	Save(*domain.Account) error
	AppendOutbox(string) error
}

type UnitOfWork interface {
	Within(context.Context, func(Transaction) error) error
}

type Command struct {
	From, To string
	Amount   int64
}

type TransferMoney struct{ uow UnitOfWork }

func NewTransferMoney(uow UnitOfWork) *TransferMoney { return &TransferMoney{uow: uow} }

func (uc *TransferMoney) Execute(ctx context.Context, cmd Command) error {
	return uc.uow.Within(ctx, func(tx Transaction) error {
		from, err := tx.Find(cmd.From)
		if err != nil {
			return err
		}
		to, err := tx.Find(cmd.To)
		if err != nil {
			return err
		}
		if err := from.Withdraw(cmd.Amount); err != nil {
			return err
		}
		if err := to.Deposit(cmd.Amount); err != nil {
			return err
		}
		if err := tx.Save(from); err != nil {
			return err
		}
		if err := tx.Save(to); err != nil {
			return err
		}
		return tx.AppendOutbox(cmd.From + "->" + cmd.To)
	})
}
