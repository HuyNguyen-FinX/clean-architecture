package postgres

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"example.com/cleanarch/lab04/solution/application"
)

func TestRepositoryIntegration(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration test")
	}
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := Migrate(context.Background(), pool); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(), "TRUNCATE accounts"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(context.Background(),
		"INSERT INTO accounts(id,balance_minor,currency,overdraft_minor,status) VALUES('A',100,'VND',0,'active')",
	); err != nil {
		t.Fatal(err)
	}

	repo := New(pool)
	account, err := repo.FindByID(context.Background(), "A")
	if err != nil || account.Balance() != 100 {
		t.Fatalf("account=%v err=%v", account, err)
	}
	_, err = repo.FindByID(context.Background(), "missing")
	if !errors.Is(err, application.ErrAccountNotFound) {
		t.Fatalf("got %v", err)
	}

	_, err = pool.Exec(context.Background(),
		"INSERT INTO accounts(id,balance_minor,currency,overdraft_minor,status) VALUES('bad',-2,'VND',1,'active')")
	if err == nil {
		t.Fatal("database accepted invariant violation")
	}
}
