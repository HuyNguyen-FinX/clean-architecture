package account

import (
	"errors"
	"testing"
)

func TestMoneyNormalizesCurrencyAndHasValueEquality(t *testing.T) {
	left := mustMoney(t, 100_000, " vnd ")
	right := mustMoney(t, 100_000, "VND")

	if !left.Equal(right) {
		t.Fatalf("expected equal money values")
	}
}

func TestMoneyAddDoesNotMutateOperands(t *testing.T) {
	left := mustMoney(t, 100_000, "VND")
	right := mustMoney(t, 50_000, "VND")

	sum, err := left.Add(right)
	if err != nil {
		t.Fatal(err)
	}

	if sum.Amount() != 150_000 {
		t.Fatalf("unexpected sum: %d", sum.Amount())
	}
	if left.Amount() != 100_000 {
		t.Fatalf("left operand was mutated: %d", left.Amount())
	}
}

func TestNewAccountRejectsInvalidInitialState(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		balance Money
		limit   Money
		wantErr error
	}{
		{"empty id", "", mustMoney(t, 0, "VND"), mustMoney(t, 0, "VND"), ErrInvalidAccountID},
		{"zero value money", "A-100", Money{}, Money{}, ErrInvalidCurrency},
		{"currency mismatch", "A-100", mustMoney(t, 0, "VND"), mustMoney(t, 0, "USD"), ErrCurrencyMismatch},
		{"negative limit", "A-100", mustMoney(t, 0, "VND"), mustMoney(t, -1, "VND"), ErrInvalidOverdraftRule},
		{"balance below limit", "A-100", mustMoney(t, -50_001, "VND"), mustMoney(t, 50_000, "VND"), ErrInvalidOverdraftRule},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewAccount(test.id, test.balance, test.limit)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("expected %v, got %v", test.wantErr, err)
			}
		})
	}
}

func TestWithdrawBoundary(t *testing.T) {
	tests := []struct {
		name        string
		amount      int64
		wantBalance int64
		wantErr     error
	}{
		{"within limit", 149_999, -49_999, nil},
		{"exact limit", 150_000, -50_000, nil},
		{"beyond limit", 150_001, 100_000, ErrInsufficientBalance},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			account := mustAccount(t, "A-100", 100_000, 50_000, "VND")
			err := account.Withdraw(mustMoney(t, test.amount, "VND"))

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("expected %v, got %v", test.wantErr, err)
			}
			if account.Balance().Amount() != test.wantBalance {
				t.Fatalf("expected balance %d, got %d", test.wantBalance, account.Balance().Amount())
			}
		})
	}
}

func TestMovementRejectsNonPositiveAmount(t *testing.T) {
	for _, amount := range []int64{0, -1} {
		account := mustAccount(t, "A-100", 100_000, 0, "VND")
		movement := mustMoney(t, amount, "VND")

		if err := account.Withdraw(movement); !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("withdraw %d: expected ErrInvalidAmount, got %v", amount, err)
		}
		if err := account.Deposit(movement); !errors.Is(err, ErrInvalidAmount) {
			t.Fatalf("deposit %d: expected ErrInvalidAmount, got %v", amount, err)
		}
	}
}

func TestMovementRejectsCurrencyMismatchWithoutMutation(t *testing.T) {
	account := mustAccount(t, "A-100", 100_000, 0, "VND")

	err := account.Deposit(mustMoney(t, 10, "USD"))

	if !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("expected ErrCurrencyMismatch, got %v", err)
	}
	if account.Balance().Amount() != 100_000 {
		t.Fatalf("balance changed after rejected deposit: %d", account.Balance().Amount())
	}
}

func TestFrozenAccountRejectsWithdrawButAcceptsDeposit(t *testing.T) {
	account := mustAccount(t, "A-100", 100_000, 0, "VND")
	account.Freeze()

	if err := account.Withdraw(mustMoney(t, 10_000, "VND")); !errors.Is(err, ErrAccountFrozen) {
		t.Fatalf("expected ErrAccountFrozen, got %v", err)
	}
	if err := account.Deposit(mustMoney(t, 10_000, "VND")); err != nil {
		t.Fatalf("deposit failed: %v", err)
	}
	if account.Balance().Amount() != 110_000 {
		t.Fatalf("unexpected balance: %d", account.Balance().Amount())
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

func mustAccount(
	t *testing.T,
	id string,
	balanceAmount int64,
	overdraftAmount int64,
	currency string,
) *Account {
	t.Helper()
	account, err := NewAccount(
		id,
		mustMoney(t, balanceAmount, currency),
		mustMoney(t, overdraftAmount, currency),
	)
	if err != nil {
		t.Fatal(err)
	}
	return account
}
