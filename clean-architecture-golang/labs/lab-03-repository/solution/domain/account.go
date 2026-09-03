package domain

import "errors"

type AccountID string

var (
	ErrInvalidAccount = errors.New("invalid account")
	ErrInvalidAmount  = errors.New("amount must be positive")
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

func RehydrateAccount(id AccountID, balance int64) (*Account, error) {
	return NewAccount(id, balance)
}

func (a *Account) ID() AccountID { return a.id }

func (a *Account) Balance() int64 { return a.balance }

func (a *Account) Deposit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	a.balance += amount
	return nil
}
