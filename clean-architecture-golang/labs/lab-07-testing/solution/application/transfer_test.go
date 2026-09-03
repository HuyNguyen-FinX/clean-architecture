package application

import (
	"context"
	"errors"
	"testing"

	"example.com/cleanarch/lab07/solution/domain"
)

type fakeRepo struct {
	accounts map[string]*domain.Account
	saved    []string
}

func (f *fakeRepo) Find(_ context.Context, id string) (*domain.Account, error) {
	account, ok := f.accounts[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return account.Clone(), nil
}

func (f *fakeRepo) Save(_ context.Context, account *domain.Account) error {
	f.accounts[account.ID()] = account.Clone()
	f.saved = append(f.saved, account.ID())
	return nil
}

func TestTransferSavesBoth(t *testing.T) {
	repo := &fakeRepo{accounts: map[string]*domain.Account{
		"A": domain.NewAccount("A", 100), "B": domain.NewAccount("B", 0),
	}}
	if err := NewTransfer(repo).Execute(context.Background(), "A", "B", 40); err != nil {
		t.Fatal(err)
	}
	if len(repo.saved) != 2 || repo.accounts["A"].Balance() != 60 || repo.accounts["B"].Balance() != 40 {
		t.Fatalf("saved=%v", repo.saved)
	}
}

func TestTransferDoesNotSaveRejectedDomainOperation(t *testing.T) {
	repo := &fakeRepo{accounts: map[string]*domain.Account{
		"A": domain.NewAccount("A", 10), "B": domain.NewAccount("B", 0),
	}}
	err := NewTransfer(repo).Execute(context.Background(), "A", "B", 11)
	if !errors.Is(err, domain.ErrInsufficient) || len(repo.saved) != 0 {
		t.Fatalf("err=%v saved=%v", err, repo.saved)
	}
}
