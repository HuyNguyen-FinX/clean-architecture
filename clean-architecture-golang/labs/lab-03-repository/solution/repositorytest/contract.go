package repositorytest

import (
	"context"
	"errors"
	"testing"

	"example.com/cleanarch/lab03/solution/application"
	"example.com/cleanarch/lab03/solution/domain"
)

type Factory func(t *testing.T, accounts ...*domain.Account) application.AccountRepository

func Contract(t *testing.T, factory Factory) {
	t.Helper()

	t.Run("round trip and detached read", func(t *testing.T) {
		seed, err := domain.NewAccount("A", 100)
		if err != nil {
			t.Fatal(err)
		}
		repo := factory(t, seed)
		loaded, err := repo.FindByID(context.Background(), "A")
		if err != nil {
			t.Fatal(err)
		}
		if err := loaded.Deposit(50); err != nil {
			t.Fatal(err)
		}

		beforeSave, err := repo.FindByID(context.Background(), "A")
		if err != nil {
			t.Fatal(err)
		}
		if beforeSave.Balance() != 100 {
			t.Fatalf("store changed before Save: %d", beforeSave.Balance())
		}

		if err := repo.Save(context.Background(), loaded); err != nil {
			t.Fatal(err)
		}
		afterSave, err := repo.FindByID(context.Background(), "A")
		if err != nil {
			t.Fatal(err)
		}
		if afterSave.Balance() != 150 {
			t.Fatalf("balance=%d, want 150", afterSave.Balance())
		}

		if err := loaded.Deposit(25); err != nil {
			t.Fatal(err)
		}
		detached, _ := repo.FindByID(context.Background(), "A")
		if detached.Balance() != 150 {
			t.Fatalf("store retained Save pointer: %d", detached.Balance())
		}
	})

	t.Run("not found semantics", func(t *testing.T) {
		repo := factory(t)
		_, err := repo.FindByID(context.Background(), "missing")
		if !errors.Is(err, application.ErrAccountNotFound) {
			t.Fatalf("got %v, want ErrAccountNotFound", err)
		}
	})

	t.Run("canceled context", func(t *testing.T) {
		repo := factory(t)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := repo.FindByID(ctx, "A"); !errors.Is(err, context.Canceled) {
			t.Fatalf("find error=%v", err)
		}
		account, _ := domain.NewAccount("A", 1)
		if err := repo.Save(ctx, account); !errors.Is(err, context.Canceled) {
			t.Fatalf("save error=%v", err)
		}
	})
}
