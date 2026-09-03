package composition

import (
	"errors"
	"net/http"
	"strings"
	"sync"

	"example.com/cleanarch/lab06/solution/application"
	"example.com/cleanarch/lab06/solution/httpapi"
	"example.com/cleanarch/lab06/solution/memory"
)

type Config struct {
	HTTPAddress string
	Seed        map[string]int64
}

type App struct {
	Address string
	Handler http.Handler
	Store   *memory.Store
}

func Build(cfg Config) (*App, func(), error) {
	if strings.TrimSpace(cfg.HTTPAddress) == "" {
		return nil, nil, errors.New("HTTP address is required")
	}

	store := memory.New(cfg.Seed)
	getBalance := application.NewGetBalance(store)
	handler := httpapi.New(getBalance)
	app := &App{Address: cfg.HTTPAddress, Handler: handler.Routes(), Store: store}

	var once sync.Once
	cleanup := func() {
		once.Do(store.Close)
	}
	return app, cleanup, nil
}
