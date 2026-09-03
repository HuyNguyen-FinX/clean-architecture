package domain

import "errors"

var (
	ErrInvalidAmount       = errors.New("amount must be positive")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type Account struct {
	id      string
	balance int64
}

func NewAccount(id string, balance int64) *Account {
	return &Account{id: id, balance: balance}
}

func (a *Account) ID() string     { return a.id }
func (a *Account) Balance() int64 { return a.balance }
func (a *Account) Clone() *Account {
	return &Account{id: a.id, balance: a.balance}
}

func (a *Account) Withdraw(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > a.balance {
		return ErrInsufficientBalance
	}
	a.balance -= amount
	return nil
}

func (a *Account) Deposit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	a.balance += amount
	return nil
}
