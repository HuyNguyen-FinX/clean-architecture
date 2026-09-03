package starter

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidTransfer     = errors.New("invalid transfer")
	ErrAccountNotFound     = errors.New("account not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

// Account cố ý là anemic record để người học nhận diện và refactor.
type Account struct {
	ID      string
	Balance int64
}

// TransferService cố ý trộn storage với business mutation trong starter.
type TransferService struct {
	accounts map[string]*Account
}

func NewTransferService(accounts map[string]*Account) *TransferService {
	return &TransferService{accounts: accounts}
}

func (s *TransferService) Transfer(fromID, toID string, amount int64) error {
	if fromID == "" || toID == "" || fromID == toID || amount <= 0 {
		return ErrInvalidTransfer
	}

	from, ok := s.accounts[fromID]
	if !ok {
		return fmt.Errorf("source %q: %w", fromID, ErrAccountNotFound)
	}
	to, ok := s.accounts[toID]
	if !ok {
		return fmt.Errorf("destination %q: %w", toID, ErrAccountNotFound)
	}
	if from.Balance < amount {
		return ErrInsufficientBalance
	}

	from.Balance -= amount
	to.Balance += amount
	return nil
}
