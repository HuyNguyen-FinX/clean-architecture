package starter

import "testing"

func TestTransferSmoke(t *testing.T) {
	accounts := map[string]*Account{"A": {ID: "A", Balance: 100}, "B": {ID: "B"}}
	if err := Transfer(accounts, "A", "B", 50); err != nil {
		t.Fatal(err)
	}
	if accounts["A"].Balance != 50 || accounts["B"].Balance != 50 {
		t.Fatal("unexpected balances")
	}
}
