package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"example.com/cleanarch/lab08/solution/application"
	"example.com/cleanarch/lab08/solution/domain"
	"example.com/cleanarch/lab08/solution/memory"
)

func TestTransferCommitsBothAccounts(t *testing.T) {
	store := memory.New(domain.NewAccount("A", 1_000), domain.NewAccount("B", 100))
	uc := application.NewTransferMoney(store, store)
	if err := uc.Execute(context.Background(), "A", "B", 300); err != nil {
		t.Fatal(err)
	}
	assertBalances(t, store, 700, 400)
}

func TestTransferRollsBackWhenSecondSaveFails(t *testing.T) {
	store := memory.New(domain.NewAccount("A", 1_000), domain.NewAccount("B", 100))
	want := errors.New("receiver save failed")
	repo := &failSecondSave{next: store, err: want}
	uc := application.NewTransferMoney(repo, store)
	err := uc.Execute(context.Background(), "A", "B", 300)
	if !errors.Is(err, want) {
		t.Fatalf("got %v", err)
	}
	assertBalances(t, store, 1_000, 100)
}

func TestConcurrentTransfersAreAtomic(t *testing.T) {
	store := memory.New(domain.NewAccount("A", 1_000), domain.NewAccount("B", 1_000))
	uc := application.NewTransferMoney(store, store)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := uc.Execute(context.Background(), "A", "B", 10); err != nil {
				t.Error(err)
			}
		}()
		go func() {
			defer wg.Done()
			if err := uc.Execute(context.Background(), "B", "A", 10); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	assertBalances(t, store, 1_000, 1_000)
}

type failSecondSave struct {
	next  application.AccountRepository
	calls int
	err   error
}

func (f *failSecondSave) FindByID(ctx context.Context, id string) (*domain.Account, error) {
	return f.next.FindByID(ctx, id)
}

func (f *failSecondSave) Save(ctx context.Context, account *domain.Account) error {
	f.calls++
	if f.calls == 2 {
		return f.err
	}
	return f.next.Save(ctx, account)
}

func assertBalances(t *testing.T, store *memory.Store, wantA, wantB int64) {
	t.Helper()
	snapshot := store.Snapshot()
	if snapshot["A"].Balance() != wantA || snapshot["B"].Balance() != wantB {
		t.Fatalf("A=%d B=%d, want A=%d B=%d",
			snapshot["A"].Balance(), snapshot["B"].Balance(), wantA, wantB)
	}
}
