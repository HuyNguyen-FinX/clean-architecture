package domain

import (
	"fmt"
	"strings"
)

type Money struct {
	amount   int64
	currency string
}

func NewMoney(amount int64, currency string) (Money, error) {
	currency = strings.TrimSpace(strings.ToUpper(currency))
	if currency == "" {
		return Money{}, ErrInvalidCurrency
	}

	return Money{amount: amount, currency: currency}, nil
}

func NewPositiveMoney(amount int64, currency string) (Money, error) {
	if amount <= 0 {
		return Money{}, ErrInvalidAmount
	}

	return NewMoney(amount, currency)
}

func MustMoney(amount int64, currency string) Money {
	money, err := NewMoney(amount, currency)
	if err != nil {
		panic(err)
	}

	return money
}

func (m Money) Amount() int64 {
	return m.amount
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) IsPositive() bool {
	return m.amount > 0
}

func (m Money) IsNegative() bool {
	return m.amount < 0
}

func (m Money) Add(other Money) (Money, error) {
	if err := m.ensureSameCurrency(other); err != nil {
		return Money{}, err
	}

	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if err := m.ensureSameCurrency(other); err != nil {
		return Money{}, err
	}

	return Money{amount: m.amount - other.amount, currency: m.currency}, nil
}

func (m Money) Negate() Money {
	return Money{amount: -m.amount, currency: m.currency}
}

func (m Money) LessThan(other Money) (bool, error) {
	if err := m.ensureSameCurrency(other); err != nil {
		return false, err
	}

	return m.amount < other.amount, nil
}

func (m Money) Format() string {
	return fmt.Sprintf("%d %s", m.amount, m.currency)
}

func (m Money) ensureSameCurrency(other Money) error {
	if m.currency != other.currency {
		return ErrCurrencyMismatch
	}

	return nil
}
