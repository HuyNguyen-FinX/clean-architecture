package application

import (
	"context"
	"errors"
	"testing"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/infrastructure/memory"
)

func TestTransferMoneyMovesMoneyBetweenAccounts(t *testing.T) {
	repo := memory.NewRepository(
		mustAccount(t, "A-100", 1_000_000, 0, "VND"),
		mustAccount(t, "B-200", 250_000, 0, "VND"),
	)
	uc := NewTransferMoneyUseCase(repo, NoopTransactor{})

	err := uc.Execute(context.Background(), TransferMoneyCommand{
		FromAccountID: "A-100",
		ToAccountID:   "B-200",
		Amount:        500_000,
		Currency:      "VND",
	})
	if err != nil {
		t.Fatalf("transfer failed: %v", err)
	}

	sender := findAccount(t, repo, "A-100")
	receiver := findAccount(t, repo, "B-200")

	if sender.Balance().Amount() != 500_000 {
		t.Fatalf("unexpected sender balance: %d", sender.Balance().Amount())
	}
	if receiver.Balance().Amount() != 750_000 {
		t.Fatalf("unexpected receiver balance: %d", receiver.Balance().Amount())
	}
}

func TestTransferMoneyRejectsInsufficientBalance(t *testing.T) {
	repo := memory.NewRepository(
		mustAccount(t, "A-100", 100_000, 0, "VND"),
		mustAccount(t, "B-200", 250_000, 0, "VND"),
	)
	uc := NewTransferMoneyUseCase(repo, NoopTransactor{})

	err := uc.Execute(context.Background(), TransferMoneyCommand{
		FromAccountID: "A-100",
		ToAccountID:   "B-200",
		Amount:        500_000,
		Currency:      "VND",
	})

	if !errors.Is(err, domain.ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}

	sender := findAccount(t, repo, "A-100")
	if sender.Balance().Amount() != 100_000 {
		t.Fatalf("sender balance changed after rejected transfer: %d", sender.Balance().Amount())
	}
}

func mustAccount(t *testing.T, id string, balanceAmount int64, overdraftAmount int64, currency string) *domain.Account {
	t.Helper()

	accountID, err := domain.NewAccountID(id)
	if err != nil {
		t.Fatal(err)
	}

	balance := domain.MustMoney(balanceAmount, currency)
	overdraftLimit := domain.MustMoney(overdraftAmount, currency)

	account, err := domain.NewAccount(accountID, balance, overdraftLimit)
	if err != nil {
		t.Fatal(err)
	}

	return account
}

func findAccount(t *testing.T, repo *memory.Repository, id string) *domain.Account {
	t.Helper()

	accountID, err := domain.NewAccountID(id)
	if err != nil {
		t.Fatal(err)
	}

	account, err := repo.FindByID(context.Background(), accountID)
	if err != nil {
		t.Fatal(err)
	}

	return account
}
