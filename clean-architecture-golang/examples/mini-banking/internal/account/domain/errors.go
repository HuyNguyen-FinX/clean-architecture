package domain

import "errors"

var (
	ErrInvalidAccountID     = errors.New("invalid account id")
	ErrInvalidTransferID    = errors.New("invalid transfer id")
	ErrInvalidTransferTime  = errors.New("invalid transfer time")
	ErrInvalidCurrency      = errors.New("invalid currency")
	ErrInvalidAmount        = errors.New("invalid amount")
	ErrMoneyOverflow        = errors.New("money arithmetic overflow")
	ErrCurrencyMismatch     = errors.New("currency mismatch")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrAccountNotFound      = errors.New("account not found")
	ErrAccountFrozen        = errors.New("account is frozen")
	ErrInvalidAccountStatus = errors.New("invalid account status")
	ErrSameAccountTransfer  = errors.New("cannot transfer to the same account")
	ErrInvalidOverdraftRule = errors.New("invalid overdraft rule")
)
