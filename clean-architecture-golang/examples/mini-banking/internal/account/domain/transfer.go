package domain

import (
	"strings"
	"time"
)

type TransferID string

func NewTransferID(raw string) (TransferID, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ErrInvalidTransferID
	}
	return TransferID(raw), nil
}

type Transfer struct {
	id        TransferID
	from      AccountID
	to        AccountID
	amount    Money
	createdAt time.Time
}

func NewTransfer(
	id TransferID,
	from AccountID,
	to AccountID,
	amount Money,
	createdAt time.Time,
) (*Transfer, error) {
	if id == "" {
		return nil, ErrInvalidTransferID
	}
	if from == "" || to == "" {
		return nil, ErrInvalidAccountID
	}
	if from == to {
		return nil, ErrSameAccountTransfer
	}
	if !amount.IsPositive() {
		return nil, ErrInvalidAmount
	}
	if err := amount.validate(); err != nil {
		return nil, err
	}
	if createdAt.IsZero() {
		return nil, ErrInvalidTransferTime
	}
	return &Transfer{id: id, from: from, to: to, amount: amount, createdAt: createdAt.UTC()}, nil
}

func (t *Transfer) ID() TransferID       { return t.id }
func (t *Transfer) From() AccountID      { return t.from }
func (t *Transfer) To() AccountID        { return t.to }
func (t *Transfer) Amount() Money        { return t.amount }
func (t *Transfer) CreatedAt() time.Time { return t.createdAt }

func (t *Transfer) Clone() *Transfer {
	if t == nil {
		return nil
	}
	return &Transfer{id: t.id, from: t.from, to: t.to, amount: t.amount, createdAt: t.createdAt}
}
