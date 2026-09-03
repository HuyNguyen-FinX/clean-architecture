package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type TransferCommand struct {
	IdempotencyKey string
	From           string
	To             string
	Amount         int64
}

type TransferResult struct {
	ID       string
	Replayed bool
}

type TransferMoney struct {
	uow   UnitOfWork
	ids   IDGenerator
	clock Clock
}

func NewTransferMoney(uow UnitOfWork, ids IDGenerator, clock Clock) *TransferMoney {
	if uow == nil || ids == nil || clock == nil {
		panic("application: nil dependency")
	}
	return &TransferMoney{uow: uow, ids: ids, clock: clock}
}

func (uc *TransferMoney) Execute(ctx context.Context, cmd TransferCommand) (TransferResult, error) {
	if cmd.IdempotencyKey == "" || cmd.From == "" || cmd.To == "" ||
		cmd.From == cmd.To || cmd.Amount <= 0 {
		return TransferResult{}, ErrInvalidCommand
	}
	hash := requestHash(cmd)
	var result TransferResult
	err := uc.uow.WithinTransaction(ctx, func(tx Transaction) error {
		if previous, found := tx.FindIdempotency(cmd.IdempotencyKey); found {
			if previous.RequestHash != hash {
				return ErrIdempotencyConflict
			}
			result = TransferResult{ID: previous.TransferID, Replayed: true}
			return nil
		}

		from, err := tx.FindAccount(cmd.From)
		if err != nil {
			return fmt.Errorf("load source: %w", err)
		}
		to, err := tx.FindAccount(cmd.To)
		if err != nil {
			return fmt.Errorf("load destination: %w", err)
		}
		if err := from.Withdraw(cmd.Amount); err != nil {
			return err
		}
		if err := to.Deposit(cmd.Amount); err != nil {
			return err
		}

		id, now := uc.ids.NewID(), uc.clock.Now().UTC()
		record := TransferRecord{ID: id, From: cmd.From, To: cmd.To, Amount: cmd.Amount, CreatedAt: now}
		payload, err := json.Marshal(record)
		if err != nil {
			return err
		}
		if err := tx.SaveAccount(from); err != nil {
			return err
		}
		if err := tx.SaveAccount(to); err != nil {
			return err
		}
		if err := tx.AppendTransfer(record); err != nil {
			return err
		}
		if err := tx.AppendOutbox(OutboxEvent{
			ID: uc.ids.NewID(), Type: "money_transferred.v1", Key: cmd.From,
			Payload: payload, CreatedAt: now,
		}); err != nil {
			return err
		}
		if err := tx.SaveIdempotency(IdempotencyRecord{
			Key: cmd.IdempotencyKey, RequestHash: hash, TransferID: id,
		}); err != nil {
			return err
		}
		result = TransferResult{ID: id}
		return nil
	})
	return result, err
}

func requestHash(cmd TransferCommand) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%d", cmd.From, cmd.To, cmd.Amount)))
	return hex.EncodeToString(sum[:])
}
