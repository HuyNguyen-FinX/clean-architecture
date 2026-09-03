package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"example.com/cleanarch/lab12/solution/application"
	"example.com/cleanarch/lab12/solution/domain"
	"example.com/cleanarch/lab12/solution/httpapi"
	"example.com/cleanarch/lab12/solution/memory"
	"example.com/cleanarch/lab12/solution/support"
)

func TestHTTPTransferReplayAndHistory(t *testing.T) {
	store := memory.New(domain.NewAccount("A", 1_000), domain.NewAccount("B", 100))
	transfer := application.NewTransferMoney(store, &support.IDs{}, support.Clock{Value: time.Unix(1_000, 0)})
	handler := httpapi.New(transfer, store).Routes()
	body := `{"from":"A","to":"B","amount":300}`

	first := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/transfers", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "KEY-1")
	handler.ServeHTTP(first, req)
	if first.Code != http.StatusCreated {
		t.Fatalf("first code=%d body=%s", first.Code, first.Body.String())
	}

	replay := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/transfers", strings.NewReader(body))
	req.Header.Set("Idempotency-Key", "KEY-1")
	handler.ServeHTTP(replay, req)
	if replay.Code != http.StatusOK || !strings.Contains(replay.Body.String(), `"replayed":true`) {
		t.Fatalf("replay code=%d body=%s", replay.Code, replay.Body.String())
	}

	history := httptest.NewRecorder()
	handler.ServeHTTP(history, httptest.NewRequest(http.MethodGet, "/accounts/A/transfers", nil))
	if history.Code != http.StatusOK || !strings.Contains(history.Body.String(), `"Amount":300`) {
		t.Fatalf("history code=%d body=%s", history.Code, history.Body.String())
	}
}
