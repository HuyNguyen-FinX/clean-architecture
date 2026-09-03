package consumer_test

import (
	"context"
	"errors"
	"testing"

	"example.com/cleanarch/lab09/solution/consumer"
	"example.com/cleanarch/lab09/solution/memory"
)

type fakeApply struct {
	calls int
	err   error
}

func (f *fakeApply) Apply(_ context.Context, _ consumer.TransferEvent) error {
	f.calls++
	return f.err
}

var validMessage = []byte(`{"event_id":"E-1","type":"money_transferred","version":1,"amount":100}`)

func TestDuplicateEventIsAppliedOnce(t *testing.T) {
	apply := &fakeApply{}
	handler := consumer.New(memory.NewInbox(), apply)
	if err := handler.Handle(context.Background(), validMessage); err != nil {
		t.Fatal(err)
	}
	if err := handler.Handle(context.Background(), validMessage); err != nil {
		t.Fatal(err)
	}
	if apply.calls != 1 {
		t.Fatalf("calls=%d, want 1", apply.calls)
	}
}

func TestFailureIsNotMarkedAndCanRetry(t *testing.T) {
	apply := &fakeApply{err: errors.New("database down")}
	handler := consumer.New(memory.NewInbox(), apply)
	if err := handler.Handle(context.Background(), validMessage); err == nil {
		t.Fatal("expected first error")
	}
	apply.err = nil
	if err := handler.Handle(context.Background(), validMessage); err != nil {
		t.Fatal(err)
	}
	if apply.calls != 2 {
		t.Fatalf("calls=%d, want 2", apply.calls)
	}
}

func TestMalformedEventIsPermanent(t *testing.T) {
	handler := consumer.New(memory.NewInbox(), &fakeApply{})
	err := handler.Handle(context.Background(), []byte(`{"version":99}`))
	if !consumer.IsPermanent(err) {
		t.Fatalf("got %v, want PermanentError", err)
	}
}
