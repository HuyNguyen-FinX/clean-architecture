package starter

import (
	"errors"
	"testing"
)

func TestTransferLeavesPartialWrite(t *testing.T) {
	store := &Store{Accounts: map[string]*Account{
		"A": {ID: "A", Balance: 1_000},
		"B": {ID: "B", Balance: 100},
	}}
	err := store.Transfer("A", "B", 300)
	if !errors.Is(err, ErrReceiverSave) {
		t.Fatalf("got %v", err)
	}
	if store.Accounts["A"].Balance != 700 || store.Accounts["B"].Balance != 100 {
		t.Fatalf("baseline changed: A=%d B=%d",
			store.Accounts["A"].Balance, store.Accounts["B"].Balance)
	}
}
