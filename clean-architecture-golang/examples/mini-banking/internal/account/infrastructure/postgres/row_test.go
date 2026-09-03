package postgres

import (
	"errors"
	"testing"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

func TestAccountRowToDomain(t *testing.T) {
	account, err := (accountRow{
		id:             "A-100",
		balance:        -50,
		currency:       "VND",
		overdraftLimit: 100,
		status:         "active",
	}).toDomain()
	if err != nil {
		t.Fatal(err)
	}
	if account.Balance().Amount() != -50 || account.OverdraftLimit().Amount() != 100 {
		t.Fatalf("unexpected account: balance=%d overdraft=%d",
			account.Balance().Amount(), account.OverdraftLimit().Amount())
	}
}

func TestAccountRowRejectsCorruptPersistenceState(t *testing.T) {
	_, err := (accountRow{
		id:             "A-100",
		balance:        -101,
		currency:       "VND",
		overdraftLimit: 100,
		status:         "active",
	}).toDomain()
	if !errors.Is(err, domain.ErrInvalidOverdraftRule) {
		t.Fatalf("got %v, want ErrInvalidOverdraftRule", err)
	}
}
