package domain

import (
	"errors"
	"testing"
)

func TestWithdrawProtectsInvariant(t *testing.T) {
	account := NewAccount("A", 100)
	err := account.Withdraw(101)
	if !errors.Is(err, ErrRejected) || account.Balance() != 100 {
		t.Fatalf("err=%v balance=%d", err, account.Balance())
	}
}
