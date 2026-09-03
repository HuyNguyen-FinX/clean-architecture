package starter

import (
	"fmt"
	"net/http"
)

var globalStore = map[string]int64{"A": 100}

func BalanceHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	balance, ok := globalStore[id]
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "%d", balance)
}
