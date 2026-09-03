package domain

import (
	"errors"
	"math"
	"testing"
)

func TestNewMoneyNormalizesCurrency(t *testing.T) {
	money, err := NewMoney(500_000, " vnd ")
	if err != nil {
		t.Fatal(err)
	}

	if money.Currency() != Currency("VND") {
		t.Fatalf("expected VND, got %q", money.Currency())
	}
}

func TestNewMoneyRejectsInvalidCurrency(t *testing.T) {
	for _, currency := range []string{"", "VN", "VN1", "EURO"} {
		_, err := NewMoney(100, currency)
		if !errors.Is(err, ErrInvalidCurrency) {
			t.Fatalf("currency %q: expected ErrInvalidCurrency, got %v", currency, err)
		}
	}
}

func TestMoneyOperationsRejectNonCanonicalCurrencyState(t *testing.T) {
	invalid := Money{amount: 100, currency: Currency("vnd")}

	_, err := invalid.Add(MustMoney(100, "VND"))

	if !errors.Is(err, ErrInvalidCurrency) {
		t.Fatalf("expected ErrInvalidCurrency, got %v", err)
	}
}

func TestMoneyHasValueEqualityAndImmutableArithmetic(t *testing.T) {
	left := MustMoney(100_000, "VND")
	right := MustMoney(50_000, "VND")

	sum, err := left.Add(right)
	if err != nil {
		t.Fatal(err)
	}

	if !sum.Equal(MustMoney(150_000, "VND")) {
		t.Fatalf("unexpected sum: %s", sum.Format())
	}
	if !left.Equal(MustMoney(100_000, "VND")) {
		t.Fatalf("Add mutated receiver: %s", left.Format())
	}
}

func TestMoneyArithmeticRejectsCurrencyMismatch(t *testing.T) {
	_, err := MustMoney(100, "VND").Add(MustMoney(100, "USD"))

	if !errors.Is(err, ErrCurrencyMismatch) {
		t.Fatalf("expected ErrCurrencyMismatch, got %v", err)
	}
}

func TestMoneyAdditionRejectsOverflow(t *testing.T) {
	_, err := MustMoney(math.MaxInt64, "VND").Add(MustMoney(1, "VND"))

	if !errors.Is(err, ErrMoneyOverflow) {
		t.Fatalf("expected ErrMoneyOverflow, got %v", err)
	}
}

func TestMoneySubtractionRejectsUnderflow(t *testing.T) {
	_, err := MustMoney(math.MinInt64, "VND").Sub(MustMoney(1, "VND"))

	if !errors.Is(err, ErrMoneyOverflow) {
		t.Fatalf("expected ErrMoneyOverflow, got %v", err)
	}
}

func TestMoneyNegateRejectsMinimumInt64(t *testing.T) {
	_, err := MustMoney(math.MinInt64, "VND").Negate()

	if !errors.Is(err, ErrMoneyOverflow) {
		t.Fatalf("expected ErrMoneyOverflow, got %v", err)
	}
}
