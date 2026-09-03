package domain

import "errors"

var (
	ErrInvalidAmount = errors.New("invalid amount")
	ErrRejected      = errors.New("transfer rejected")
)

type Account struct {
	id      string
	balance int64
	frozen  bool
}

func NewAccount(id string, balance int64) *Account { return &Account{id: id, balance: balance} }
func (a *Account) ID() string                      { return a.id }
func (a *Account) Balance() int64                  { return a.balance }
func (a *Account) Clone() *Account {
	return &Account{id: a.id, balance: a.balance, frozen: a.frozen}
}

func (a *Account) Withdraw(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.frozen || amount > a.balance {
		return ErrRejected
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
