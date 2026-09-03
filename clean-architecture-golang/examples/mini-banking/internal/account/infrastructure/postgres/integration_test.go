package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/infrastructure/postgres"
)

func TestPostgresTransferCommitAndRollback(t *testing.T) {
	pool := testPool(t)
	seed(t, pool)
	repo := postgres.NewRepository(pool)
	tx := postgres.NewTransactor(pool)

	uc := newPostgresTransfer(repo, tx)
	result, err := uc.Execute(context.Background(), application.TransferMoneyCommand{
		IdempotencyKey: "commit-1",
		FromAccountID:  "A-100",
		ToAccountID:    "B-200",
		Amount:         300,
		Currency:       "VND",
	})
	if err != nil {
		t.Fatalf("commit transfer: %v", err)
	}
	assertBalance(t, repo, "A-100", 700)
	assertBalance(t, repo, "B-200", 400)
	if result.TransferID == "" {
		t.Fatal("transfer id is empty")
	}
	history, err := repo.ListTransfersByAccount(context.Background(), "A-100", 100)
	if err != nil || len(history) != 1 {
		t.Fatalf("history=%d err=%v", len(history), err)
	}
	pending, err := repo.PendingOutbox(context.Background(), 100)
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending outbox=%d err=%v", len(pending), err)
	}
	replay, err := uc.Execute(context.Background(), application.TransferMoneyCommand{
		IdempotencyKey: "commit-1",
		FromAccountID:  "A-100", ToAccountID: "B-200", Amount: 300, Currency: "VND",
	})
	if err != nil || !replay.Replayed || replay.TransferID != result.TransferID {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}

	want := errors.New("injected receiver save failure")
	failing := &failSecondSave{TransferStore: repo, err: want}
	uc = newPostgresTransfer(failing, tx)
	_, err = uc.Execute(context.Background(), application.TransferMoneyCommand{
		IdempotencyKey: "rollback-1",
		FromAccountID:  "A-100",
		ToAccountID:    "B-200",
		Amount:         200,
		Currency:       "VND",
	})
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want injected error", err)
	}
	assertBalance(t, repo, "A-100", 700)
	assertBalance(t, repo, "B-200", 400)
}

func TestPostgresOppositeTransfersUseStableLockOrder(t *testing.T) {
	pool := testPool(t)
	seed(t, pool)
	repo := postgres.NewRepository(pool)
	uc := newPostgresTransfer(repo, postgres.NewTransactor(pool))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	commands := []application.TransferMoneyCommand{
		{IdempotencyKey: "opposite-1", FromAccountID: "A-100", ToAccountID: "B-200", Amount: 100, Currency: "VND"},
		{IdempotencyKey: "opposite-2", FromAccountID: "B-200", ToAccountID: "A-100", Amount: 50, Currency: "VND"},
	}
	var wg sync.WaitGroup
	errs := make(chan error, len(commands))
	for _, command := range commands {
		command := command
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := uc.Execute(ctx, command)
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent transfer: %v", err)
		}
	}
	assertBalance(t, repo, "A-100", 950)
	assertBalance(t, repo, "B-200", 1_050)
}

func TestPostgresConcurrentDuplicateRequestHasOneBusinessEffect(t *testing.T) {
	pool := testPool(t)
	seed(t, pool)
	repo := postgres.NewRepository(pool)
	uc := newPostgresTransfer(repo, postgres.NewTransactor(pool))
	command := application.TransferMoneyCommand{
		IdempotencyKey: "same-key", FromAccountID: "A-100", ToAccountID: "B-200",
		Amount: 100, Currency: "VND",
	}

	results := make(chan application.TransferMoneyResult, 2)
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := uc.Execute(context.Background(), command)
			results <- result
			errs <- err
		}()
	}
	wg.Wait()
	close(results)
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent duplicate: %v", err)
		}
	}
	var transferID domain.TransferID
	replayed := 0
	for result := range results {
		if transferID == "" {
			transferID = result.TransferID
		}
		if result.TransferID != transferID {
			t.Fatalf("different transfer IDs: %q and %q", transferID, result.TransferID)
		}
		if result.Replayed {
			replayed++
		}
	}
	if replayed != 1 {
		t.Fatalf("replayed responses=%d, want 1", replayed)
	}
	assertBalance(t, repo, "A-100", 900)
	assertBalance(t, repo, "B-200", 200)
	history, err := repo.ListTransfersByAccount(context.Background(), "A-100", 100)
	if err != nil || len(history) != 1 {
		t.Fatalf("history=%d err=%v", len(history), err)
	}
}

type failSecondSave struct {
	application.TransferStore
	calls int
	err   error
}

func (f *failSecondSave) Save(ctx context.Context, account *domain.Account) error {
	f.calls++
	if f.calls == 2 {
		return f.err
	}
	return f.TransferStore.Save(ctx, account)
}

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := pool.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	if err := postgres.Migrate(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, "TRUNCATE outbox, idempotency_keys, transfers, accounts"); err != nil {
		t.Fatal(err)
	}
	return pool
}

func newPostgresTransfer(
	store application.TransferStore,
	tx application.Transactor,
) *application.TransferMoneyUseCase {
	return application.NewTransferMoneyUseCase(store, tx, &testIDs{}, testClock{})
}

type testIDs struct{ next atomic.Uint64 }

func (g *testIDs) NewID() string {
	return fmt.Sprintf("TEST-%06d", g.next.Add(1))
}

type testClock struct{}

func (testClock) Now() time.Time { return time.Unix(1_000, 0).UTC() }

func seed(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	_, err := pool.Exec(context.Background(),
		"INSERT INTO accounts (id, balance_minor, currency, overdraft_limit_minor, status) "+
			"VALUES ('A-100', 1000, 'VND', 0, 'active'), ('B-200', 100, 'VND', 0, 'active')")
	if err != nil {
		t.Fatal(err)
	}
}

func assertBalance(t *testing.T, repo *postgres.Repository, rawID string, want int64) {
	t.Helper()
	id, err := domain.NewAccountID(rawID)
	if err != nil {
		t.Fatal(err)
	}
	account, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if got := account.Balance().Amount(); got != want {
		t.Fatalf("%s balance=%d, want %d", rawID, got, want)
	}
}
