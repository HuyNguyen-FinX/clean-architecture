package starter

import "testing"

func TestMemoryStoreExposesPointerAlias(t *testing.T) {
	store := NewMemoryStore(&Account{ID: "A", Balance: 100})
	loaded, err := store.FindByID("A")
	if err != nil {
		t.Fatal(err)
	}

	loaded.Balance = 999
	loadedAgain, err := store.FindByID("A")
	if err != nil {
		t.Fatal(err)
	}
	if loadedAgain.Balance != 999 {
		t.Fatal("baseline changed: store no longer exposes pointer alias")
	}
}

func TestDepositService(t *testing.T) {
	store := NewMemoryStore(&Account{ID: "A", Balance: 100})
	if err := NewDepositService(store).Deposit("A", 50); err != nil {
		t.Fatal(err)
	}
	account, _ := store.FindByID("A")
	if account.Balance != 150 {
		t.Fatalf("balance=%d, want 150", account.Balance)
	}
}
