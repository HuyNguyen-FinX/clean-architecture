package application

import (
	"context"
	"errors"
	"testing"

	"example.com/cleanarch/lab02/solution/domain"
)

type txMarker struct{}

type fakeRepository struct {
	accounts  map[domain.AccountID]*domain.Account
	saved     []domain.AccountID
	findErr   error
	saveErr   error
	outsideTx bool
}

func (f *fakeRepository) FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	if ctx.Value(txMarker{}) == nil {
		f.outsideTx = true
	}
	if f.findErr != nil {
		return nil, f.findErr
	}
	account, ok := f.accounts[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return account, nil
}

func (f *fakeRepository) Save(ctx context.Context, account *domain.Account) error {
	if ctx.Value(txMarker{}) == nil {
		f.outsideTx = true
	}
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, account.ID())
	return nil
}

type spyTransactor struct {
	calls int
	err   error
}

func (s *spyTransactor) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	s.calls++
	if s.err != nil {
		return s.err
	}
	return fn(context.WithValue(ctx, txMarker{}, true))
}

func account(t *testing.T, id domain.AccountID, balance int64) *domain.Account {
	t.Helper()
	a, err := domain.NewAccount(id, balance)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestTransferMoney(t *testing.T) {
	repo := &fakeRepository{accounts: map[domain.AccountID]*domain.Account{
		"A": account(t, "A", 1_000),
		"B": account(t, "B", 100),
	}}
	tx := &spyTransactor{}
	uc := NewTransferMoney(repo, tx)

	err := uc.Execute(context.Background(), TransferCommand{From: "A", To: "B", Amount: 300})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if repo.accounts["A"].Balance() != 700 || repo.accounts["B"].Balance() != 400 {
		t.Fatal("unexpected balances")
	}
	if tx.calls != 1 || repo.outsideTx {
		t.Fatalf("transaction calls=%d outsideTx=%v", tx.calls, repo.outsideTx)
	}
	if len(repo.saved) != 2 {
		t.Fatalf("saved %d accounts, want 2", len(repo.saved))
	}
}

func TestTransferMoneyRejectsBeforeTransaction(t *testing.T) {
	repo := &fakeRepository{accounts: map[domain.AccountID]*domain.Account{}}
	tx := &spyTransactor{}
	err := NewTransferMoney(repo, tx).Execute(context.Background(), TransferCommand{
		From: "A", To: "A", Amount: 100,
	})
	if !errors.Is(err, ErrInvalidCommand) || tx.calls != 0 {
		t.Fatalf("err=%v transaction calls=%d", err, tx.calls)
	}
}

func TestTransferMoneyDomainRejectionDoesNotSave(t *testing.T) {
	repo := &fakeRepository{accounts: map[domain.AccountID]*domain.Account{
		"A": account(t, "A", 100),
		"B": account(t, "B", 100),
	}}
	err := NewTransferMoney(repo, &spyTransactor{}).Execute(context.Background(), TransferCommand{
		From: "A", To: "B", Amount: 101,
	})
	if !errors.Is(err, domain.ErrInsufficientBalance) {
		t.Fatalf("got %v", err)
	}
	if len(repo.saved) != 0 {
		t.Fatalf("saved %d accounts", len(repo.saved))
	}
}

func TestTransferMoneyPropagatesInfrastructureErrors(t *testing.T) {
	want := errors.New("database unavailable")
	repo := &fakeRepository{accounts: map[domain.AccountID]*domain.Account{}, findErr: want}
	err := NewTransferMoney(repo, &spyTransactor{}).Execute(context.Background(), TransferCommand{
		From: "A", To: "B", Amount: 100,
	})
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want wrapped error", err)
	}
}

func TestNewTransferMoneyRequiresDependencies(t *testing.T) {
	tests := []struct {
		name string
		fn   func()
	}{
		{"repository", func() { NewTransferMoney(nil, &spyTransactor{}) }},
		{"transactor", func() { NewTransferMoney(&fakeRepository{}, nil) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected panic")
				}
			}()
			tt.fn()
		})
	}
}
