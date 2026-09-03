package memory

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

var ErrNestedTransaction = errors.New("memory: nested transaction is not supported")

type state struct {
	accounts    map[domain.AccountID]*domain.Account
	transfers   []*domain.Transfer
	idempotency map[string]application.IdempotencyRecord
	outbox      []application.OutboxMessage
}

type transactionContextKey struct{}

type Repository struct {
	mu    sync.RWMutex
	state state
}

func NewRepository(accounts ...*domain.Account) *Repository {
	repo := &Repository{state: state{
		accounts:    make(map[domain.AccountID]*domain.Account, len(accounts)),
		idempotency: make(map[string]application.IdempotencyRecord),
	}}
	for _, account := range accounts {
		repo.state.accounts[account.ID()] = account.Clone()
	}
	return repo
}

func (r *Repository) WithinTransaction(
	ctx context.Context,
	fn func(context.Context) error,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, exists := stateFromContext(ctx); exists {
		return ErrNestedTransaction
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	staged := cloneState(r.state)
	txCtx := context.WithValue(ctx, transactionContextKey{}, &staged)
	if err := fn(txCtx); err != nil {
		return err
	}
	r.state = staged
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if staged, ok := stateFromContext(ctx); ok {
		return findAccount(*staged, id)
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return findAccount(r.state, id)
}

func (r *Repository) Save(ctx context.Context, account *domain.Account) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if account == nil {
		return domain.ErrInvalidAccountID
	}
	if staged, ok := stateFromContext(ctx); ok {
		staged.accounts[account.ID()] = account.Clone()
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.accounts[account.ID()] = account.Clone()
	return nil
}

func (r *Repository) SaveTransfer(ctx context.Context, transfer *domain.Transfer) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if transfer == nil {
		return domain.ErrInvalidTransferID
	}
	if staged, ok := stateFromContext(ctx); ok {
		staged.transfers = append(staged.transfers, transfer.Clone())
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.transfers = append(r.state.transfers, transfer.Clone())
	return nil
}

func (r *Repository) ListTransfersByAccount(
	ctx context.Context,
	accountID domain.AccountID,
	limit int,
) ([]*domain.Transfer, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if staged, ok := stateFromContext(ctx); ok {
		return listTransfers(*staged, accountID, limit), nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	return listTransfers(r.state, accountID, limit), nil
}

func (r *Repository) ClaimIdempotency(
	ctx context.Context,
	key string,
	requestHash string,
) (domain.TransferID, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, err
	}
	if staged, ok := stateFromContext(ctx); ok {
		record, found := staged.idempotency[key]
		if found {
			if record.RequestHash != requestHash {
				return "", false, application.ErrIdempotencyConflict
			}
			return record.TransferID, record.TransferID != "", nil
		}
		staged.idempotency[key] = application.IdempotencyRecord{Key: key, RequestHash: requestHash}
		return "", false, nil
	}
	return "", false, errors.New("memory: idempotency claim requires transaction")
}

func (r *Repository) CompleteIdempotency(
	ctx context.Context,
	key string,
	transferID domain.TransferID,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if staged, ok := stateFromContext(ctx); ok {
		record := staged.idempotency[key]
		record.TransferID = transferID
		staged.idempotency[key] = record
		return nil
	}
	return errors.New("memory: idempotency completion requires transaction")
}

func (r *Repository) AddOutbox(ctx context.Context, message application.OutboxMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	message = cloneOutboxMessage(message)
	if staged, ok := stateFromContext(ctx); ok {
		staged.outbox = append(staged.outbox, message)
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.state.outbox = append(r.state.outbox, message)
	return nil
}

func (r *Repository) PendingOutbox(
	ctx context.Context,
	limit int,
) ([]application.OutboxMessage, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]application.OutboxMessage, 0, limit)
	for _, message := range r.state.outbox {
		if message.PublishedAt != nil {
			continue
		}
		result = append(result, cloneOutboxMessage(message))
		if len(result) == limit {
			break
		}
	}
	return result, nil
}

func (r *Repository) MarkOutboxPublished(
	ctx context.Context,
	id string,
	publishedAt time.Time,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.state.outbox {
		if r.state.outbox[i].ID == id {
			value := publishedAt.UTC()
			r.state.outbox[i].PublishedAt = &value
			return nil
		}
	}
	return errors.New("memory: outbox message not found")
}

func (r *Repository) Snapshot() map[domain.AccountID]*domain.Account {
	r.mu.RLock()
	defer r.mu.RUnlock()
	snapshot := make(map[domain.AccountID]*domain.Account, len(r.state.accounts))
	for id, account := range r.state.accounts {
		snapshot[id] = account.Clone()
	}
	return snapshot
}

func findAccount(source state, id domain.AccountID) (*domain.Account, error) {
	account, ok := source.accounts[id]
	if !ok {
		return nil, domain.ErrAccountNotFound
	}
	return account.Clone(), nil
}

func listTransfers(source state, accountID domain.AccountID, limit int) []*domain.Transfer {
	result := make([]*domain.Transfer, 0, limit)
	for i := len(source.transfers) - 1; i >= 0 && len(result) < limit; i-- {
		transfer := source.transfers[i]
		if transfer.From() == accountID || transfer.To() == accountID {
			result = append(result, transfer.Clone())
		}
	}
	return result
}

func stateFromContext(ctx context.Context) (*state, bool) {
	value, ok := ctx.Value(transactionContextKey{}).(*state)
	return value, ok
}

func cloneState(source state) state {
	target := state{
		accounts:    make(map[domain.AccountID]*domain.Account, len(source.accounts)),
		transfers:   make([]*domain.Transfer, 0, len(source.transfers)),
		idempotency: make(map[string]application.IdempotencyRecord, len(source.idempotency)),
		outbox:      make([]application.OutboxMessage, 0, len(source.outbox)),
	}
	for id, account := range source.accounts {
		target.accounts[id] = account.Clone()
	}
	for _, transfer := range source.transfers {
		target.transfers = append(target.transfers, transfer.Clone())
	}
	for key, record := range source.idempotency {
		target.idempotency[key] = record
	}
	for _, message := range source.outbox {
		target.outbox = append(target.outbox, cloneOutboxMessage(message))
	}
	return target
}

func cloneOutboxMessage(message application.OutboxMessage) application.OutboxMessage {
	message.Payload = append([]byte(nil), message.Payload...)
	if message.PublishedAt != nil {
		value := *message.PublishedAt
		message.PublishedAt = &value
	}
	return message
}
