package application

import (
	"context"
	"errors"

	"example.com/cleanarch/lab04/solution/domain"
)

var ErrAccountNotFound = errors.New("account not found")

type AccountRepository interface {
	FindByID(context.Context, string) (*domain.Account, error)
	Save(context.Context, *domain.Account) error
}
