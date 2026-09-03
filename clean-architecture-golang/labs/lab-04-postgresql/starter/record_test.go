package starter

import "testing"

func TestStoreAllowsInvalidDomainState(t *testing.T) {
	store := NewStore(AccountRecord{
		ID: "A", Balance: -200, OverdraftLimit: 100, Currency: "VND", Status: "active",
	})
	row, err := store.FindByID("A")
	if err != nil {
		t.Fatal(err)
	}
	if row.Balance >= -row.OverdraftLimit {
		t.Fatal("baseline no longer demonstrates invalid state")
	}
}
