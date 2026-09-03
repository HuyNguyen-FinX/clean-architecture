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

func TestNewAccountRejectsInitialBalanceBelowOverdraftLimit(t *testing.T) {
	id, err := NewAccountID("A-100")
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewAccount(id, MustMoney(-50_001, "VND"), MustMoney(50_000, "VND"))

	if !errors.Is(err, ErrInvalidOverdraftRule) {
		t.Fatalf("expected ErrInvalidOverdraftRule, got %v", err)
	}
}

func TestNewAccountRejectsInvalidZeroValueMoney(t *testing.T) {
	id, err := NewAccountID("A-100")
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewAccount(id, Money{}, Money{})

	if !errors.Is(err, ErrInvalidCurrency) {
		t.Fatalf("expected ErrInvalidCurrency, got %v", err)
	}
}

func TestFrozenAccountCannotWithdraw(t *testing.T) {
	account := newTestAccount(t, "A-100", 100_000, 0, "VND")
	account.Freeze()

	err := account.Withdraw(MustMoney(10_000, "VND"))

	if !errors.Is(err, ErrAccountFrozen) {
		t.Fatalf("expected ErrAccountFrozen, got %v", err)
	}
	if account.Balance().Amount() != 100_000 {
		t.Fatalf("balance changed after rejected withdraw: %d", account.Balance().Amount())
	}
}

func TestFrozenAccountCanReceiveDeposit(t *testing.T) {
	account := newTestAccount(t, "A-100", 100_000, 0, "VND")
	account.Freeze()

	if err := account.Deposit(MustMoney(10_000, "VND")); err != nil {
		t.Fatalf("deposit into frozen account failed: %v", err)
	}
	if account.Balance().Amount() != 110_000 {
		t.Fatalf("unexpected balance: %d", account.Balance().Amount())
	}
}

func TestAccountRejectsNonPositiveMovement(t *testing.T) {
	account := newTestAccount(t, "A-100", 100_000, 0, "VND")

	for _, amount := range []int64{0, -1} {
		amount := MustMoney(amount, "VND")
		if err := account.Withdraw(amount); !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("withdraw %d: expected ErrInvalidAmount, got %v", amount.Amount(), err)
		}
		if err := account.Deposit(amount); !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("deposit %d: expected ErrInvalidAmount, got %v", amount.Amount(), err)
		}
	}
}

func TestClonePreservesStatusWithoutSharingMutableState(t *testing.T) {
	account := newTestAccount(t, "A-100", 100_000, 0, "VND")
	account.Freeze()
	clone := account.Clone()

	clone.Activate()
	if clone.Status() != AccountStatusActive {
		t.Fatalf("clone status: expected active, got %s", clone.Status())
	}
	if account.Status() != AccountStatusFrozen {
		t.Fatalf("original status changed through clone: %s", account.Status())
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
