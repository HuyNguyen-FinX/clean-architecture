package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

const moneyTransferredTopic = "money-transferred-v1"

type TransferMoneyCommand struct {
	IdempotencyKey string
	FromAccountID  string
	ToAccountID    string
	Amount         int64
	Currency       string
}

type TransferMoneyResult struct {
	TransferID domain.TransferID
	Replayed   bool
}

type TransferMoneyUseCase struct {
	store      TransferStore
	transactor Transactor
	ids        IDGenerator
	clock      Clock
}

func NewTransferMoneyUseCase(
	store TransferStore,
	transactor Transactor,
	ids IDGenerator,
	clock Clock,
) *TransferMoneyUseCase {
	if store == nil || transactor == nil || ids == nil || clock == nil {
		panic("application: nil transfer dependency")
	}
	return &TransferMoneyUseCase{store: store, transactor: transactor, ids: ids, clock: clock}
}

func (uc *TransferMoneyUseCase) Execute(
	ctx context.Context,
	cmd TransferMoneyCommand,
) (TransferMoneyResult, error) {
	key := strings.TrimSpace(cmd.IdempotencyKey)
	if key == "" || len(key) > 128 {
		return TransferMoneyResult{}, ErrInvalidCommand
	}
	fromID, err := domain.NewAccountID(cmd.FromAccountID)
	if err != nil {
		return TransferMoneyResult{}, err
	}
	toID, err := domain.NewAccountID(cmd.ToAccountID)
	if err != nil {
		return TransferMoneyResult{}, err
	}
	if fromID == toID {
		return TransferMoneyResult{}, domain.ErrSameAccountTransfer
	}
	amount, err := domain.NewPositiveMoney(cmd.Amount, cmd.Currency)
	if err != nil {
		return TransferMoneyResult{}, err
	}

	hash := transferRequestHash(fromID, toID, amount)
	var result TransferMoneyResult
	err = uc.transactor.WithinTransaction(ctx, func(txCtx context.Context) error {
		previousID, completed, err := uc.store.ClaimIdempotency(txCtx, key, hash)
		if err != nil {
			return fmt.Errorf("claim idempotency key: %w", err)
		}
		if completed {
			result = TransferMoneyResult{TransferID: previousID, Replayed: true}
			return nil
		}

		firstID, secondID := stableAccountOrder(fromID, toID)
		first, err := uc.store.FindByID(txCtx, firstID)
		if err != nil {
			return err
		}
		second, err := uc.store.FindByID(txCtx, secondID)
		if err != nil {
			return err
		}
		sender, receiver := first, second
		if first.ID() != fromID {
			sender, receiver = second, first
		}
		if err := sender.Withdraw(amount); err != nil {
			return err
		}
		if err := receiver.Deposit(amount); err != nil {
			return err
		}

		transferID, err := domain.NewTransferID(uc.ids.NewID())
		if err != nil {
			return err
		}
		now := uc.clock.Now().UTC()
		transfer, err := domain.NewTransfer(transferID, fromID, toID, amount, now)
		if err != nil {
			return err
		}
		eventID := uc.ids.NewID()
		payload, err := json.Marshal(moneyTransferredEvent{
			EventID:    eventID,
			TransferID: string(transfer.ID()),
			From:       string(fromID),
			To:         string(toID),
			Amount:     amount.Amount(),
			Currency:   amount.Currency().String(),
			OccurredAt: now,
		})
		if err != nil {
			return fmt.Errorf("encode outbox event: %w", err)
		}

		if err := uc.store.Save(txCtx, sender); err != nil {
			return err
		}
		if err := uc.store.Save(txCtx, receiver); err != nil {
			return err
		}
		if err := uc.store.SaveTransfer(txCtx, transfer); err != nil {
			return err
		}
		if err := uc.store.AddOutbox(txCtx, OutboxMessage{
			ID: eventID, Topic: moneyTransferredTopic, Key: string(fromID),
			Payload: payload, CreatedAt: now,
		}); err != nil {
			return err
		}
		if err := uc.store.CompleteIdempotency(txCtx, key, transferID); err != nil {
			return err
		}
		result = TransferMoneyResult{TransferID: transferID}
		return nil
	})
	return result, err
}

type moneyTransferredEvent struct {
	EventID    string    `json:"event_id"`
	TransferID string    `json:"transfer_id"`
	From       string    `json:"from_account_id"`
	To         string    `json:"to_account_id"`
	Amount     int64     `json:"amount"`
	Currency   string    `json:"currency"`
	OccurredAt time.Time `json:"occurred_at"`
}

func stableAccountOrder(left, right domain.AccountID) (domain.AccountID, domain.AccountID) {
	if string(left) > string(right) {
		return right, left
	}
	return left, right
}

func transferRequestHash(from, to domain.AccountID, amount domain.Money) string {
	canonical := fmt.Sprintf(
		"%s\x00%s\x00%d\x00%s",
		from,
		to,
		amount.Amount(),
		amount.Currency(),
	)
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:])
}
