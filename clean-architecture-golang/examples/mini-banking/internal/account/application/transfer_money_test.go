package application_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/infrastructure/memory"
)

func TestTransferPersistsAtomicArtifactsAndReplays(t *testing.T) {
	repo := memory.NewRepository(
		mustAccount(t, "A-100", 1_000_000),
		mustAccount(t, "B-200", 250_000),
	)
	uc := application.NewTransferMoneyUseCase(
		repo, repo, &sequenceIDs{}, fixedClock{time.Unix(1_000, 0)},
	)
	command := application.TransferMoneyCommand{
		IdempotencyKey: "KEY-1",
		FromAccountID:  "A-100",
		ToAccountID:    "B-200",
		Amount:         500_000,
		Currency:       "vnd",
	}

	first, err := uc.Execute(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := uc.Execute(context.Background(), command)
	if err != nil {
		t.Fatal(err)
	}
	if first.TransferID != replayed.TransferID || !replayed.Replayed {
		t.Fatalf("first=%+v replayed=%+v", first, replayed)
	}

	snapshot := repo.Snapshot()
	if got := balance(snapshot, "A-100"); got != 500_000 {
		t.Fatalf("sender balance=%d", got)
	}
	if got := balance(snapshot, "B-200"); got != 750_000 {
		t.Fatalf("receiver balance=%d", got)
	}
	history := application.NewListTransfersUseCase(repo)
	items, err := history.Execute(context.Background(), "A-100", 100)
	if err != nil {
		t.Fatal(err)
	}
	pending, err := repo.PendingOutbox(context.Background(), 100)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || len(pending) != 1 {
		t.Fatalf("history=%d pending_outbox=%d", len(items), len(pending))
	}
}

func TestTransferRejectsChangedRequestForSameIdempotencyKey(t *testing.T) {
	repo := memory.NewRepository(
		mustAccount(t, "A-100", 1_000),
		mustAccount(t, "B-200", 0),
	)
	uc := application.NewTransferMoneyUseCase(
		repo, repo, &sequenceIDs{}, fixedClock{time.Unix(1_000, 0)},
	)
	command := application.TransferMoneyCommand{
		IdempotencyKey: "KEY-1", FromAccountID: "A-100", ToAccountID: "B-200",
		Amount: 100, Currency: "VND",
	}
	if _, err := uc.Execute(context.Background(), command); err != nil {
		t.Fatal(err)
	}
	command.Amount = 50
	_, err := uc.Execute(context.Background(), command)
	if !errors.Is(err, application.ErrIdempotencyConflict) {
		t.Fatalf("expected idempotency conflict, got %v", err)
	}
}

func TestRejectedTransferRollsBackIdempotencyAndArtifacts(t *testing.T) {
	repo := memory.NewRepository(
		mustAccount(t, "A-100", 100),
		mustAccount(t, "B-200", 0),
	)
	uc := application.NewTransferMoneyUseCase(
		repo, repo, &sequenceIDs{}, fixedClock{time.Unix(1_000, 0)},
	)
	command := application.TransferMoneyCommand{
		IdempotencyKey: "KEY-1", FromAccountID: "A-100", ToAccountID: "B-200",
		Amount: 200, Currency: "VND",
	}
	_, err := uc.Execute(context.Background(), command)
	if !errors.Is(err, domain.ErrInsufficientBalance) {
		t.Fatalf("expected insufficient balance, got %v", err)
	}
	command.Amount = 50
	if _, err := uc.Execute(context.Background(), command); err != nil {
		t.Fatalf("rolled-back key was not reusable: %v", err)
	}
}

func TestRelayMarksOnlySuccessfullyPublishedMessages(t *testing.T) {
	repo := memory.NewRepository(
		mustAccount(t, "A-100", 1_000),
		mustAccount(t, "B-200", 0),
	)
	clock := fixedClock{time.Unix(1_000, 0)}
	transfer := application.NewTransferMoneyUseCase(repo, repo, &sequenceIDs{}, clock)
	_, err := transfer.Execute(context.Background(), application.TransferMoneyCommand{
		IdempotencyKey: "KEY-1", FromAccountID: "A-100", ToAccountID: "B-200",
		Amount: 100, Currency: "VND",
	})
	if err != nil {
		t.Fatal(err)
	}

	want := errors.New("broker unavailable")
	failing := &fakePublisher{err: want}
	relay := application.NewRelayOutboxUseCase(repo, failing, clock)
	if _, err := relay.RunOnce(context.Background(), 10); !errors.Is(err, want) {
		t.Fatalf("expected publish error, got %v", err)
	}
	pending, _ := repo.PendingOutbox(context.Background(), 10)
	if len(pending) != 1 {
		t.Fatalf("failed publish was marked: pending=%d", len(pending))
	}

	success := &fakePublisher{}
	relay = application.NewRelayOutboxUseCase(repo, success, clock)
	count, err := relay.RunOnce(context.Background(), 10)
	if err != nil || count != 1 || len(success.messages) != 1 {
		t.Fatalf("count=%d published=%d err=%v", count, len(success.messages), err)
	}
	pending, _ = repo.PendingOutbox(context.Background(), 10)
	if len(pending) != 0 {
		t.Fatalf("published message remains pending: %d", len(pending))
	}
}

type fixedClock struct{ value time.Time }

func (c fixedClock) Now() time.Time { return c.value }

type sequenceIDs struct{ next int }

func (g *sequenceIDs) NewID() string {
	g.next++
	return fmt.Sprintf("ID-%03d", g.next)
}

type fakePublisher struct {
	err      error
	messages []application.OutboxMessage
}

func (p *fakePublisher) Publish(_ context.Context, message application.OutboxMessage) error {
	if p.err != nil {
		return p.err
	}
	p.messages = append(p.messages, message)
	return nil
}

func balance(accounts map[domain.AccountID]*domain.Account, rawID string) int64 {
	id, _ := domain.NewAccountID(rawID)
	return accounts[id].Balance().Amount()
}

func mustAccount(t *testing.T, rawID string, balance int64) *domain.Account {
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
