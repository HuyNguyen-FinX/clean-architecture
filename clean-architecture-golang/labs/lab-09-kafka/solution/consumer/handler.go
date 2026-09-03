package consumer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type TransferEvent struct {
	EventID string `json:"event_id"`
	Type    string `json:"type"`
	Version int    `json:"version"`
	Amount  int64  `json:"amount"`
}

type ApplyTransfer interface {
	Apply(context.Context, TransferEvent) error
}

type Inbox interface {
	ProcessOnce(context.Context, string, func(context.Context) error) (bool, error)
}

type PermanentError struct{ Cause error }

func (e *PermanentError) Error() string { return e.Cause.Error() }
func (e *PermanentError) Unwrap() error { return e.Cause }

func IsPermanent(err error) bool {
	var permanent *PermanentError
	return errors.As(err, &permanent)
}

type Handler struct {
	inbox Inbox
	apply ApplyTransfer
}

func New(inbox Inbox, apply ApplyTransfer) *Handler {
	if inbox == nil || apply == nil {
		panic("consumer: nil dependency")
	}
	return &Handler{inbox: inbox, apply: apply}
}

func (h *Handler) Handle(ctx context.Context, value []byte) error {
	event, err := decode(value)
	if err != nil {
		return &PermanentError{Cause: err}
	}
	_, err = h.inbox.ProcessOnce(ctx, event.EventID, func(processCtx context.Context) error {
		return h.apply.Apply(processCtx, event)
	})
	return err
}

func decode(value []byte) (TransferEvent, error) {
	decoder := json.NewDecoder(bytes.NewReader(value))
	decoder.DisallowUnknownFields()
	var event TransferEvent
	if err := decoder.Decode(&event); err != nil {
		return TransferEvent{}, fmt.Errorf("decode event: %w", err)
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return TransferEvent{}, errors.New("multiple JSON values")
	}
	if event.EventID == "" || event.Type != "money_transferred" || event.Version != 1 || event.Amount <= 0 {
		return TransferEvent{}, errors.New("invalid or unsupported event")
	}
	return event, nil
}
