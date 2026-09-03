package domain

import (
	"errors"
	"fmt"
)

type AccountID string

var (
	ErrInvalidAccount      = errors.New("invalid account")
	ErrInvalidAmount       = errors.New("amount must be positive")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

type Account struct {
	id      AccountID
	balance int64
}

func NewAccount(id AccountID, balance int64) (*Account, error) {
	if id == "" || balance < 0 {
		return nil, ErrInvalidAccount
	}
	return &Account{id: id, balance: balance}, nil
}

func (a *Account) ID() AccountID { return a.id }

func (a *Account) Balance() int64 { return a.balance }

func (a *Account) Withdraw(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if amount > a.balance {
		return fmt.Errorf("balance %d, amount %d: %w", a.balance, amount, ErrInsufficientBalance)
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
