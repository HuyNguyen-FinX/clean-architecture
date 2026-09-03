package kafkaadapter

import "testing"

func TestDecodeTransferRequestedMapsStableEventIdentity(t *testing.T) {
	command, err := DecodeTransferRequested([]byte(`{
		"event_id":"EVT-1",
		"type":"transfer_requested.v1",
		"from_account_id":"A-100",
		"to_account_id":"B-200",
		"amount":500000,
		"currency":"VND"
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if command.IdempotencyKey != "EVT-1" || command.Amount != 500_000 {
		t.Fatalf("unexpected command: %+v", command)
	}
}

func TestDecodeTransferRequestedRejectsUnsupportedEnvelope(t *testing.T) {
	_, err := DecodeTransferRequested([]byte(`{"event_id":"EVT-1","type":"transfer_requested.v2"}`))
	if err == nil {
		t.Fatal("expected unsupported envelope error")
	}
}
