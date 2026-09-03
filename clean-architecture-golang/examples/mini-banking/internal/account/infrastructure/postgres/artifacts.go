package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

var (
	_ application.TransferRepository    = (*Repository)(nil)
	_ application.IdempotencyRepository = (*Repository)(nil)
	_ application.OutboxWriter          = (*Repository)(nil)
	_ application.OutboxStore           = (*Repository)(nil)
)

type dbTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (r *Repository) database(ctx context.Context) dbTX {
	if tx, ok := transactionFromContext(ctx); ok {
		return tx
	}
	return r.pool
}

func (r *Repository) SaveTransfer(ctx context.Context, transfer *domain.Transfer) error {
	if transfer == nil {
		return domain.ErrInvalidTransferID
	}
	_, err := r.database(ctx).Exec(ctx,
		`INSERT INTO transfers
			(id, from_account_id, to_account_id, amount_minor, currency, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		transfer.ID(),
		transfer.From(),
		transfer.To(),
		transfer.Amount().Amount(),
		transfer.Amount().Currency().String(),
		transfer.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("save transfer %q: %w", transfer.ID(), err)
	}
	return nil
}

func (r *Repository) ListTransfersByAccount(
	ctx context.Context,
	accountID domain.AccountID,
	limit int,
) ([]*domain.Transfer, error) {
	rows, err := r.database(ctx).Query(ctx,
		`SELECT id, from_account_id, to_account_id, amount_minor, currency, created_at
		   FROM transfers
		  WHERE from_account_id = $1 OR to_account_id = $1
		  ORDER BY created_at DESC, id DESC
		  LIMIT $2`,
		accountID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list transfers for account %q: %w", accountID, err)
	}
	defer rows.Close()

	result := make([]*domain.Transfer, 0)
	for rows.Next() {
		var (
			rawID, rawFrom, rawTo, currency string
			amount                          int64
			createdAt                       time.Time
		)
		if err := rows.Scan(&rawID, &rawFrom, &rawTo, &amount, &currency, &createdAt); err != nil {
			return nil, fmt.Errorf("scan transfer: %w", err)
		}
		id, err := domain.NewTransferID(rawID)
		if err != nil {
			return nil, fmt.Errorf("map transfer id: %w", err)
		}
		from, err := domain.NewAccountID(rawFrom)
		if err != nil {
			return nil, fmt.Errorf("map source account: %w", err)
		}
		to, err := domain.NewAccountID(rawTo)
		if err != nil {
			return nil, fmt.Errorf("map destination account: %w", err)
		}
		money, err := domain.NewPositiveMoney(amount, currency)
		if err != nil {
			return nil, fmt.Errorf("map transfer amount: %w", err)
		}
		transfer, err := domain.NewTransfer(id, from, to, money, createdAt)
		if err != nil {
			return nil, fmt.Errorf("rehydrate transfer: %w", err)
		}
		result = append(result, transfer)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate transfers: %w", err)
	}
	return result, nil
}

func (r *Repository) ClaimIdempotency(
	ctx context.Context,
	key string,
	requestHash string,
) (domain.TransferID, bool, error) {
	if _, ok := transactionFromContext(ctx); !ok {
		return "", false, errors.New("claim idempotency key: transaction required")
	}
	if _, err := r.database(ctx).Exec(ctx,
		`INSERT INTO idempotency_keys (key, request_hash)
		 VALUES ($1, $2)
		 ON CONFLICT (key) DO NOTHING`,
		key,
		requestHash,
	); err != nil {
		return "", false, fmt.Errorf("claim idempotency key: %w", err)
	}
	var (
		hash          string
		rawTransferID *string
	)
	err := r.database(ctx).QueryRow(ctx,
		`SELECT request_hash, transfer_id FROM idempotency_keys WHERE key = $1 FOR UPDATE`,
		key,
	).Scan(&hash, &rawTransferID)
	if err != nil {
		return "", false, fmt.Errorf("read claimed idempotency key: %w", err)
	}
	if hash != requestHash {
		return "", false, application.ErrIdempotencyConflict
	}
	if rawTransferID == nil {
		return "", false, nil
	}
	transferID, err := domain.NewTransferID(*rawTransferID)
	if err != nil {
		return "", false, fmt.Errorf("map idempotency transfer: %w", err)
	}
	return transferID, true, nil
}

func (r *Repository) CompleteIdempotency(
	ctx context.Context,
	key string,
	transferID domain.TransferID,
) error {
	if _, ok := transactionFromContext(ctx); !ok {
		return errors.New("complete idempotency key: transaction required")
	}
	tag, err := r.database(ctx).Exec(ctx,
		`UPDATE idempotency_keys SET transfer_id = $1
		  WHERE key = $2 AND transfer_id IS NULL`,
		transferID,
		key,
	)
	if err != nil {
		return fmt.Errorf("complete idempotency key: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("complete idempotency key: claim missing or already complete")
	}
	return nil
}

func (r *Repository) AddOutbox(ctx context.Context, message application.OutboxMessage) error {
	_, err := r.database(ctx).Exec(ctx,
		`INSERT INTO outbox (id, topic, message_key, payload, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		message.ID,
		message.Topic,
		message.Key,
		message.Payload,
		message.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("add outbox message %q: %w", message.ID, err)
	}
	return nil
}

func (r *Repository) PendingOutbox(
	ctx context.Context,
	limit int,
) ([]application.OutboxMessage, error) {
	rows, err := r.database(ctx).Query(ctx,
		`SELECT id, topic, message_key, payload, created_at
		   FROM outbox
		  WHERE published_at IS NULL
		  ORDER BY created_at, id
		  LIMIT $1`,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("query pending outbox: %w", err)
	}
	defer rows.Close()

	result := make([]application.OutboxMessage, 0)
	for rows.Next() {
		var message application.OutboxMessage
		if err := rows.Scan(
			&message.ID,
			&message.Topic,
			&message.Key,
			&message.Payload,
			&message.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan outbox message: %w", err)
		}
		result = append(result, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate outbox messages: %w", err)
	}
	return result, nil
}

func (r *Repository) MarkOutboxPublished(
	ctx context.Context,
	id string,
	publishedAt time.Time,
) error {
	tag, err := r.database(ctx).Exec(ctx,
		`UPDATE outbox SET published_at = $1 WHERE id = $2 AND published_at IS NULL`,
		publishedAt,
		id,
	)
	if err != nil {
		return fmt.Errorf("mark outbox %q published: %w", id, err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("mark outbox %q published: no pending message", id)
	}
	return nil
}
