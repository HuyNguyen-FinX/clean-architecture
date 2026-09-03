package kafkaadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	segmentio "github.com/segmentio/kafka-go"

	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/application"
	"github.com/huynguyen/clean-architecture-golang/examples/mini-banking/internal/account/domain"
)

type TransferExecutor interface {
	Execute(context.Context, application.TransferMoneyCommand) (application.TransferMoneyResult, error)
}

type Consumer struct {
	reader   *segmentio.Reader
	dlq      *segmentio.Writer
	transfer TransferExecutor
}

func NewConsumer(
	brokers []string,
	topic string,
	groupID string,
	dlqTopic string,
	transfer TransferExecutor,
) *Consumer {
	if len(brokers) == 0 || topic == "" || groupID == "" || dlqTopic == "" || transfer == nil {
		panic("kafka adapter: invalid consumer dependency")
	}
	return &Consumer{
		reader: segmentio.NewReader(segmentio.ReaderConfig{
			Brokers:  brokers,
			Topic:    topic,
			GroupID:  groupID,
			MinBytes: 1,
			MaxBytes: 10 << 20,
		}),
		dlq: &segmentio.Writer{
			Addr: segmentio.TCP(brokers...), Topic: dlqTopic,
			RequiredAcks: segmentio.RequireAll, Async: false,
		},
		transfer: transfer,
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		message, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return fmt.Errorf("fetch transfer request: %w", err)
		}

		err = c.handle(ctx, message.Value)
		if err != nil {
			if !isPermanent(err) {
				return err
			}
			if dlqErr := c.writeDLQ(ctx, message, err); dlqErr != nil {
				return errors.Join(err, dlqErr)
			}
		}
		if err := c.reader.CommitMessages(ctx, message); err != nil {
			return fmt.Errorf("commit transfer request offset: %w", err)
		}
	}
}

func (c *Consumer) handle(ctx context.Context, payload []byte) error {
	command, err := DecodeTransferRequested(payload)
	if err != nil {
		return permanentError{err}
	}
	_, err = c.transfer.Execute(ctx, command)
	if isBusinessRejection(err) {
		return permanentError{err}
	}
	return err
}

func DecodeTransferRequested(payload []byte) (application.TransferMoneyCommand, error) {
	var event struct {
		EventID       string `json:"event_id"`
		Type          string `json:"type"`
		FromAccountID string `json:"from_account_id"`
		ToAccountID   string `json:"to_account_id"`
		Amount        int64  `json:"amount"`
		Currency      string `json:"currency"`
	}
	decoder := json.NewDecoder(io.LimitReader(bytes.NewReader(payload), 1<<20))
	if err := decoder.Decode(&event); err != nil {
		return application.TransferMoneyCommand{}, fmt.Errorf("decode transfer request: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return application.TransferMoneyCommand{}, errors.New("decode transfer request: multiple JSON values")
	}
	if strings.TrimSpace(event.EventID) == "" || event.Type != "transfer_requested.v1" {
		return application.TransferMoneyCommand{}, errors.New("unsupported transfer request envelope")
	}
	return application.TransferMoneyCommand{
		IdempotencyKey: event.EventID,
		FromAccountID:  event.FromAccountID,
		ToAccountID:    event.ToAccountID,
		Amount:         event.Amount,
		Currency:       event.Currency,
	}, nil
}

func (c *Consumer) writeDLQ(
	ctx context.Context,
	source segmentio.Message,
	cause error,
) error {
	headers := append([]segmentio.Header(nil), source.Headers...)
	headers = append(headers,
		segmentio.Header{Key: "failure_class", Value: []byte("permanent")},
		segmentio.Header{Key: "failure_reason", Value: []byte(cause.Error())},
	)
	if err := c.dlq.WriteMessages(ctx, segmentio.Message{
		Key: source.Key, Value: source.Value, Headers: headers,
	}); err != nil {
		return fmt.Errorf("publish transfer request DLQ: %w", err)
	}
	return nil
}

func (c *Consumer) Close() error {
	return errors.Join(c.reader.Close(), c.dlq.Close())
}

type permanentError struct{ error }

func isPermanent(err error) bool {
	var target permanentError
	return errors.As(err, &target)
}

func isBusinessRejection(err error) bool {
	return errors.Is(err, domain.ErrInsufficientBalance) ||
		errors.Is(err, domain.ErrAccountFrozen) ||
		errors.Is(err, domain.ErrAccountNotFound) ||
		errors.Is(err, domain.ErrInvalidAccountID) ||
		errors.Is(err, domain.ErrInvalidAmount) ||
		errors.Is(err, domain.ErrInvalidCurrency) ||
		errors.Is(err, domain.ErrCurrencyMismatch) ||
		errors.Is(err, domain.ErrSameAccountTransfer) ||
		errors.Is(err, application.ErrIdempotencyConflict) ||
		errors.Is(err, application.ErrInvalidCommand)
}
