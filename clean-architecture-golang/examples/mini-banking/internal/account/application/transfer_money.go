package application

import (
	"context"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

type TransferMoneyCommand struct {
	FromAccountID string
	ToAccountID   string
	Amount        int64
	Currency      string
}

type TransferMoneyUseCase struct {
	accounts   AccountRepository
	transactor Transactor
}

func NewTransferMoneyUseCase(accounts AccountRepository, transactor Transactor) *TransferMoneyUseCase {
	if transactor == nil {
		transactor = NoopTransactor{}
	}

	return &TransferMoneyUseCase{
		accounts:   accounts,
		transactor: transactor,
	}
}

func (uc *TransferMoneyUseCase) Execute(ctx context.Context, cmd TransferMoneyCommand) error {
	fromID, err := domain.NewAccountID(cmd.FromAccountID)
	if err != nil {
		return err
	}

	toID, err := domain.NewAccountID(cmd.ToAccountID)
	if err != nil {
		return err
	}
	if fromID == toID {
		return domain.ErrSameAccountTransfer
	}

	amount, err := domain.NewPositiveMoney(cmd.Amount, cmd.Currency)
	if err != nil {
		return err
	}

	return uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		sender, err := uc.accounts.FindByID(txCtx, fromID)
		if err != nil {
			return err
		}

		receiver, err := uc.accounts.FindByID(txCtx, toID)
		if err != nil {
			return err
		}

		if err := sender.Withdraw(amount); err != nil {
			return err
		}
		if err := receiver.Deposit(amount); err != nil {
			return err
		}

		if err := uc.accounts.Save(txCtx, sender); err != nil {
			return err
		}

		return uc.accounts.Save(txCtx, receiver)
	})
}
