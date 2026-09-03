package application

import (
	"context"
	"fmt"
)

type RelayOutboxUseCase struct {
	store     OutboxStore
	publisher EventPublisher
	clock     Clock
}

func NewRelayOutboxUseCase(
	store OutboxStore,
	publisher EventPublisher,
	clock Clock,
) *RelayOutboxUseCase {
	if store == nil || publisher == nil || clock == nil {
		panic("application: nil outbox relay dependency")
	}
	return &RelayOutboxUseCase{store: store, publisher: publisher, clock: clock}
}

func (uc *RelayOutboxUseCase) RunOnce(ctx context.Context, limit int) (int, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	messages, err := uc.store.PendingOutbox(ctx, limit)
	if err != nil {
		return 0, fmt.Errorf("load pending outbox: %w", err)
	}
	published := 0
	for _, message := range messages {
		if err := uc.publisher.Publish(ctx, message); err != nil {
			return published, fmt.Errorf("publish outbox %q: %w", message.ID, err)
		}
		if err := uc.store.MarkOutboxPublished(ctx, message.ID, uc.clock.Now().UTC()); err != nil {
			return published, fmt.Errorf("mark outbox %q published: %w", message.ID, err)
		}
		published++
	}
	return published, nil
}
