package application_test

import (
	"context"
	"testing"

	"example.com/cleanarch/lab03/solution/application"
	"example.com/cleanarch/lab03/solution/domain"
	"example.com/cleanarch/lab03/solution/infrastructure/memory"
)

func TestDepositMoney(t *testing.T) {
	account, err := domain.NewAccount("A", 100)
	if err != nil {
		t.Fatal(err)
	}
	repo, err := memory.New(account)
	if err != nil {
		t.Fatal(err)
	}
	if err := application.NewDepositMoney(repo).Execute(context.Background(), "A", 50); err != nil {
		t.Fatal(err)
	}
	got, err := repo.FindByID(context.Background(), "A")
	if err != nil {
		t.Fatal(err)
	}
	if got.Balance() != 150 {
		t.Fatalf("balance=%d, want 150", got.Balance())
	}
}
