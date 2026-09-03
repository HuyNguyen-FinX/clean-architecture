package account

import (
	"errors"
	"testing"
)

func TestWithdrawRejectsAmountBeyondOverdraftLimit(t *testing.T) {
	acc := NewAccount("A-100", mustMoney(t, 100_000, "VND"), mustMoney(t, 50_000, "VND"))
	amount := mustMoney(t, 200_000, "VND")

	err := acc.Withdraw(amount)

	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
	if acc.Balance().Amount() != 100_000 {
		t.Fatalf("balance should not change after rejected withdraw, got %d", acc.Balance().Amount())
	}
}

func TestWithdrawAllowsBalanceUntilOverdraftLimit(t *testing.T) {
	acc := NewAccount("A-100", mustMoney(t, 100_000, "VND"), mustMoney(t, 50_000, "VND"))
	amount := mustMoney(t, 150_000, "VND")

	if err := acc.Withdraw(amount); err != nil {
		t.Fatalf("withdraw failed: %v", err)
	}
	if acc.Balance().Amount() != -50_000 {
		t.Fatalf("unexpected balance: %d", acc.Balance().Amount())
	}
}

func TestDepositRejectsCurrencyMismatch(t *testing.T) {
	acc := NewAccount("A-100", mustMoney(t, 100_000, "VND"), mustMoney(t, 0, "VND"))
	amount := mustMoney(t, 10, "USD")

	err := acc.Deposit(amount)

	if !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("expected ErrCurrencyMismatch, got %v", err)
	}
}

func mustMoney(t *testing.T, amount int64, currency string) Money {
	t.Helper()

	money, err := NewMoney(amount, currency)
	if err != nil {
		t.Fatal(err)
	}

	return money
}
