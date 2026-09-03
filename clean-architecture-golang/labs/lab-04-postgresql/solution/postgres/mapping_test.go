package postgres

import (
	"errors"
	"testing"

	"example.com/cleanarch/lab04/solution/domain"
)

func TestRowMapping(t *testing.T) {
	account, err := (accountRow{"A", -50, "VND", 100, "active"}).toDomain()
	if err != nil {
		t.Fatal(err)
	}
	if account.Balance() != -50 {
		t.Fatalf("balance=%d", account.Balance())
	}
}

func TestRowMappingRejectsCorruptState(t *testing.T) {
	_, err := (accountRow{"A", -101, "VND", 100, "active"}).toDomain()
	if !errors.Is(err, domain.ErrInvalidAccount) {
		t.Fatalf("got %v", err)
	}
}
