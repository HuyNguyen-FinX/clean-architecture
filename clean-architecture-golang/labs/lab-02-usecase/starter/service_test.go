package starter

import (
	"errors"
	"testing"
)

func TestTransferService(t *testing.T) {
	accounts := map[string]*Account{
		"A": {ID: "A", Balance: 1_000},
		"B": {ID: "B", Balance: 100},
	}
	service := NewTransferService(accounts)

	if err := service.Transfer("A", "B", 300); err != nil {
		t.Fatalf("transfer: %v", err)
	}
	if accounts["A"].Balance != 700 || accounts["B"].Balance != 400 {
		t.Fatalf("unexpected balances: A=%d B=%d", accounts["A"].Balance, accounts["B"].Balance)
	}
}

func TestTransferServiceRejectsInsufficientBalance(t *testing.T) {
	accounts := map[string]*Account{
		"A": {ID: "A", Balance: 100},
		"B": {ID: "B", Balance: 100},
	}
	err := NewTransferService(accounts).Transfer("A", "B", 101)
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("got %v, want ErrInsufficientBalance", err)
	}
	if accounts["A"].Balance != 100 || accounts["B"].Balance != 100 {
		t.Fatal("rejected transfer changed balances")
	}
}
