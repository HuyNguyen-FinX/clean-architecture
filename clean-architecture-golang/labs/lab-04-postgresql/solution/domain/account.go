package domain

import "errors"

var (
	ErrInvalidAccount  = errors.New("invalid account")
	ErrInvalidCurrency = errors.New("invalid currency")
)

type Account struct {
	id        string
	balance   int64
	currency  string
	overdraft int64
	status    string
}

func RehydrateAccount(id string, balance int64, currency string, overdraft int64, status string) (*Account, error) {
	if id == "" || overdraft < 0 || balance < -overdraft {
		return nil, ErrInvalidAccount
	}
	if len(currency) != 3 {
		return nil, ErrInvalidCurrency
	}
	if status != "active" && status != "frozen" {
		return nil, ErrInvalidAccount
	}
	return &Account{id: id, balance: balance, currency: currency, overdraft: overdraft, status: status}, nil
}

func (a *Account) ID() string       { return a.id }
func (a *Account) Balance() int64   { return a.balance }
func (a *Account) Currency() string { return a.currency }
func (a *Account) Overdraft() int64 { return a.overdraft }
func (a *Account) Status() string   { return a.status }
