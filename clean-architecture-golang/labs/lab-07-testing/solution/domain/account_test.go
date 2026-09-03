package domain

import (
	"errors"
	"testing"
)

func TestWithdrawBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		amount  int64
		want    int64
		wantErr error
	}{
		{"happy", 40, 60, nil},
		{"exact", 100, 0, nil},
		{"insufficient", 101, 100, ErrInsufficient},
		{"zero", 0, 100, ErrInvalidAmount},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			account := NewAccount("A", 100)
			err := account.Withdraw(test.amount)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("err=%v want=%v", err, test.wantErr)
			}
			if account.Balance() != test.want {
				t.Fatalf("balance=%d want=%d", account.Balance(), test.want)
			}
		})
	}
}
