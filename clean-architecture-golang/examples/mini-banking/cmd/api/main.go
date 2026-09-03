package main

import (
	"log"
	"net/http"
	"time"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
	httpadapter "github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/delivery/http"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/infrastructure/memory"
)

func main() {
	repo := memory.NewRepository(
		mustAccount("A-100", 1_000_000, 0, "VND"),
		mustAccount("B-200", 250_000, 0, "VND"),
	)

	transferMoney := application.NewTransferMoneyUseCase(repo, application.NoopTransactor{})
	handler := httpadapter.NewHandler(transferMoney)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("mini banking API listening on http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}

func mustAccount(id string, balanceAmount int64, overdraftAmount int64, currency string) *domain.Account {
	accountID, err := domain.NewAccountID(id)
	if err != nil {
		panic(err)
	}

	balance, err := domain.NewMoney(balanceAmount, currency)
	if err != nil {
		panic(err)
	}

	overdraftLimit, err := domain.NewMoney(overdraftAmount, currency)
	if err != nil {
		panic(err)
	}

	account, err := domain.NewAccount(accountID, balance, overdraftLimit)
	if err != nil {
		panic(err)
	}

	return account
}
