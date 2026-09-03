package starter

import "testing"

func TestCacheStaysStaleForever(t *testing.T) {
	primary := map[string]int64{"A": 100}
	reader := New(primary)
	_ = reader.Balance("A")
	primary["A"] = 200
	if got := reader.Balance("A"); got != 100 {
		t.Fatalf("baseline changed: got %d", got)
	}
}
