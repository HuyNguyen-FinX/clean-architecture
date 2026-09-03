package domain

import "errors"

var (
	ErrInvalidAccountID     = errors.New("invalid account id")
	ErrInvalidCurrency      = errors.New("invalid currency")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrCurrencyMismatch     = errors.New("currency mismatch")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrAccountNotFound      = errors.New("account not found")
	ErrSameAccountTransfer  = errors.New("cannot transfer to the same account")
	ErrInvalidOverdraftRule = errors.New("invalid overdraft rule")
)
