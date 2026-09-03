package domain

import (
	"fmt"
	"math"
	"strings"
)

type Currency string

func NewCurrency(raw string) (Currency, error) {
	normalized := strings.TrimSpace(strings.ToUpper(raw))
	if len(normalized) != 3 {
		return "", ErrInvalidCurrency
	}
	for _, char := range normalized {
		if char < 'A' || char > 'Z' {
			return "", ErrInvalidCurrency
		}
	}

	return Currency(normalized), nil
}

func (c Currency) String() string {
	return string(c)
}

func (c Currency) valid() bool {
	normalized, err := NewCurrency(string(c))
	return err == nil && normalized == c
}

type Money struct {
	amount   int64
	currency Currency
}

func NewMoney(amount int64, currency string) (Money, error) {
	parsedCurrency, err := NewCurrency(currency)
	if err != nil {
		return Money{}, err
	}

	return Money{amount: amount, currency: parsedCurrency}, nil
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

func (m Money) Currency() Currency {
	return m.currency
}

func (m Money) Equal(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
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

	amount, err := checkedAdd(m.amount, other.amount)
	if err != nil {
		return Money{}, err
	}

	return Money{amount: amount, currency: m.currency}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if err := m.ensureSameCurrency(other); err != nil {
		return Money{}, err
	}

	amount, err := checkedSub(m.amount, other.amount)
	if err != nil {
		return Money{}, err
	}

	return Money{amount: amount, currency: m.currency}, nil
}

func (m Money) Negate() (Money, error) {
	if err := m.validate(); err != nil {
		return Money{}, err
	}
	if m.amount == math.MinInt64 {
		return Money{}, ErrMoneyOverflow
	}

	return Money{amount: -m.amount, currency: m.currency}, nil
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
	if err := m.validate(); err != nil {
		return err
	}
	if err := other.validate(); err != nil {
		return err
	}
	if m.currency != other.currency {
		return ErrCurrencyMismatch
	}

	return nil
}

func (m Money) validate() error {
	if !m.currency.valid() {
		return ErrInvalidCurrency
	}
	return nil
}

func checkedAdd(left, right int64) (int64, error) {
	if right > 0 && left > math.MaxInt64-right {
		return 0, ErrMoneyOverflow
	}
	if right < 0 && left < math.MinInt64-right {
		return 0, ErrMoneyOverflow
	}

	return left + right, nil
}

func checkedSub(left, right int64) (int64, error) {
	if right > 0 && left < math.MinInt64+right {
		return 0, ErrMoneyOverflow
	}
	if right < 0 && left > math.MaxInt64+right {
		return 0, ErrMoneyOverflow
	}

	return left - right, nil
}
