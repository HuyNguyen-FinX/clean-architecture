package application

import (
	"context"

	"example.com/cleanarch/lab07/solution/domain"
)

type Repository interface {
	Find(context.Context, string) (*domain.Account, error)
	Save(context.Context, *domain.Account) error
}

type Transfer struct{ repo Repository }

func NewTransfer(repo Repository) *Transfer { return &Transfer{repo: repo} }

func (uc *Transfer) Execute(ctx context.Context, fromID, toID string, amount int64) error {
	from, err := uc.repo.Find(ctx, fromID)
	if err != nil {
		return err
	}
	to, err := uc.repo.Find(ctx, toID)
	if err != nil {
		return err
	}
	if err := from.Withdraw(amount); err != nil {
		return err
	}
	if err := to.Deposit(amount); err != nil {
		return err
	}
	if err := uc.repo.Save(ctx, from); err != nil {
		return err
	}
	return uc.repo.Save(ctx, to)
}
