package account

import (
	"errors"
	"strings"
)

var (
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrInvalidCurrency     = errors.New("invalid currency")
	ErrCurrencyMismatch    = errors.New("currency mismatch")
	ErrInsufficientBalance = errors.New("insufficient balance")
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

func (m Money) Amount() int64 {
	return m.amount
}

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrCurrencyMismatch
	}

	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

func (m Money) Sub(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, ErrCurrencyMismatch
	}

	return Money{amount: m.amount - other.amount, currency: m.currency}, nil
}

func (m Money) Negate() Money {
	return Money{amount: -m.amount, currency: m.currency}
}

func (m Money) LessThan(other Money) (bool, error) {
	if m.currency != other.currency {
		return false, ErrCurrencyMismatch
	}

	return m.amount < other.amount, nil
}

type Account struct {
	id             string
	balance        Money
	overdraftLimit Money
}

func NewAccount(id string, balance Money, overdraftLimit Money) *Account {
	return &Account{
		id:             id,
		balance:        balance,
		overdraftLimit: overdraftLimit,
	}
}

func (a *Account) Balance() Money {
	return a.balance
}

func (a *Account) Withdraw(amount Money) error {
	if amount.amount <= 0 {
		return ErrInvalidAmount
	}

	next, err := a.balance.Sub(amount)
	if err != nil {
		return err
	}

	tooLow, err := next.LessThan(a.overdraftLimit.Negate())
	if err != nil {
		return err
	}
	if tooLow {
		return ErrInsufficientBalance
	}

	a.balance = next
	return nil
}

func (a *Account) Deposit(amount Money) error {
	if amount.amount <= 0 {
		return ErrInvalidAmount
	}

	next, err := a.balance.Add(amount)
	if err != nil {
		return err
	}

	a.balance = next
	return nil
}
