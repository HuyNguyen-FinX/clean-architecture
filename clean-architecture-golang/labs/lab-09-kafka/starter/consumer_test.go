package starter

import (
	"context"
	"testing"
)

type counter struct{ total int64 }

func (c *counter) Apply(_ context.Context, amount int64) error {
	c.total += amount
	return nil
}

func TestDuplicateDeliveryDuplicatesEffect(t *testing.T) {
	useCase := &counter{}
	consumer := New(useCase)
	message := []byte(`{"event_id":"E-1","amount":100}`)
	_ = consumer.Handle(context.Background(), message)
	_ = consumer.Handle(context.Background(), message)
	if useCase.total != 200 {
		t.Fatalf("baseline changed: total=%d", useCase.total)
	}
}
