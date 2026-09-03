package account

import (
	"errors"
	"strings"
)

var (
	ErrInvalidAccountID     = errors.New("invalid account id")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrInvalidCurrency      = errors.New("invalid currency")
	ErrCurrencyMismatch     = errors.New("currency mismatch")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrInvalidOverdraftRule = errors.New("invalid overdraft rule")
	ErrAccountFrozen        = errors.New("account is frozen")
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

type Money struct {
	amount   int64
	currency Currency
}

func NewMoney(amount int64, currency string) (Money, error) {
	parsed, err := NewCurrency(currency)
	if err != nil {
		return Money{}, err
	}
	return Money{amount: amount, currency: parsed}, nil
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

func (m Money) ensureSameCurrency(other Money) error {
	if m.currency == "" || other.currency == "" {
		return ErrInvalidCurrency
	}
	if m.currency != other.currency {
		return ErrCurrencyMismatch
	}
	return nil
}

type AccountID string

type AccountStatus string

const (
	AccountStatusActive AccountStatus = "active"
	AccountStatusFrozen AccountStatus = "frozen"
)

type Account struct {
	id             AccountID
	balance        Money
	overdraftLimit Money
	status         AccountStatus
}

func NewAccount(id string, balance, overdraftLimit Money) (*Account, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrInvalidAccountID
	}
	if balance.currency == "" || overdraftLimit.currency == "" {
		return nil, ErrInvalidCurrency
	}
	if balance.currency != overdraftLimit.currency {
		return nil, ErrCurrencyMismatch
	}
	if overdraftLimit.IsNegative() {
		return nil, ErrInvalidOverdraftRule
	}

	tooLow, err := balance.LessThan(overdraftLimit.Negate())
	if err != nil {
		return nil, err
	}
	if tooLow {
		return nil, ErrInvalidOverdraftRule
	}

	return &Account{
		id:             AccountID(id),
		balance:        balance,
		overdraftLimit: overdraftLimit,
		status:         AccountStatusActive,
	}, nil
}

func (a *Account) ID() AccountID {
	return a.id
}

func (a *Account) Balance() Money {
	return a.balance
}

func (a *Account) Status() AccountStatus {
	return a.status
}

func (a *Account) Freeze() {
	a.status = AccountStatusFrozen
}

func (a *Account) Activate() {
	a.status = AccountStatusActive
}

func (a *Account) Withdraw(amount Money) error {
	if a.status == AccountStatusFrozen {
		return ErrAccountFrozen
	}
	if !amount.IsPositive() {
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
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}

	next, err := a.balance.Add(amount)
	if err != nil {
		return err
	}
	a.balance = next
	return nil
}
