package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/cleanarch/lab12/solution/application"
	"example.com/cleanarch/lab12/solution/domain"
	"example.com/cleanarch/lab12/solution/memory"
	"example.com/cleanarch/lab12/solution/support"
)

func TestTransferIsAtomicAndIdempotent(t *testing.T) {
	store := memory.New(domain.NewAccount("A", 1_000), domain.NewAccount("B", 100))
	uc := application.NewTransferMoney(store, &support.IDs{}, support.Clock{Value: time.Unix(1_000, 0)})
	command := application.TransferCommand{IdempotencyKey: "KEY-1", From: "A", To: "B", Amount: 300}

	first, err := uc.Execute(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	second, err := uc.Execute(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || !second.Replayed {
		t.Fatalf("first=%+v second=%+v", first, second)
	}
	a, _ := store.Balance("A")
	b, _ := store.Balance("B")
	history, _ := store.ByAccount(context.Background(), "A")
	if a != 700 || b != 400 || len(history) != 1 || len(store.Outbox()) != 1 {
		t.Fatalf("A=%d B=%d history=%d outbox=%d", a, b, len(history), len(store.Outbox()))
	}
}

func TestIdempotencyConflictAndDomainRejectionLeaveNoArtifacts(t *testing.T) {
	store := memory.New(domain.NewAccount("A", 100), domain.NewAccount("B", 0))
	uc := application.NewTransferMoney(store, &support.IDs{}, support.Clock{Value: time.Unix(1_000, 0)})
	base := application.TransferCommand{IdempotencyKey: "KEY-1", From: "A", To: "B", Amount: 50}
	if _, err := uc.Execute(context.Background(), base); err != nil {
		t.Fatal(err)
	}
	changed := base
	changed.Amount = 10
	if _, err := uc.Execute(context.Background(), changed); !errors.Is(err, application.ErrIdempotencyConflict) {
		t.Fatalf("got %v", err)
	}

	rejected := application.TransferCommand{IdempotencyKey: "KEY-2", From: "A", To: "B", Amount: 100}
	if _, err := uc.Execute(context.Background(), rejected); !errors.Is(err, domain.ErrInsufficient) {
		t.Fatalf("got %v", err)
	}
	if len(store.Outbox()) != 1 {
		t.Fatalf("rejected transfer added outbox: %d", len(store.Outbox()))
	}
}
