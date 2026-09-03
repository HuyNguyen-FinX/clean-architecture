package memory

import (
	"context"
	"errors"
	"testing"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

var (
	_ application.AccountRepository = (*Repository)(nil)
	_ application.TransferStore     = (*Repository)(nil)
	_ application.OutboxStore       = (*Repository)(nil)
	_ application.Transactor        = (*Repository)(nil)
)

func TestRepositoryKeepsCopiesAtItsBoundary(t *testing.T) {
	seed := mustMemoryAccount(t, "A-100", 100_000)
	repo := NewRepository(seed)

	if err := seed.Deposit(domain.MustMoney(50_000, "VND")); err != nil {
		t.Fatal(err)
	}
	stored := findMemoryAccount(t, repo, "A-100")
	if stored.Balance().Amount() != 100_000 {
		t.Fatalf("repository retained caller pointer: %d", stored.Balance().Amount())
	}

	if err := stored.Deposit(domain.MustMoney(25_000, "VND")); err != nil {
		t.Fatal(err)
	}
	again := findMemoryAccount(t, repo, "A-100")
	if again.Balance().Amount() != 100_000 {
		t.Fatalf("FindByID leaked stored pointer: %d", again.Balance().Amount())
	}
}

func TestRepositorySavePersistsACopy(t *testing.T) {
	repo := NewRepository()
	account := mustMemoryAccount(t, "A-100", 100_000)

	if err := repo.Save(context.Background(), account); err != nil {
		t.Fatal(err)
	}
	if err := account.Deposit(domain.MustMoney(10_000, "VND")); err != nil {
		t.Fatal(err)
	}

	stored := findMemoryAccount(t, repo, "A-100")
	if stored.Balance().Amount() != 100_000 {
		t.Fatalf("Save retained caller pointer: %d", stored.Balance().Amount())
	}
}

func TestRepositoryReturnsNotFound(t *testing.T) {
	repo := NewRepository()
	id, err := domain.NewAccountID("missing")
	if err != nil {
		t.Fatal(err)
	}

	_, err = repo.FindByID(context.Background(), id)

	if !errors.Is(err, domain.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestRepositoryHonorsCanceledContext(t *testing.T) {
	repo := NewRepository(mustMemoryAccount(t, "A-100", 100_000))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	id, err := domain.NewAccountID("A-100")
	if err != nil {
		t.Fatal(err)
	}

	_, findErr := repo.FindByID(ctx, id)
	saveErr := repo.Save(ctx, mustMemoryAccount(t, "B-200", 0))

	if !errors.Is(findErr, context.Canceled) {
		t.Fatalf("FindByID: expected context.Canceled, got %v", findErr)
	}
	if !errors.Is(saveErr, context.Canceled) {
		t.Fatalf("Save: expected context.Canceled, got %v", saveErr)
	}
}

func TestRepositoryTransactionRollsBackAllStagedChanges(t *testing.T) {
	repo := NewRepository(mustMemoryAccount(t, "A-100", 100_000))
	want := errors.New("injected failure")
	err := repo.WithinTransaction(context.Background(), func(ctx context.Context) error {
		account := findMemoryAccountWithContext(t, ctx, repo, "A-100")
		if err := account.Deposit(domain.MustMoney(50_000, "VND")); err != nil {
			t.Fatal(err)
		}
		if err := repo.Save(ctx, account); err != nil {
			t.Fatal(err)
		}
		return want
	})
	if !errors.Is(err, want) {
		t.Fatalf("expected injected failure, got %v", err)
	}
	if got := findMemoryAccount(t, repo, "A-100").Balance().Amount(); got != 100_000 {
		t.Fatalf("rolled-back balance=%d", got)
	}
}

func findMemoryAccount(t *testing.T, repo *Repository, rawID string) *domain.Account {
	return findMemoryAccountWithContext(t, context.Background(), repo, rawID)
}

func findMemoryAccountWithContext(
	t *testing.T,
	ctx context.Context,
	repo *Repository,
	rawID string,
) *domain.Account {
	t.Helper()
	id, err := domain.NewAccountID(rawID)
	if err != nil {
		t.Fatal(err)
	}
	account, err := repo.FindByID(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	return account
}

func mustMemoryAccount(t *testing.T, rawID string, balance int64) *domain.Account {
	t.Helper()
	id, err := domain.NewAccountID(rawID)
	if err != nil {
		t.Fatal(err)
	}
	account, err := domain.NewAccount(
		id,
		domain.MustMoney(balance, "VND"),
		domain.MustMoney(0, "VND"),
	)
	if err != nil {
		t.Fatal(err)
	}
	return account
}
