package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"example.com/cleanarch/lab12/solution/application"
	"example.com/cleanarch/lab12/solution/domain"
	"example.com/cleanarch/lab12/solution/httpapi"
	"example.com/cleanarch/lab12/solution/memory"
	"example.com/cleanarch/lab12/solution/support"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	store := memory.New(domain.NewAccount("A", 1_000_000), domain.NewAccount("B", 250_000))
	transfer := application.NewTransferMoney(store, &support.IDs{}, support.Clock{Value: time.Now()})
	handler := httpapi.New(transfer, store)
	server := &http.Server{Addr: ":8080", Handler: handler.Routes(), ReadHeaderTimeout: 5 * time.Second}
	logger.Info("lab 12 API listening", "address", server.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", "error", err)
	}
}
