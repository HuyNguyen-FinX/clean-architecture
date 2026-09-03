package domain

import (
	"errors"
	"testing"
)

func TestAccountWithdrawProtectsOverdraftInvariant(t *testing.T) {
	account := newTestAccount(t, "A-100", 100_000, 50_000, "VND")
	amount := MustMoney(200_000, "VND")

	err := account.Withdraw(amount)

	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
	if account.Balance().Amount() != 100_000 {
		t.Fatalf("balance changed after rejected withdraw: %d", account.Balance().Amount())
	}
}

func TestAccountWithdrawAllowsBalanceUntilOverdraftLimit(t *testing.T) {
	account := newTestAccount(t, "A-100", 100_000, 50_000, "VND")
	amount := MustMoney(150_000, "VND")

	if err := account.Withdraw(amount); err != nil {
		t.Fatalf("withdraw failed: %v", err)
	}

	if account.Balance().Amount() != -50_000 {
		t.Fatalf("unexpected balance: %d", account.Balance().Amount())
	}
}

func TestAccountDepositRejectsDifferentCurrency(t *testing.T) {
	account := newTestAccount(t, "A-100", 100_000, 0, "VND")
	amount := MustMoney(1_000, "USD")

	err := account.Deposit(amount)

	if !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("expected ErrCurrencyMismatch, got %v", err)
	}
}

func newTestAccount(t *testing.T, id string, balanceAmount int64, overdraftAmount int64, currency string) *Account {
	t.Helper()

	accountID, err := NewAccountID(id)
	if err != nil {
		t.Fatal(err)
	}

	balance := MustMoney(balanceAmount, currency)
	overdraftLimit := MustMoney(overdraftAmount, currency)

	account, err := NewAccount(accountID, balance, overdraftLimit)
	if err != nil {
		t.Fatal(err)
	}

	return account
}
