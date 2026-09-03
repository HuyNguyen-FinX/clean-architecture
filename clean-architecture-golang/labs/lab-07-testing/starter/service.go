package starter

import "errors"

var ErrInsufficient = errors.New("insufficient balance")

type Account struct {
	ID      string
	Balance int64
}

func Transfer(accounts map[string]*Account, from, to string, amount int64) error {
	if accounts[from].Balance < amount {
		return ErrInsufficient
	}
	accounts[from].Balance -= amount
	accounts[to].Balance += amount
	return nil
}
