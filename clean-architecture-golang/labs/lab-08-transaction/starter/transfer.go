package starter

import "errors"

var ErrReceiverSave = errors.New("receiver save failed")

type Account struct {
	ID      string
	Balance int64
}

type Store struct {
	Accounts map[string]*Account
}

func (s *Store) Transfer(fromID, toID string, amount int64) error {
	from := *s.Accounts[fromID]
	to := *s.Accounts[toID]
	from.Balance -= amount
	s.Accounts[fromID] = &from
	to.Balance += amount
	return ErrReceiverSave
}
