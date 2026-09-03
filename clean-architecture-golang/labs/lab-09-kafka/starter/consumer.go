package starter

import (
	"context"
	"encoding/json"
)

type ApplyTransfer interface {
	Apply(context.Context, int64) error
}

type Consumer struct{ apply ApplyTransfer }

func New(apply ApplyTransfer) *Consumer { return &Consumer{apply: apply} }

func (c *Consumer) Handle(ctx context.Context, value []byte) error {
	var event struct {
		Amount int64 `json:"amount"`
	}
	if err := json.Unmarshal(value, &event); err != nil {
		return err
	}
	return c.apply.Apply(ctx, event.Amount)
}
