package outbox

import (
	"context"
	"fmt"
)

type Event struct {
	ID      string
	Key     []byte
	Payload []byte
}

type Repository interface {
	Claim(context.Context, int) ([]Event, error)
	MarkPublished(context.Context, string) error
}

type Publisher interface {
	Publish(context.Context, Event) error
}

type Worker struct {
	repo      Repository
	publisher Publisher
}

func NewWorker(repo Repository, publisher Publisher) *Worker {
	return &Worker{repo: repo, publisher: publisher}
}

func (w *Worker) RunBatch(ctx context.Context, limit int) error {
	events, err := w.repo.Claim(ctx, limit)
	if err != nil {
		return fmt.Errorf("claim outbox: %w", err)
	}
	for _, event := range events {
		if err := w.publisher.Publish(ctx, event); err != nil {
			return fmt.Errorf("publish event %q: %w", event.ID, err)
		}
		if err := w.repo.MarkPublished(ctx, event.ID); err != nil {
			return fmt.Errorf("mark event %q: %w", event.ID, err)
		}
	}
	return nil
}
