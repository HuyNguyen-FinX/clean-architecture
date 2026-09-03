package composition

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildWiresApplicationAndCleanup(t *testing.T) {
	app, cleanup, err := Build(Config{
		HTTPAddress: ":8080",
		Seed:        map[string]int64{"A": 100},
	})
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/accounts/A", nil)
	response := httptest.NewRecorder()
	app.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Body.String() != "100" {
		t.Fatalf("code=%d body=%q", response.Code, response.Body.String())
	}

	cleanup()
	cleanup()
	if !app.Store.Closed() {
		t.Fatal("cleanup did not close owned store")
	}
}

func TestBuildRejectsInvalidConfig(t *testing.T) {
	app, cleanup, err := Build(Config{})
	if err == nil || app != nil || cleanup != nil {
		t.Fatalf("app=%v cleanup-is-nil=%v err=%v", app, cleanup == nil, err)
	}
}
