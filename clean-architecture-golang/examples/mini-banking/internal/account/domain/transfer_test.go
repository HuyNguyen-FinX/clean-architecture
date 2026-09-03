package domain

import (
	"errors"
	"testing"
	"time"
)

func TestTransferRequiresIdentityAccountsAmountAndTime(t *testing.T) {
	id, _ := NewTransferID("T-100")
	from, _ := NewAccountID("A-100")
	to, _ := NewAccountID("B-200")
	amount := MustMoney(100, "VND")
	now := time.Unix(1_000, 0)

	tests := []struct {
		name    string
		id      TransferID
		from    AccountID
		to      AccountID
		amount  Money
		at      time.Time
		wantErr error
	}{
		{"valid", id, from, to, amount, now, nil},
		{"missing id", "", from, to, amount, now, ErrInvalidTransferID},
		{"same account", id, from, from, amount, now, ErrSameAccountTransfer},
		{"non-positive", id, from, to, MustMoney(0, "VND"), now, ErrInvalidAmount},
		{"missing time", id, from, to, amount, time.Time{}, ErrInvalidTransferTime},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			transfer, err := NewTransfer(test.id, test.from, test.to, test.amount, test.at)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error=%v, want %v", err, test.wantErr)
			}
			if test.wantErr == nil && (transfer.ID() != id || transfer.CreatedAt().Location() != time.UTC) {
				t.Fatalf("unexpected transfer: %+v", transfer)
			}
		})
	}
}
