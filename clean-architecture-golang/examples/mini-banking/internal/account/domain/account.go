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

func NewAccount(id AccountID, balance Money, overdraftLimit Money) (*Account, error) {
	return RehydrateAccount(id, balance, overdraftLimit, AccountStatusActive)
}

func RehydrateAccount(
	id AccountID,
	balance Money,
	overdraftLimit Money,
	status AccountStatus,
) (*Account, error) {
	if id == "" {
		return nil, ErrInvalidAccountID
	}
	if err := balance.validate(); err != nil {
		return nil, err
	}
	if err := overdraftLimit.validate(); err != nil {
		return nil, err
	}
	if balance.Currency() != overdraftLimit.Currency() {
		return nil, ErrCurrencyMismatch
	}
	if overdraftLimit.IsNegative() {
		return nil, ErrInvalidOverdraftRule
	}
	if status != AccountStatusActive && status != AccountStatusFrozen {
		return nil, ErrInvalidAccountStatus
	}

	minimumBalance, err := overdraftLimit.Negate()
	if err != nil {
		return nil, err
	}
	tooLow, err := balance.LessThan(minimumBalance)
	if err != nil {
		return nil, err
	}
	if tooLow {
		return nil, ErrInvalidOverdraftRule
	}

	return &Account{
		id:             id,
		balance:        balance,
		overdraftLimit: overdraftLimit,
		status:         status,
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

func (a *Account) Status() AccountStatus {
	return a.status
}

func (a *Account) Freeze() {
	a.status = AccountStatusFrozen
}

func (a *Account) Activate() {
	a.status = AccountStatusActive
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

	minimumBalance, err := a.overdraftLimit.Negate()
	if err != nil {
		return err
	}
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
		status:         a.status,
	}
}
