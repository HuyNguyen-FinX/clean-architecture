package outbox

import (
	"context"
	"errors"
	"testing"
)

type fakeRepo struct {
	events []Event
	marked []string
}

func (r *fakeRepo) Claim(context.Context, int) ([]Event, error) { return r.events, nil }
func (r *fakeRepo) MarkPublished(_ context.Context, id string) error {
	r.marked = append(r.marked, id)
	return nil
}

type fakePublisher struct {
	err       error
	published []string
}

func (p *fakePublisher) Publish(_ context.Context, event Event) error {
	if p.err != nil {
		return p.err
	}
	p.published = append(p.published, event.ID)
	return nil
}

func TestWorkerMarksOnlyAfterPublish(t *testing.T) {
	repo := &fakeRepo{events: []Event{{ID: "E-1"}}}
	publisher := &fakePublisher{err: errors.New("broker down")}
	err := NewWorker(repo, publisher).RunBatch(context.Background(), 10)
	if err == nil || len(repo.marked) != 0 {
		t.Fatalf("err=%v marked=%v", err, repo.marked)
	}

	publisher.err = nil
	if err := NewWorker(repo, publisher).RunBatch(context.Background(), 10); err != nil {
		t.Fatal(err)
	}
	if len(repo.marked) != 1 || repo.marked[0] != "E-1" {
		t.Fatalf("marked=%v", repo.marked)
	}
}
