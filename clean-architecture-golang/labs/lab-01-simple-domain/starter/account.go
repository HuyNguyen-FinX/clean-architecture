package account

import "errors"

var ErrInsufficientBalance = errors.New("insufficient balance")

type Account struct {
	ID             string
	Balance        int64
	OverdraftLimit int64
	Currency       string
}

func Withdraw(account *Account, amount int64, currency string) error {
	if account.Currency != currency {
		return errors.New("currency mismatch")
	}
	if account.Balance-amount < -account.OverdraftLimit {
		return ErrInsufficientBalance
	}

	account.Balance -= amount
	return nil
}
