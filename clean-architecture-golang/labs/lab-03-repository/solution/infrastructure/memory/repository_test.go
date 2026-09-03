package memory

import (
	"testing"

	"example.com/cleanarch/lab03/solution/application"
	"example.com/cleanarch/lab03/solution/domain"
	"example.com/cleanarch/lab03/solution/repositorytest"
)

func TestRepositoryContract(t *testing.T) {
	repositorytest.Contract(t, func(t *testing.T, accounts ...*domain.Account) application.AccountRepository {
		t.Helper()
		repo, err := New(accounts...)
		if err != nil {
			t.Fatal(err)
		}
		return repo
	})
}
