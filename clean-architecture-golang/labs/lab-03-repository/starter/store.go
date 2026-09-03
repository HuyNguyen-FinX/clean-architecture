package starter

import "errors"

var ErrNotFound = errors.New("not found")

type Account struct {
	ID      string
	Balance int64
}

type MemoryStore struct {
	accounts map[string]*Account
}

func NewMemoryStore(accounts ...*Account) *MemoryStore {
	store := &MemoryStore{accounts: make(map[string]*Account)}
	for _, account := range accounts {
		store.accounts[account.ID] = account
	}
	return store
}

func (s *MemoryStore) FindByID(id string) (*Account, error) {
	account, ok := s.accounts[id]
	if !ok {
		return nil, ErrNotFound
	}
	return account, nil
}

func (s *MemoryStore) Save(account *Account) {
	s.accounts[account.ID] = account
}

type DepositService struct {
	store *MemoryStore
}

func NewDepositService(store *MemoryStore) *DepositService {
	return &DepositService{store: store}
}

func (s *DepositService) Deposit(id string, amount int64) error {
	account, err := s.store.FindByID(id)
	if err != nil {
		return err
	}
	account.Balance += amount
	s.store.Save(account)
	return nil
}
