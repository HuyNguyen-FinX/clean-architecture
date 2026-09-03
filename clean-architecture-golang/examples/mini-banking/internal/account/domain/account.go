package domain

import "strings"

type AccountID string

func NewAccountID(raw string) (AccountID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidAccountID
	}

	return AccountID(raw), nil
}

type Account struct {
	id             AccountID
	balance        Money
	overdraftLimit Money
}

func NewAccount(id AccountID, balance Money, overdraftLimit Money) (*Account, error) {
	if id == "" {
		return nil, ErrInvalidAccountID
	}
	if balance.Currency() != overdraftLimit.Currency() {
		return nil, ErrCurrencyMismatch
	}
	if overdraftLimit.IsNegative() {
		return nil, ErrInvalidOverdraftRule
	}

	return &Account{
		id:             id,
		balance:        balance,
		overdraftLimit: overdraftLimit,
	}, nil
}

func (a *Account) ID() AccountID {
	return a.id
}

func (a *Account) Balance() Money {
	return a.balance
}

func (a *Account) OverdraftLimit() Money {
	return a.overdraftLimit
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

func (a *Account) Withdraw(amount Money) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}

	next, err := a.balance.Sub(amount)
	if err != nil {
		return err
	}

	minimumBalance := a.overdraftLimit.Negate()
	tooLow, err := next.LessThan(minimumBalance)
	if err != nil {
		return err
	}
	if tooLow {
		return ErrInsufficientBalance
	}

	a.balance = next
	return nil
}

func (a *Account) Clone() *Account {
	return &Account{
		id:             a.id,
		balance:        a.balance,
		overdraftLimit: a.overdraftLimit,
	}
}
