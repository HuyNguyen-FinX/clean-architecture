package account

import (
	"errors"
	"testing"
)

func TestWithdrawRejectsAmountBeyondOverdraftLimit(t *testing.T) {
	acc := &Account{
		ID:             "A-100",
		Balance:        100_000,
		OverdraftLimit: 50_000,
		Currency:       "VND",
	}

	err := Withdraw(acc, 200_000, "VND")

	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
	if acc.Balance != 100_000 {
		t.Fatalf("balance should not change after rejected withdraw, got %d", acc.Balance)
	}
}

func TestPublicBalanceCanBreakInvariant(t *testing.T) {
	acc := &Account{
		ID:             "A-100",
		Balance:        100_000,
		OverdraftLimit: 50_000,
		Currency:       "VND",
	}

	acc.Balance = -999_999

	if acc.Balance >= -acc.OverdraftLimit {
		t.Fatal("test setup is invalid")
	}
}
